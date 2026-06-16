package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/matialvarez7/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID     `json:"id"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Body      string        `json:"body"`
	UserID    uuid.NullUUID `json:"user_id"`
}

type chirpParams struct {
	Body   string        `json:"body"`
	UserId uuid.NullUUID `json:"user_id"`
}

func (cfg *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	reqChirp := chirpParams{}
	err := decoder.Decode(&reqChirp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(reqChirp.Body) >= 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	reqChirp.Body = cleanMessage(reqChirp.Body)

	newChirp := database.CreateChirpParams{
		Body:   reqChirp.Body,
		UserID: reqChirp.UserId,
	}

	dbChirp, err := cfg.db.CreateChirp(r.Context(), newChirp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}

	respondWithJSON(w, http.StatusCreated, response)
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
