package comd

import (
	"bytes"
	"log/slog"
	"net/http"
	"path"
)

// HTTPServer is the HTTP request handler for the Command Daemon.
type HTTPServer struct {
	config *Config
	router *http.ServeMux
	logger *slog.Logger
	execer CommandExecutor
}

// NewHTTPServer creates a new Handler instance.
func NewHTTPServer(config *Config, opts ...HandlerOpts) *HTTPServer {
	s := &HTTPServer{
		config: config,
		router: http.NewServeMux(),
		logger: slog.Default(),
		execer: NewCommandExecutor(slog.Default(), config.Execute),
	}

	basePath := config.BasePath
	if basePath == "" {
		basePath = "/"
	}

	for name, command := range config.Commands {
		s.router.Handle("POST "+path.Join(basePath, name), s.commandHandler(command))
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// HandlerOpts is a function type for setting Handler options.
type HandlerOpts func(*HTTPServer)

// WithLogger sets the logger for the Handler.
func WithLogger(logger *slog.Logger) HandlerOpts {
	return func(h *HTTPServer) {
		h.logger = logger
		h.execer = NewCommandExecutor(logger, h.config.Execute)
	}
}

func (s *HTTPServer) commandHandler(command CommandOpts) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var out bytes.Buffer

		if err := s.execer.ExecuteCommand(r.Context(), command, r.Body, &out); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			WriteJSONError(w, err)
			return
		}

		if out.Len() > 0 {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			w.Write(out.Bytes())
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	})
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
