package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/matialvarez7/chirpy/internal/auth"
	"github.com/matialvarez7/chirpy/internal/database"
)

type loginInfo struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type response struct {
	User
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)

	info := loginInfo{}

	err := decoder.Decode(&info)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	dbUser, err := cfg.db.GetUserByEmail(r.Context(), info.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, errors.New("Incorrect email or password").Error())
			return
		}

		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	match, err := auth.CheckPassword(info.Password, dbUser.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !match {
		respondWithError(w, http.StatusUnauthorized, errors.New("Incorrect email or password").Error())
		return
	}

	token, err := auth.MakeJWT(dbUser.ID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	refreshToken := auth.MakeRefreshToken()

	refreshTokenParams := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    dbUser.ID,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
	}

	refreshTokenCreated, err := cfg.db.CreateRefreshToken(r.Context(), refreshTokenParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	user := response{
		User: User{
			ID:          dbUser.ID,
			CreatedAt:   dbUser.CreatedAt,
			UpdatedAt:   dbUser.UpdatedAt,
			Email:       dbUser.Email,
			IsChirpiRed: dbUser.IsChirpyRed,
		},
		Token:        token,
		RefreshToken: refreshTokenCreated.Token,
	}

	respondWithJSON(w, http.StatusOK, user)
}
