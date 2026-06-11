package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) getHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	template := `<html>
  					<body>
						<h1>Welcome, Chirpy Admin</h1>
						<p>Chirpy has been visited %d times!</p>
					</body>
				</html>`

	fmt.Fprintf(w, template, cfg.fileserverHits.Load())
}

func (cfg *apiConfig) resetHits(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hits: %d", cfg.fileserverHits.Load())
}

func main() {

	mux := http.NewServeMux()

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.getHits)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHits)
	mux.HandleFunc("POST /api/validate_chirp", validateChirp)

	myServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	myServer.ListenAndServe()
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func validateChirp(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(params.Body) >= 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	body := cleanMessage(params.Body)

	type validChirp struct {
		CleanedBody string `json:"cleaned_body"`
	}

	response := validChirp{
		CleanedBody: body,
	}

	err = respondWithJSON(w, http.StatusOK, response)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

func cleanMessage(message string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	replacement := "****"

	splitedMessage := strings.Split(message, " ")

	for i, word := range splitedMessage {
		for _, badWord := range badWords {
			if strings.ToLower(word) == badWord {
				splitedMessage[i] = replacement
			}
		}
	}

	cleanMsg := strings.Join(splitedMessage, " ")

	return cleanMsg
}
