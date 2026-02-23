package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/jamesphm04/splose-clone-be/internal/config"
	"github.com/jamesphm04/splose-clone-be/internal/container"
	"github.com/jamesphm04/splose-clone-be/internal/logger"
)

func main() {
	// Init
	// Logger
	log := logger.Must(os.Getenv("APP_ENV"))
	defer log.Sync()

	// Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load configuration", zap.Error(err))
	}
	log.Info("configuration loaded", zap.String("env", cfg.AppEnv))

	// Dependency Injection Container
	ctr, err := container.New(cfg, log)
	if err != nil {
		log.Fatal("failed to create container", zap.Error(err))
	}
	log.Info("container created")
	defer ctr.Close()

	// HTTP Server
	router := ctr.Router()
	ginEngine, ok := router.(interface {
		ServeHTTP(http.ResponseWriter, *http.Request)
	})
	if !ok {
		log.Fatal("router does not implement http.Handler")
	}

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      ginEngine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start the server in a goroutine; main goroutine blocks on OS signal.
	go func() {
		log.Info("HTTP server starting", zap.String("addr", addr), zap.String("env", cfg.AppEnv))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("HTTP server error", zap.Error(err))
		}
	}()
	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Info("shutdown signal received", zap.String("signal", sig.String()))

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("graceful shutdown failed – forcing close", zap.Error(err))
	} else {
		log.Info("HTTP server stopped gracefully")
	}
}
