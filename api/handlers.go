package api

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

func wrapHandler(fn func(w http.ResponseWriter, r *http.Request)) apiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		fn(w, r)
		return nil
	}
}

func verifyToken(tokenString string) error {
	password := strings.TrimSpace(os.Getenv("PASSWORD"))

	if tokenString != password {
		return fmt.Errorf("invalid token")
	}
	return nil
}

func makeHandler(fn apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Missing Authorization header"}`, http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error": "Invalid Authorization header format"}`, http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimSpace(authHeader[len("Bearer "):])
		if err := verifyToken(tokenString); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusUnauthorized)
			return
		}

		if err := fn(w, r); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		}
	}
}

func (s *ApiServer) verifyJWTToken(tokenString string) error {

	exists, err := s.Store.IsKeyInStorage(tokenString)
	if err != nil {
		return fmt.Errorf("error checking key in storage: %s", err.Error())
	}

	if !exists {
		return fmt.Errorf("invalid or expired key")
	}

	return nil
}

func makeJWTHandler(fn apiFunc, s *ApiServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Missing Authorization header"}`, http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error": "Invalid Authorization header format"}`, http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimSpace(authHeader[len("Bearer "):])
		if err := s.verifyJWTToken(tokenString); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusUnauthorized)
			return
		}

		if err := fn(w, r); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)
		}
	}
}
