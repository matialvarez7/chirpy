package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}

	return hash, nil
}

func CheckPassword(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}

	return match, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := &jwt.RegisteredClaims{
		Issuer:    "chirpy-acces",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}
	keyFunc := func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &claims, keyFunc)
	if err != nil {
		return uuid.Nil, err
	}

	id, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	parsedId, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, err
	}
	return parsedId, nil
}

func GetBearerToken(header http.Header) (string, error) {
	token := header.Get("Authorization")
	if token == "" {
		err := errors.New("The header was empty or doesn't exist")
		return "", err
	}

	token = strings.TrimPrefix(token, "Bearer ")

	return token, nil
}

func MakeRefreshToken() string {
	token := make([]byte, 32)
	rand.Read(token)
	encodedToken := hex.EncodeToString(token)
	return encodedToken
}

func GetAPIKey(headers http.Header) (string, error) {
	apikey := headers.Get("Authorization")
	if apikey == "" {
		err := errors.New("The header was empty or doesn't exist")
		return "", err
	}

	apikey = strings.TrimPrefix(apikey, "ApiKey ")

	return apikey, nil
}
