package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wa-notif/config"
	httpAdapter "wa-notif/internal/adapter/http"
	"wa-notif/internal/adapter/http/handler"
	"wa-notif/internal/adapter/http/middleware"
	"wa-notif/internal/adapter/whatsapp"
	"wa-notif/internal/usecase"
	"wa-notif/pkg/logger"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// 2. Initialize Logger
	log := logger.InitLogger(cfg.Log.Level)
	log.Info("Starting WhatsApp Notification Gateway",
		"port", cfg.Server.Port,
		"default_recipient", cfg.WhatsApp.DefaultRecipient,
		"db_path", cfg.WhatsApp.SessionDBPath,
	)

	// 3. Initialize WhatsApp Adapter (Whatsmeow)
	waAdapter, err := whatsapp.NewWhatsmeowAdapter(
		cfg.WhatsApp.SessionDBPath,
		cfg.WhatsApp.RateLimitSeconds,
		cfg.Log.Level,
	)
	if err != nil {
		log.Error("Failed to initialize WhatsApp adapter", "error", err)
		os.Exit(1)
	}

	// 4. Start WhatsApp background connection
	if err := waAdapter.Connect(context.Background()); err != nil {
		log.Warn("WhatsApp initial connect warning (will retry automatically)", "error", err)
	}

	// 5. Initialize Use Cases
	sendMsgUC := usecase.NewSendMessageUseCase(waAdapter, cfg.WhatsApp.DefaultRecipient)
	sessionUC := usecase.NewSessionUseCase(waAdapter)

	// 6. Initialize HTTP Handlers & Middlewares
	msgHandler := handler.NewMessageHandler(sendMsgUC)
	sessionHandler := handler.NewSessionHandler(sessionUC)
	openapiHandler := handler.NewOpenAPIHandler()
	authMiddleware := middleware.NewAuthMiddleware(cfg.Auth.BearerToken)

	// 7. Initialize Router
	router := httpAdapter.NewRouter(httpAdapter.RouterConfig{
		MsgHandler:     msgHandler,
		SessionHandler: sessionHandler,
		OpenAPIHandler: openapiHandler,
		AuthMiddleware: authMiddleware,
	})

	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 8. Start HTTP Server in background
	go func() {
		log.Info("HTTP server running", "address", fmt.Sprintf("http://%s", serverAddr))
		log.Info("Open pairing Web UI in browser", "url", fmt.Sprintf("http://localhost:%d/login", cfg.Server.Port))
		log.Info("Interactive API documentation available", "docs_url", fmt.Sprintf("http://localhost:%d/docs", cfg.Server.Port))

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	// 9. Graceful Shutdown on SIGINT / SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down service gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP server forced to shutdown", "error", err)
	}

	waAdapter.Disconnect()
	log.Info("WhatsApp service stopped cleanly. Bye!")
}
