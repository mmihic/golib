package svr

import (
	"context"
	"crypto/tls"
	"net/http"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/mmihic/golib/src/pkg/httpclient"
	"github.com/mmihic/golib/src/pkg/httpx"
	"github.com/mmihic/golib/src/pkg/netx"
)

type simpleResponse struct {
	Message string `json:"message"`
}

var (
	reHTTP  = regexp.MustCompile(`http://127.0.0.1:\d+`)
	reHTTPs = regexp.MustCompile(`https://127.0.0.1:\d+`)
)

func TestServer(t *testing.T) {

	const (
		secureEndpoint   = "secure"
		insecureEndpoint = "insecure"
	)

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	svr, err := New(log)
	require.NoError(t, err)

	// Create a server with two endpoints
	svr.AddHTTPEndpoint(insecureEndpoint,
		netx.MustParseHostPort("127.0.0.1:0"),
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpx.RespondWithJSON(w, simpleResponse{"Hello there!"})
		}), nil)

	svr.AddHTTPEndpoint(secureEndpoint,
		netx.MustParseHostPort("127.0.0.1:0"),
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpx.RespondWithJSON(w, simpleResponse{"This was over TLS"})
		}), &TLSOptions{
			CertFile: "testdata/server.crt",
			KeyFile:  "testdata/server.key",
		})

	// Spin up the server
	ctx := context.Background()
	err = svr.Start(ctx)
	require.NoError(t, err)

	// Spin up a background monitor for the server
	done := make(chan struct{})
	go func() {
		_ = svr.WaitForShutdown(ctx)
		close(done)
	}()

	// Confirm we have proper endpoints URLs
	var (
		insecureEndpointURL = svr.EndpointURL(insecureEndpoint)
		secureEndpointURL   = svr.EndpointURL(secureEndpoint)
	)
	assert.Regexpf(t, reHTTP, insecureEndpointURL, "insecure endpoint URL is not heep")
	assert.Regexpf(t, reHTTPs, secureEndpointURL, "secure endpoint URL is not heep")

	// Dispatch requests against each of the endpoints.
	for _, tt := range []struct {
		endpoint    string
		wantMessage string
	}{} {
		t.Run(tt.endpoint, func(t *testing.T) {
			endpointURL := svr.EndpointURL(tt.endpoint)
			c, err := httpclient.NewClientWithHTTP(endpointURL, &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true, // needed because we're using localhost
					},
				},
			})
			require.NoError(t, err)

			var resp simpleResponse
			err = c.Get(ctx, "/", httpclient.JSON(&resp))
			require.NoError(t, err)
			assert.Equal(t, simpleResponse{tt.wantMessage}, resp)
		})
	}

	err = svr.Stop(ctx)
	require.NoError(t, err)

	<-done
}
