package httpx

import (
	"context"
	"github.com/pkg/errors"

	"fmt"
	"net/http"
	"reflect"
)

// MustJSONHandler creates a JSON handler from a function, panicking if the
// function cannot be converted into a handler
func MustJSONHandler(method interface{}) http.Handler {
	h, err := JSONHandler(method)
	if err != nil {
		panic(err)
	}

	return h
}

// JSONHandler wraps a function that takes an input argument and returns
// an (output argument, error), decoding the input as a JSON request
// body and returning the output as a JSON response body.
func JSONHandler(method interface{}) (http.Handler, error) {
	v := reflect.ValueOf(method)
	if v.Kind() != reflect.Func {
		return nil, fmt.Errorf("%s is a %s not a func()", v, v.Kind())
	}

	methodType := v.Type()
	reader, err := buildRequestReader(methodType)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	writer, err := buildResponseWriter(methodType)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inputs, err := reader(r)
		if err != nil {
			RespondWithError(w, JSONErrorf(http.StatusBadRequest, "invalid request: %s", err))
			return
		}

		outputs := v.Call(inputs)
		writer(w, outputs)
	}), nil
}

func isStructPtr(typ reflect.Type) bool {
	return typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.Struct
}

func isError(typ reflect.Type) bool {
	return typ == typError
}

func buildRequestReader(methodType reflect.Type) (requestReader, error) {
	if methodType.NumIn() == 0 || methodType.NumIn() > 3 {
		return nil, fmt.Errorf("invalid signature for %s: must take 1-3 inputs", methodType)
	}

	// First argument must always be the context
	if methodType.In(0) != typContext {
		return nil, fmt.Errorf("invalid signature for %s: first input must be a %s", methodType, typContext)
	}

	// If there's only one argument, the function just takes the context.
	if methodType.NumIn() == 1 {
		// Just the context, this is fine
		return func(r *http.Request) ([]reflect.Value, error) {
			return []reflect.Value{
				reflect.ValueOf(r.Context()),
			}, nil
		}, nil
	}

	// If there are two arguments, the second argument can either be the full request
	// or a struct ptr for the JSON body
	if methodType.NumIn() == 2 {
		// Either takes the request, or takes a struct ptr for the body
		switch {
		case methodType.In(1) == typHTTPRequest:
			return func(r *http.Request) ([]reflect.Value, error) {
				return []reflect.Value{
					reflect.ValueOf(r.Context()),
					reflect.ValueOf(r),
				}, nil
			}, nil
		case isStructPtr(methodType.In(1)):
			return func(r *http.Request) ([]reflect.Value, error) {
				body, err := readBody(r, methodType.In(1))
				if err != nil {
					return nil, err
				}

				return []reflect.Value{
					reflect.ValueOf(r.Context()),
					body,
				}, nil
			}, nil

		default:
			return nil, fmt.Errorf("invalid signature %s: input must be a %s or a struct ptr",
				methodType, typHTTPRequest)
		}
	}

	// If there are three arguments, the second must be the HTTP request and the first must be
	// a struct ptr for the body.
	if methodType.In(1) != typHTTPRequest {
		return nil, fmt.Errorf("invalid signature %s: first input must be a %s",
			methodType, typHTTPRequest)
	}

	if !isStructPtr(methodType.In(2)) {
		return nil, fmt.Errorf("invalid signature %s: second input must be a struct ptr", methodType)
	}

	return func(r *http.Request) ([]reflect.Value, error) {
		body, err := readBody(r, methodType.In(2))
		if err != nil {
			return nil, err
		}

		return []reflect.Value{
			reflect.ValueOf(r.Context()),
			reflect.ValueOf(r),
			body,
		}, nil
	}, nil

}

func readBody(req *http.Request, ptrType reflect.Type) (reflect.Value, error) {
	body := reflect.New(ptrType.Elem())
	if err := ReadJSONBody(req, body.Interface()); err != nil {
		return reflect.Value{}, err
	}

	return body, nil
}

func buildResponseWriter(methodType reflect.Type) (responseWriter, error) {
	if methodType.NumOut() == 0 {
		return nil, fmt.Errorf("invalid signature for %s; must return at least return an error", methodType)
	}

	if methodType.NumOut() > 2 {
		return nil, fmt.Errorf("invalid signature for %s; can only return an output structure and an error",
			methodType)
	}

	// Last output must be an error
	if !isError(methodType.Out(methodType.NumOut() - 1)) {
		return nil, fmt.Errorf("invalid signature for %s; last output must be an error", methodType)
	}

	if methodType.NumOut() == 1 {
		// Only returning an error
		return func(w http.ResponseWriter, results []reflect.Value) {
			_ = checkError(w, results)
		}, nil
	}

	// If returning multiple values, first value must be a struct ptr for the response body
	if !isStructPtr(methodType.Out(0)) {
		return nil, fmt.Errorf("invalid signature for %s; can only return struct ptr for response body", methodType)
	}

	return func(w http.ResponseWriter, results []reflect.Value) {
		if checkError(w, results) {
			return
		}

		RespondWithJSON(w, results[0].Interface())
	}, nil
}

func checkError(w http.ResponseWriter, results []reflect.Value) bool {
	errVal := results[len(results)-1]
	if errVal.IsNil() {
		// No error
		return false
	}

	if err, ok := errVal.Interface().(error); ok {
		RespondWithError(w, err)
	} else {
		// TODO(mmihic): Better message here, but this is an invariant that cannot happen
		RespondWithError(w, JSONInternalServerError)
	}

	return true
}

type requestReader func(r *http.Request) ([]reflect.Value, error)
type responseWriter func(w http.ResponseWriter, results []reflect.Value)

var (
	typError       = reflect.TypeOf((*error)(nil)).Elem()
	typHTTPRequest = reflect.TypeOf(&http.Request{})
	typContext     = reflect.TypeOf((*context.Context)(nil)).Elem()
)
