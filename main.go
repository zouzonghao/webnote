package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"webnote/pkg/server"
	"webnote/pkg/storage"
	"webnote/pkg/websocket"

	"github.com/gorilla/handlers"
)

func main() {
	// Configuration
	maxStorageSize := int64(10 * 1024 * 1024) // 10MB
	if maxSizeStr := os.Getenv("MAX_STORAGE_SIZE"); maxSizeStr != "" {
		if size, err := strconv.ParseInt(maxSizeStr, 10, 64); err == nil {
			maxStorageSize = size
		}
	}

	defaultPort := "8080"
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	rand.Seed(time.Now().UnixNano())

	// Initialization
	storage.InitStorage(maxStorageSize)
	hub := websocket.NewHub()
	go hub.Run()

	srv := server.NewServer(hub)

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: selectiveCompress(srv),
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Listening on :%s...", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe failed: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

// selectiveCompress avoids compressing WebSocket upgrade requests.
func selectiveCompress(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade") &&
			strings.ToLower(r.Header.Get("Upgrade")) == "websocket" {
			h.ServeHTTP(w, r)
			return
		}
		handlers.CompressHandler(h).ServeHTTP(w, r)
	})
}
