package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	auth "github.com/matialvarez7/chirpy/internal/auth"
	"github.com/matialvarez7/chirpy/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type userInfo struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)

	userInf := userInfo{}

	err := decoder.Decode(&userInf)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	userInf.Password, err = auth.HashPassword(userInf.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	userParams := database.CreateUserParams{
		Email:          userInf.Email,
		HashedPassword: userInf.Password,
	}

	dbUser, err := cfg.db.CreateUser(r.Context(), userParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}

	respondWithJSON(w, http.StatusCreated, user)
}
