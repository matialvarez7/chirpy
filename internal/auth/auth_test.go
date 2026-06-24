package auth

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	id := uuid.New()
	tokenSecret := "secret"
	expiresIn := 3 * time.Minute

	token, err := MakeJWT(id, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("Error generating token: %s", err)
	}

	if token == "" {
		t.Errorf("Expected a signed token string, got an empty string")
	}
}

func TestValidateJWT(t *testing.T) {

	mockID := uuid.New()
	tokenSecret := "secreto"
	duration := time.Hour

	validToken, err := MakeJWT(mockID, tokenSecret, duration)
	if err != nil {
		t.Fatalf("Unespected error: %v", err)
	}

	expiredToken, err := MakeJWT(mockID, tokenSecret, -time.Second)
	if err != nil {
		t.Fatalf("Unespected error: %v", err)
	}

	tests := []struct {
		name        string
		tokenString string
		tokenSecret string
		wantUserId  uuid.UUID
		wantErr     bool
	}{
		{
			name:        "valid token",
			tokenString: validToken,
			tokenSecret: tokenSecret,
			wantUserId:  mockID,
			wantErr:     false,
		},
		{
			name:        "expired token",
			tokenString: expiredToken,
			tokenSecret: tokenSecret,
			wantUserId:  uuid.Nil,
			wantErr:     true,
		},
		{
			name:        "invalid secret",
			tokenString: validToken,
			tokenSecret: "error",
			wantUserId:  uuid.Nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			returnedID, err := ValidateJWT(tt.tokenString, tt.tokenSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("got err = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if returnedID != tt.wantUserId {
				t.Errorf("want %v, got %v", tt.wantUserId, returnedID)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {

	tests := []struct {
		name      string
		headers   http.Header
		wantToken string
		wantErr   bool
	}{
		{
			name: "Bearer stripped",
			headers: http.Header{
				"Authorization": []string{"Bearer test1"},
			},
			wantToken: "test1",
			wantErr:   false,
		},
		{
			name: "Token whitout Bearer",
			headers: http.Header{
				"Authorization": []string{"test2"},
			},
			wantToken: "test2",
			wantErr:   false,
		},
		{
			name:      "Return error",
			headers:   http.Header{},
			wantToken: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GetBearerToken(tt.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("got err = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if strings.Contains(token, "Bearer ") {
				t.Errorf("want %v, got %v", tt.wantToken, token)
			}
		})
	}
}
