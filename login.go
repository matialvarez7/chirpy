package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/matialvarez7/chirpy/internal/auth"
)

type loginInfo struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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

	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}

	respondWithJSON(w, http.StatusOK, user)
}
