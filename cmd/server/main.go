package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gubernator/pkg/limiter"
	"gubernator/pkg/middleware"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

var (
	httpPort   = flag.String("http-port", "8080", "HTTP server port")
	grpcPort   = flag.String("grpc-port", "9090", "gRPC server port")
	redisAddr  = flag.String("redis-addr", "localhost:6379", "Redis server address")
	redisDB    = flag.Int("redis-db", 0, "Redis database number")
	capacity   = flag.Int("capacity", 10, "Token bucket capacity")
	refillRate = flag.Float64("refill-rate", 1.0, "Tokens per second refill rate")
)

func main() {
	flag.Parse()

	// Initialize rate limiter
	limiterConfig := limiter.Config{
		Capacity:   *capacity,
		RefillRate: *refillRate,
		RedisAddr:  *redisAddr,
		RedisDB:    *redisDB,
	}

	rl, err := limiter.NewRateLimiter(limiterConfig)
	if err != nil {
		log.Fatalf("Failed to create rate limiter: %v", err)
	}
	defer rl.Close()

	// Create HTTP router
	router := mux.NewRouter()

	// Health check endpoint (no rate limiting) - register before middleware
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		jsonResponse(w, map[string]string{"status": "healthy"})
	}).Methods("GET")

	// API endpoints with rate limiting
	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.RateLimitMiddleware(rl, middleware.DefaultKeyExtractor))
	api.HandleFunc("/test", handleTest).Methods("GET", "POST")
	api.HandleFunc("/data", handleData).Methods("GET")

	// Start HTTP server
	httpServer := &http.Server{
		Addr:    ":" + *httpPort,
		Handler: router,
	}

	// Start gRPC server
	grpcServer := grpc.NewServer()
	// TODO: Register gRPC services here

	grpcListener, err := net.Listen("tcp", ":"+*grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port: %v", err)
	}

	// Start servers in goroutines
	go func() {
		log.Printf("HTTP server starting on port %s", *httpPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	go func() {
		log.Printf("gRPC server starting on port %s", *grpcPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down servers...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	grpcServer.GracefulStop()
	log.Println("Servers stopped")
}

func handleTest(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, map[string]interface{}{
		"message":   "Request successful",
		"timestamp": time.Now().Unix(),
		"method":    r.Method,
	})
}

func handleData(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, map[string]interface{}{
		"data": []map[string]interface{}{
			{"id": 1, "name": "Item 1"},
			{"id": 2, "name": "Item 2"},
			{"id": 3, "name": "Item 3"},
		},
		"count": 3,
	})
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
