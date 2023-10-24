// Package httpclient contains helpers for making calls to HTTP servers.
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
)

const (
	ContentTypeJSON = "application/json"
)

// Error is an HTTP error.
type Error struct {
	StatusCode int
	Body       bytes.Buffer
}

// Error returns the error message.
func (err *Error) Error() string {
	return fmt.Sprintf("http error %d", err.StatusCode)
}

// NewError returns a new HTTP client error.
func NewError(statusCode int, r io.Reader) error {
	err := &Error{StatusCode: statusCode}
	_, _ = io.Copy(&err.Body, r)
	return err
}

// A Marshaller is data that is sent as part of a request.
type Marshaller interface {
	Marshal(w io.Writer) error
	ContentType() string
}

// An Unmarshaller is data that is read from a response.
type Unmarshaller interface {
	Unmarshal(r io.Reader) error
}

// A Content can send and receive content of a given type.
type Content interface {
	Marshaller
	Unmarshaller
}

// Empty returns an Content that doesn't send or receive data.
func Empty() Content {
	return empty{}
}

type empty struct{}

func (_ empty) Marshal(_ io.Writer) error   { return nil }
func (_ empty) Unmarshal(_ io.Reader) error { return nil }
func (_ empty) ContentType() string         { return "" }

// JSON returns a Content that handles JSON.
func JSON(v any) Content {
	return jsonContent{v: v, pretty: false}
}

// JSONPretty returns a Content that submits JSON with pretty-printing
func JSONPretty(v any) Content {
	return jsonContent{v: v, pretty: true}
}

type jsonContent struct {
	v      any
	pretty bool
}

func (c jsonContent) Marshal(w io.Writer) error {
	enc := json.NewEncoder(w)
	if c.pretty {
		enc.SetIndent("", "  ")
	}

	return enc.Encode(c.v)
}

func (c jsonContent) Unmarshal(r io.Reader) error {
	return json.NewDecoder(r).Decode(c.v)
}

func (c jsonContent) ContentType() string {
	return ContentTypeJSON
}

// A Client is a client that can send and retrieve JSON documents.
type Client interface {
	SetOptions(opts ...CallOption) error
	Get(ctx context.Context, path string, out Unmarshaller, opts ...CallOption) error
	Put(ctx context.Context, path string, in Marshaller, out Unmarshaller, opts ...CallOption) error
	Post(ctx context.Context, path string, in Marshaller, out Unmarshaller, opts ...CallOption) error
	Delete(ctx context.Context, path string, in Marshaller, out Unmarshaller, opts ...CallOption) error
	SetTraceLogger(logger *zap.Logger)
}

// NewClient creates a new HTTP client pointed at the given URL.
func NewClient(baseURL string, opts ...CallOption) (Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	return &client{
		url:  u,
		opts: opts,
		http: &http.Client{}, // TODO(mmihic): Options for the HTTP client
	}, nil
}

type client struct {
	url         *url.URL
	http        *http.Client
	opts        []CallOption
	traceLogger *zap.Logger
}

func (c *client) SetTraceLogger(traceLogger *zap.Logger) {
	c.traceLogger = traceLogger
}

func (c *client) SetOptions(opts ...CallOption) error {
	c.opts = append(c.opts, opts...)
	return nil
}

func (c *client) Get(ctx context.Context, path string, out Unmarshaller, opts ...CallOption) error {
	return c.do(ctx, "GET", path, nil, out, opts...)
}

func (c *client) Put(ctx context.Context, path string, in Marshaller, out Unmarshaller, opts ...CallOption) error {
	return c.do(ctx, "PUT", path, in, out, opts...)
}

func (c *client) Post(ctx context.Context, path string, in Marshaller, out Unmarshaller, opts ...CallOption) error {
	return c.do(ctx, "POST", path, in, out, opts...)
}

func (c *client) Delete(ctx context.Context, path string, in Marshaller, out Unmarshaller, opts ...CallOption) error {
	return c.do(ctx, "DELETE", path, in, out, opts...)
}

func (c *client) do(ctx context.Context, method, path string, in Marshaller, out Unmarshaller, opts ...CallOption) error {
	u := *c.url

	if len(path) != 0 {
		if len(u.Path) != 0 {
			// Append path to built-in path
			hasTrailingPathSep := u.Path[len(u.Path)-1] == '/'
			hasLeadingPathSep := path[0] == '/'
			switch {
			case hasTrailingPathSep && hasLeadingPathSep:
				// Both have a separator, strip one of them
				u.Path += path[1:]
			case hasTrailingPathSep && !hasLeadingPathSep:
				// Base path has a separator but request path does not, can just join
				u.Path += path
			case !hasTrailingPathSep && hasLeadingPathSep:
				// Base path lacks a separator but request path has one, can just join
				u.Path += path
			case !hasTrailingPathSep && !hasLeadingPathSep:
				// Neither the base path nor the request path have a separator, need to add one
				u.Path += "/" + path
			}
		} else if path[0] != '/' {
			// Needs a path separator prefix
			u.Path = "/" + path
		} else {
			// Path starts with a separator prefix
			u.Path = path
		}
	}

	var rqstbuf bytes.Buffer
	if in != nil {
		if err := in.Marshal(&rqstbuf); err != nil {
			return err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), &rqstbuf)
	if err != nil {
		return err
	}

	if in != nil {
		req.Header.Set("Content-Type", in.ContentType())
	}

	for _, opt := range c.opts {
		if err := opt(req); err != nil {
			return err
		}
	}

	for _, opt := range opts {
		if err := opt(req); err != nil {
			return err
		}
	}

	if c.traceLogger != nil {
		c.traceLogger.Info("sending request",
			zap.String("url", req.URL.String()),
			zap.String("body", rqstbuf.String()))
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}

	body := io.Reader(resp.Body)
	defer func() { _ = resp.Body.Close() }()

	if c.traceLogger != nil {
		// Parse into buffer and dump
		var respbuf bytes.Buffer
		_, _ = io.Copy(&respbuf, body)
		c.traceLogger.Info("received response",
			zap.String("url", req.URL.String()),
			zap.Int("response_code", resp.StatusCode),
			zap.String("response_body", respbuf.String()))

		body = &respbuf
	}

	if resp.StatusCode > 299 {
		return NewError(resp.StatusCode, body)
	}

	if out != nil {
		if err := out.Unmarshal(body); err != nil {
			return nil
		}
	}

	return nil
}

var (
	_ Content    = empty{}
	_ Marshaller = jsonContent{}
	_ Client     = &client{}
	_ error      = &Error{}
)
