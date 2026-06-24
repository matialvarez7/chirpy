package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/matialvarez7/chirpy/internal/auth"
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
	Body string `json:"body"`
}

func (cfg *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	reqChirp := chirpParams{}
	err := decoder.Decode(&reqChirp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if len(reqChirp.Body) >= 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	reqChirp.Body = cleanMessage(reqChirp.Body)

	newChirp := database.CreateChirpParams{
		Body: reqChirp.Body,
		UserID: uuid.NullUUID{
			UUID:  userID,
			Valid: true,
		},
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

func (cfg *apiConfig) listChirps(w http.ResponseWriter, r *http.Request) {

	dbChirps, err := cfg.db.ListChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	chirps := mapChirps(dbChirps)

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) getChirp(w http.ResponseWriter, r *http.Request) {
	chirpId := r.PathValue("chirpID")
	dbChirp, err := cfg.db.GetChirp(r.Context(), uuid.MustParse(chirpId))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	chirp := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}

	respondWithJSON(w, http.StatusOK, chirp)
}

func mapChirps(dbChirps []database.Chirp) []Chirp {

	allChirps := []Chirp{}
	if len(dbChirps) > 0 {
		for _, dbChirp := range dbChirps {
			newChirp := Chirp{
				ID:        dbChirp.ID,
				CreatedAt: dbChirp.CreatedAt,
				UpdatedAt: dbChirp.UpdatedAt,
				Body:      dbChirp.Body,
				UserID:    dbChirp.UserID,
			}

			allChirps = append(allChirps, newChirp)
		}
	}

	return allChirps
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
