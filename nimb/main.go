package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	app := NewApp()

	mux := http.NewServeMux()

	// Serve static frontend files
	frontendFS, _ := fs.Sub(assets, "frontend")
	fileServer := http.FileServer(http.FS(frontendFS))
	mux.Handle("/", fileServer)

	// API endpoints
	mux.HandleFunc("/api/health", app.handleHealth)
	mux.HandleFunc("/api/config", app.handleConfig)
	mux.HandleFunc("/api/config/save", app.handleSaveConfig)
	mux.HandleFunc("/api/model", app.handleSetModel)
	mux.HandleFunc("/api/apikey", app.handleSetAPIKey)
	mux.HandleFunc("/api/stats", app.handleStats)
	mux.HandleFunc("/api/stats/reset", app.handleResetStats)
	mux.HandleFunc("/api/tunnel/start", app.handleStartTunnel)
	mux.HandleFunc("/api/tunnel/stop", app.handleStopTunnel)
	mux.HandleFunc("/api/tunnel/status", app.handleTunnelStatus)

	// Proxy endpoints (OpenAI compatible)
	mux.HandleFunc("/health", app.handleHealthJSON)
	mux.HandleFunc("/v1/models", app.handleModels)
	mux.HandleFunc("/v1/chat/completions", app.handleChatCompletions)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		app.StopTunnel()
		os.Exit(0)
	}()

	log.Println("===========================================")
	log.Println("  NIMB Mobile - Termux Edition")
	log.Println("===========================================")
	log.Println("  UI:  http://localhost:3000")
	log.Println("  API: http://localhost:3000/v1/chat/completions")
	log.Println("===========================================")

	if err := http.ListenAndServe(":3000", corsMiddleware(mux)); err != nil {
		log.Fatal("Server error:", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
