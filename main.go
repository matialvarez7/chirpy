package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/matialvarez7/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	environment    string
	secret         string
	apiKey         string
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading environment: %v", err)
	}

	dbURL := os.Getenv("DB_URL")
	mySecret := os.Getenv("SECRET")
	polkaApikey := os.Getenv("POLKA_KEY")
	env := os.Getenv("PLATFORM")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Database error: %v", err)
	}

	dbQueries := database.New(db)

	mux := http.NewServeMux()

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		environment:    env,
		secret:         mySecret,
		apiKey:         polkaApikey,
	}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.getHits)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHits)
	mux.HandleFunc("POST /api/users", apiCfg.createUserHandler)
	mux.HandleFunc("POST /api/chirps", apiCfg.createChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.listChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirp)
	mux.HandleFunc("POST /api/login", apiCfg.loginHandler)
	mux.HandleFunc("POST /api/refresh", apiCfg.refreshHandler)
	mux.HandleFunc("POST /api/revoke", apiCfg.revokeHandler)
	mux.HandleFunc("PUT /api/users", apiCfg.updateUserHandler)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.deleteChirp)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.polkaWebhookHandler)

	myServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	myServer.ListenAndServe()
}
