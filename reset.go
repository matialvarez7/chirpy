package main

import (
	"errors"
	"fmt"
	"net/http"
)

func (cfg *apiConfig) resetHits(w http.ResponseWriter, r *http.Request) {

	if "dev" != cfg.environment {
		err := errors.New("You have to be in development platform")
		respondWithError(w, http.StatusForbidden, err.Error())
		return
	}

	cfg.fileserverHits.Store(0)
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err := cfg.db.ResetUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	fmt.Fprintf(w, "Hits: %d", cfg.fileserverHits.Load())
}
