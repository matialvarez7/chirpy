package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

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
