package http

import (
	"encoding/json"
	"net/http"

	"wa-notif/internal/adapter/http/handler"
	"wa-notif/internal/adapter/http/middleware"
	"wa-notif/internal/adapter/http/web"
)

type RouterConfig struct {
	MsgHandler     *handler.MessageHandler
	SessionHandler *handler.SessionHandler
	OpenAPIHandler *handler.OpenAPIHandler
	AuthMiddleware *middleware.AuthMiddleware
}

func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	// 1. Public Endpoints
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "UP",
		})
	})

	// Web UI Pairing
	mux.HandleFunc("GET /", web.ServePairingUI)
	mux.HandleFunc("GET /login", web.ServePairingUI)

	// API Documentation
	mux.HandleFunc("GET /openapi.json", cfg.OpenAPIHandler.ServeSpec)
	mux.HandleFunc("GET /docs", cfg.OpenAPIHandler.ServeDocs)

	// Status & QR Code (Public agar pairing via browser mudah tanpa perlu ketik bearer token)
	mux.HandleFunc("GET /api/v1/session/status", cfg.SessionHandler.GetStatus)
	mux.HandleFunc("GET /api/v1/session/qr", cfg.SessionHandler.GetQRCode)

	// 2. Protected API Endpoints (Wajib Bearer Token)
	auth := cfg.AuthMiddleware.RequireAuth

	mux.HandleFunc("POST /api/v1/messages/send", auth(cfg.MsgHandler.SendText))
	mux.HandleFunc("POST /api/v1/messages/send-media", auth(cfg.MsgHandler.SendMedia))
	mux.HandleFunc("POST /api/v1/session/logout", auth(cfg.SessionHandler.Logout))

	// CORS & Header wrapper
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		mux.ServeHTTP(w, r)
	})
}
