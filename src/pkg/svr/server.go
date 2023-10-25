// Package svr contains generic server startup/shutdown code, currently for running HTTP
// servers, but can also add in GRPC etc down the line.
package svr

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"golang.org/x/sync/errgroup"

	"go.uber.org/zap"

	"github.com/mmihic/golib/src/pkg/netx"
)

// A Server is a server mainline.
type Server interface {
	AddHTTPEndpoint(name string, _ netx.HostPort, _ http.Handler, _ *TLSOptions) Server
	EndpointURL(name string) string
	Start(_ context.Context) error
	Stop(_ context.Context) error
	Run(_ context.Context) error
	WaitForShutdown(_ context.Context) error
}

type server struct {
	log         *zap.Logger
	httpServers map[string]*httpServer
	listenGroup errgroup.Group
}

type httpServer struct {
	*http.Server
	TLSOptions *TLSOptions
	listenURL  string
}

// Options are options to a server.
type Options struct {
	Logger        *zap.Logger
	RequestLogger *zap.Logger
}

// New creates a new server.
func New(log *zap.Logger) (Server, error) {
	return &server{
		log:         log,
		httpServers: map[string]*httpServer{},
	}, nil
}

// TLSOptions are options for transport layer security.
type TLSOptions struct {
	Config   *tls.Config
	CertFile string
	KeyFile  string
}

// Copy creates a copy of the TLSOptions.
func (opts *TLSOptions) Copy() *TLSOptions {
	if opts == nil {
		return nil
	}

	cp := *opts
	return &cp
}

// AddHTTPEndpoint adds an HTTP endpoint to the server. The server will listen
// on the given host:port and route incoming HTTP requests to the provided handler.
func (svr *server) AddHTTPEndpoint(
	name string, hostPort netx.HostPort, handler http.Handler, tlsOpts *TLSOptions,
) Server {
	if _, exists := svr.httpServers[name]; exists {
		panic(fmt.Sprintf("server '%s' already registered", name))
	}

	var tlsCfg *tls.Config
	if tlsOpts != nil {
		tlsCfg = tlsOpts.Config
	}
	s := &httpServer{
		Server: &http.Server{
			Addr:      hostPort.String(),
			Handler:   handler,
			ErrorLog:  zap.NewStdLog(svr.log),
			TLSConfig: tlsCfg,
		},
		TLSOptions: tlsOpts.Copy(),
	}

	if tlsOpts != nil {
		s.Server.TLSConfig = s.TLSConfig
	}

	svr.log.Info("registered HTTP server",
		zap.String("name", name),
		zap.String("addr", hostPort.String()))
	svr.httpServers[name] = s
	return svr
}

// EndpointURL returns the address of the named HTTP endpoint.
func (svr *server) EndpointURL(name string) string {
	if s, ok := svr.httpServers[name]; ok {
		return s.listenURL
	}

	return ""
}

// Start starts the server, and begins listening on all registered endpoints. Does not block.
func (svr *server) Start(_ context.Context) error {
	svr.log.Info("starting server")

	listeners := make(map[string]net.Listener, len(svr.httpServers))

	// Listen here, but serve asynchronously
	for name, s := range svr.httpServers {
		svr.log.Info("starting HTTP endpoint",
			zap.String("name", name),
			zap.String("addr", s.Addr))

		ln, err := net.Listen("tcp", s.Addr)
		if err != nil {
			svr.closeListeners(listeners)
			return fmt.Errorf("could not listen on %s: %w", s.Addr, err)
		}

		listeners[name] = ln

		protocol := "http"
		if s.TLSOptions != nil {
			protocol = "https"
		}

		s.listenURL = fmt.Sprintf("%s://%s", protocol, ln.Addr().String())
		svr.log.Info("listening to endpoint",
			zap.String("name", name),
			zap.String("url", s.listenURL))
	}

	// Serve in background
	for name, s := range svr.httpServers {
		name, s, ln := name, s, listeners[name]
		svr.listenGroup.Go(func() error {
			var err error
			if s.TLSOptions != nil {
				svr.log.Info("serving HTTPS endpoint",
					zap.String("name", name),
					zap.String("addr", ln.Addr().String()),
					zap.String("certFile", s.TLSOptions.CertFile),
					zap.String("keyFile", s.TLSOptions.KeyFile))
				err = s.ServeTLS(ln, s.TLSOptions.CertFile, s.TLSOptions.KeyFile)
			} else {
				svr.log.Info("serving HTTP endpoint",
					zap.String("name", name),
					zap.String("addr", ln.Addr().String()))
				err = s.Serve(ln)
			}

			if err != nil {
				return fmt.Errorf("unable to start server %s (%s): %s", name, s.Addr, err)
			}
			return nil
		})
	}

	return nil
}

func (svr *server) closeListeners(listeners map[string]net.Listener) {
	for name, ln := range listeners {
		svr.log.Info("closing listener",
			zap.String("name", name),
			zap.String("addr", ln.Addr().String()))
		_ = ln.Close()
	}
}

// Stop stops the server, closing down all listeners. Does not block.
func (svr *server) Stop(ctx context.Context) error {
	var shutdownGroup errgroup.Group

	for name, s := range svr.httpServers {
		name, s := name, s
		shutdownGroup.Go(func() error {
			if err := s.Shutdown(ctx); err != nil {
				return fmt.Errorf("unable to close server %s (%s): %w", name, s.Addr, err)
			}

			svr.log.Info("shutdown server",
				zap.String("name", name),
				zap.String("url", s.listenURL))

			return nil
		})
	}

	err := shutdownGroup.Wait()
	svr.log.Info("all servers shutdown")
	return err
}

// WaitForShutdown waits for the server to completely shutdown.
func (svr *server) WaitForShutdown(ctx context.Context) error {
	return svr.listenGroup.Wait()
}

// Run starts the server, then blocks waiting for the server to shut down.
func (svr *server) Run(ctx context.Context) error {
	if err := svr.Start(ctx); err != nil {
		return err
	}

	return svr.WaitForShutdown(ctx)
}
