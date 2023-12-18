package oauth2x

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"text/template"
	"time"

	"github.com/pkg/browser"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"github.com/mmihic/golib/src/pkg/netx"
)

// BrowserTokenSourceOptions are options to
// creating a BrowserTokenSource
type BrowserTokenSourceOptions struct {
	Log             *zap.Logger
	SuccessTemplate string
	ErrorTemplate   string
	Timeout         time.Duration
}

func (opts *BrowserTokenSourceOptions) withDefaults() *BrowserTokenSourceOptions {
	var copts BrowserTokenSourceOptions
	if opts != nil {
		copts = *opts
	}

	if copts.ErrorTemplate == "" {
		copts.ErrorTemplate = defaultErrorTemplate
	}

	if copts.SuccessTemplate == "" {
		copts.SuccessTemplate = defaultSuccesTemplate
	}

	return &copts
}

// A BrowserTokenSource retrieves tokens from an OAuth2 provider
// that presents a consent page in the local browser. Useful for
// acquiring credentials
func BrowserTokenSource(
	conf *oauth2.Config,
	callbackHostPort netx.HostPort,
	opts *BrowserTokenSourceOptions,
) (oauth2.TokenSource, error) {
	opts = opts.withDefaults()
	errorTemplate, err := template.New("error").Parse(opts.ErrorTemplate)
	if err != nil {
		return nil, fmt.Errorf("unable to parse error template: %w", err)
	}

	successTemplate, err := template.New("success").Parse(opts.SuccessTemplate)
	if err != nil {
		return nil, fmt.Errorf("unable to parse success template: %w", err)
	}

	return &browserTokenSource{
		conf:             conf,
		callbackHostPort: callbackHostPort,
		timeout:          opts.Timeout,
		successTemplate:  successTemplate,
		errorTemplate:    errorTemplate,
	}, nil
}

type browserTokenSource struct {
	conf             *oauth2.Config
	callbackHostPort netx.HostPort
	timeout          time.Duration
	successTemplate  *template.Template
	errorTemplate    *template.Template
}

func (src *browserTokenSource) Token() (*oauth2.Token, error) {
	ctx := context.Background()
	if src.timeout > 0 {
		cancelCtx, fn := context.WithTimeout(ctx, src.timeout)
		defer fn()
		ctx = cancelCtx
	}

	// Setup channels to receive tokens or errors
	tokenCh := make(chan *oauth2.Token)
	defer close(tokenCh)

	errorCh := make(chan error)
	defer close(errorCh)

	// Start a server to receive the oauth callback
	closeFn, err := src.startCallbackServer(tokenCh, errorCh)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	// Get the URL for authorization and open it in the local browser.
	url := src.conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	if err := browser.OpenURL(url); err != nil {
		return nil, fmt.Errorf("could not open browser: %w", err)
	}

	// Wait to get a token, an error, or for the context to be canceled or timeout.
	var token *oauth2.Token
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-errorCh:
	case token = <-tokenCh:
	}

	return token, err
}

func (src *browserTokenSource) startCallbackServer(
	tokenCh chan<- *oauth2.Token, errorCh chan<- error,
) (func(), error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			// TODO(mmihic): Uh, don't do this
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Spin up a server to retrieve the OAuth callbacks
	svr := &http.Server{
		Addr: src.callbackHostPort.String(),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), oauth2.HTTPClient, httpClient)
			code := req.URL.Query().Get("code")

			// Exchange will do the handshake to retrieve the initial access token.
			tok, err := src.conf.Exchange(ctx, code)
			if err != nil {
				errorCh <- err
				_ = src.errorTemplate.Execute(w, map[string]any{
					"Error": err.Error(),
				})
				return
			}

			tokenCh <- tok
			_ = src.successTemplate.Execute(w, map[string]any{})
		}),
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		if err := svr.ListenAndServe(); err != http.ErrServerClosed {
			errorCh <- fmt.Errorf("unable to start callback server: %w", err)
		}
		wg.Done()
	}()

	return func() {
		_ = svr.Close()
		wg.Wait()
	}, nil
}

const (
	defaultSuccesTemplate = `
<html>
	<head><title>Authorization complete</title></head>
	<body><p>Authorization complete. You can close this browser tab.</p></body>
</html>
`

	defaultErrorTemplate = `
<html>
	<head><title>Error exchanging auth code for token</title></head>
	<body><p><b>Error exchanging auth code for token:</b>{{.Error}}</p></body>
</html>
`
)
