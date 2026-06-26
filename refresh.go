package main

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/matialvarez7/chirpy/internal/auth"
)

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	dbRefreshToken, err := cfg.db.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if dbRefreshToken.RevokedAt.Valid || dbRefreshToken.ExpiresAt.Before(time.Now().UTC()) {
		respondWithError(w, http.StatusUnauthorized, errors.New("Token revoked").Error())
		return
	}

	accessToken, err := auth.MakeJWT(dbRefreshToken.UserID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	type refreshResponse struct {
		Token string `json:"token"`
	}

	refreshRes := refreshResponse{
		Token: accessToken,
	}

	respondWithJSON(w, http.StatusOK, refreshRes)
}
