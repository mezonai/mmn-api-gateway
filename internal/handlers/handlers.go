package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mezonai/mmn-api-gateway/internal/config"
	"github.com/mezonai/mmn-api-gateway/internal/gateway"
	"github.com/sirupsen/logrus"
)

// HTTPServer represents the HTTP server
type HTTPServer struct {
	server  *http.Server
	handler *Handler
	logger  *logrus.Logger
}

// Handler handles HTTP requests
type Handler struct {
	gateway *gateway.BlockchainGateway
	logger  *logrus.Logger
}

// NewHandler creates a new HTTP handler
func NewHandler(gateway *gateway.BlockchainGateway, logger *logrus.Logger) *Handler {
	return &Handler{
		gateway: gateway,
		logger:  logger,
	}
}

// NewHTTPServer creates and starts a new HTTP server
func NewHTTPServer(cfg *config.Config, gateway *gateway.BlockchainGateway, logger *logrus.Logger) (*HTTPServer, error) {
	handler := NewHandler(gateway, logger)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Gateway.HTTPPort),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/stats", handler.Stats)
	mux.HandleFunc("/", handler.Root)

	server.Handler = mux

	httpServer := &HTTPServer{
		server:  server,
		handler: handler,
		logger:  logger,
	}

	logger.Infof("Starting HTTP server on port %d", cfg.Gateway.HTTPPort)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("Failed to serve HTTP: %v", err)
		}
	}()

	return httpServer, nil
}

// Shutdown gracefully shuts down the HTTP server
func (h *HTTPServer) Shutdown(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}

// Health handles health check requests
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "mmn-api-gateway",
	})
}

// Stats handles stats requests
func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	response, err := h.gateway.GetStats(ctx)
	if err != nil {
		h.logger.Errorf("GetStats failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "Failed to get stats",
			"details": err.Error(),
		})
		return
	}

	finalResponse := map[string]interface{}{
		"status": "success",
		"stats":  response,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(finalResponse)
}

// Root handles root endpoint requests
func (h *Handler) Root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service": "MMN API Gateway",
		"version": "1.0.0",
		"endpoints": map[string]string{
			"health": "/health",
			"stats":  "/stats",
		},
	})
}
