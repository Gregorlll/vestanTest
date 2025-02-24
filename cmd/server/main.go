package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"vestantest/internal/database"
	"vestantest/internal/server"
	"vestantest/internal/server/config"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize logger
	logger := log.New(os.Stdout, "[CHAT] ", log.LstdFlags)

	// Initialize database
	db, err := database.NewDB(cfg.DBConnection)
	if err != nil {
		logger.Fatal("Database connection error:", err)
	}

	// Create new server
	srv := server.NewServer(db, cfg)

	// Start message processing in goroutine
	go srv.Run()

	// Configure HTTP server
	httpServer := &http.Server{
		Addr:         ":" + cfg.Port,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
	}

	// Setup routes
	http.HandleFunc("/ws", srv.HandleWebSocket)
	http.HandleFunc("/messages", srv.HandleMessages)
	http.HandleFunc("/connection-history", srv.HandleConnectionHistory)

	// Start HTTP server in goroutine
	go func() {
		logger.Printf("Server is running on :%s\n", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server start error:", err)
		}
	}()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Println("Received shutdown signal, starting graceful shutdown...")

	// Сначала завершаем работу сервера чата
	srv.Shutdown()

	// Затем завершаем работу HTTP сервера
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Printf("Server shutdown error: %v\n", err)
	}

	logger.Println("Server stopped successfully")
}
