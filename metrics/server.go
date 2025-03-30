package metrics

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server provides HTTP endpoint for metrics scraping
type Server struct {
	server *http.Server
	logger *log.Logger
}

// NewServer creates a new metrics server
func NewServer(address string, logger *log.Logger) *Server {
	if logger == nil {
		logger = log.New(log.Writer(), "metrics: ", log.LstdFlags)
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    address,
		Handler: mux,
	}

	return &Server{
		server: server,
		logger: logger,
	}
}

// Start begins serving metrics
func (s *Server) Start() error {
	s.logger.Printf("Starting metrics server on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully stops the metrics server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Println("Shutting down metrics server")

	// Create a timeout context for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.server.Shutdown(shutdownCtx)
}
