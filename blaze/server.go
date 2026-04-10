package blaze

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
)

// Server is a Blaze mock server.
type Server struct {
	config   serverConfig
	registry *StubRegistry
	listener net.Listener
	server   *http.Server
	logger   *slog.Logger
	logFile  *os.File
}

// NewServer creates a new mock server with the given options.
func NewServer(opts ...ServerOption) *Server {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	logger, logFile, err := newLogger(cfg.logOutput, cfg.logFile)
	if err != nil {
		panic(err)
	}

	return &Server{
		config:   cfg,
		registry: &StubRegistry{},
		logger:   logger,
		logFile:  logFile,
	}
}

// Stub registers a new stub and returns its ID.
func (s *Server) Stub(b *StubBuilder) string {
	stub := b.build()
	return s.registry.Add(stub)
}

// RemoveStub removes a stub by ID. Returns true if found and removed.
func (s *Server) RemoveStub(id string) bool {
	return s.registry.Remove(id)
}

// ResetStubs removes all registered stubs.
func (s *Server) ResetStubs() {
	s.registry.Reset()
}

// ListStubs returns all registered stubs.
func (s *Server) ListStubs() []Stub {
	return s.registry.List()
}

// URL returns the base URL of the running server (e.g. "http://127.0.0.1:8080").
// Only valid after Start has been called.
func (s *Server) URL() string {
	if s.listener == nil {
		return ""
	}
	return "http://" + s.listener.Addr().String()
}

// Start begins listening and serving HTTP requests. It blocks until the server is shut down.
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		s.logger.Error("failed to start server", "addr", addr, "error", err)
		return fmt.Errorf("blaze: failed to listen on %s: %w", addr, err)
	}
	s.listener = ln

	s.server = &http.Server{
		Handler: &mockHandler{registry: s.registry, logger: s.logger},
	}

	stubs := s.registry.List()
	stubAttrs := make([]slog.Attr, len(stubs))
	for i, st := range stubs {
		stubAttrs[i] = slog.String(st.ID, fmt.Sprintf("%s %s", st.Request.Method, st.Request.Path))
	}
	s.logger.Info("server started",
		"url", s.URL(),
		slog.Group("stubs", attrsToAny(stubAttrs)...),
	)

	return s.server.Serve(ln)
}

func attrsToAny(attrs []slog.Attr) []any {
	out := make([]any, len(attrs))
	for i, a := range attrs {
		out[i] = a
	}
	return out
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() error {
	if s.logFile != nil {
		s.logFile.Close()
	}
	if s.server == nil {
		return nil
	}
	return s.server.Close()
}
