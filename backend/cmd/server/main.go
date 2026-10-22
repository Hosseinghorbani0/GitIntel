package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gitintel/backend/internal/analytics"
	"gitintel/backend/internal/api"
	"gitintel/backend/internal/cache"
	gh "gitintel/backend/internal/github"
	"gitintel/backend/internal/llm"
)

func main() {
	loadLocalEnv()
	cacheStore := cache.NewCache(api.AnalysisCacheTTL)
	engine := analytics.NewEngine()
	llmConfig, _ := llm.LoadConfigFromEnv()
	credentials := llm.NewCredentialManager(llm.CredentialEnvName(llmConfig.Provider))
	credentials.Load()
	handler := api.NewHandlerWithCredentials(cacheStore, engine, credentials)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", handler.Health)
	mux.HandleFunc("/api/analyze", handler.Analyze)
	mux.HandleFunc("/api/interpret", handler.Interpret)
	mux.HandleFunc("/api/interpret/stream", handler.InterpretStream)
	mux.HandleFunc("/api/report/", handler.Report)
	mux.HandleFunc("/api/resume", handler.Resume)
	mux.HandleFunc("/api/portfolio", handler.Portfolio)
	mux.HandleFunc("/api/profile-readme", handler.ProfileReadme)
	mux.HandleFunc("/api/github/profile/", handler.Profile)
	mux.HandleFunc("/api/github/repos/", handler.Repositories)
	mux.HandleFunc("/api/llm/status", llm.HandleStatus)
	mux.HandleFunc("/api/llm/test", func(w http.ResponseWriter, r *http.Request) {
		llm.HandleTestWithManager(w, r, credentials)
	})

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           withCORS(mux),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("GitIntel backend listening on http://localhost:%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func loadLocalEnv() {
	for _, path := range []string{".env", "../.env"} {
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Load(path); err != nil {
				log.Printf("could not load local environment file %s", path)
			}
			return
		}
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		w.Header().Add("Vary", "Origin")
		if origin == "http://localhost:5173" || origin == "http://127.0.0.1:5173" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func init() {
	_ = gh.NewClientFromToken("")
}
