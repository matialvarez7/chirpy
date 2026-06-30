package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/matialvarez7/chirpy/internal/auth"
)

type dataWebhook struct {
	UserID uuid.UUID `json:"user_id"`
}

type requestInfo struct {
	Event string      `json:"event"`
	Data  dataWebhook `json:"data"`
}

func (cfg *apiConfig) polkaWebhookHandler(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if apiKey != cfg.apiKey {
		respondWithError(w, http.StatusUnauthorized, errors.New("You need an API key fot this").Error())
		return
	}
	decoder := json.NewDecoder(r.Body)

	info := requestInfo{}

	err = decoder.Decode(&info)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if info.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_, err = cfg.db.UpdateUserToChirpyRed(r.Context(), info.Data.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
