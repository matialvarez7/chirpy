package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/matialvarez7/chirpy/internal/auth"
)

type loginInfo struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds *int   `json:"expires_in_seconds"`
}

type response struct {
	User
	Token string `json:"token"`
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

	expiresIn := 3600
	if info.ExpiresInSeconds != nil && *info.ExpiresInSeconds < 3600 {
		expiresIn = *info.ExpiresInSeconds
	}

	duration := time.Duration(expiresIn) * time.Second

	token, err := auth.MakeJWT(dbUser.ID, cfg.secret, duration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	user := response{
		User: User{
			ID:        dbUser.ID,
			CreatedAt: dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
			Email:     dbUser.Email,
		},
		Token: token,
	}

	respondWithJSON(w, http.StatusOK, user)
}
