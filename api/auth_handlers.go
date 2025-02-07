package api

import (
	"AAHAOMS/OMS/models"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateUniqueKey() string {
	bytes := make([]byte, 32)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func sendMailtrap(email, token string) error {
	auth := smtp.PlainAuth("", os.Getenv("MAILTRAP_USER"), os.Getenv("MAILTRAP_PASS"), "smtp.mailtrap.io")
	to := []string{email}
	msg := []byte(fmt.Sprintf("Subject: Your Login Token\n\nUse this token to authenticate: %s", token))
	return smtp.SendMail("smtp.mailtrap.io:587", auth, "no-reply@example.com", to, msg)
}

func generateJWT(email, secretKey string) (string, error) {
	claims := &jwt.RegisteredClaims{
		Subject:   email,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey)) // Use the unique key
}

func (s *ApiServer) sendMailtrap(email, token string) error {
	auth := smtp.PlainAuth("", os.Getenv("MAILTRAP_USER"), os.Getenv("MAILTRAP_PASS"), "smtp.mailtrap.io")
	to := []string{email}
	msg := []byte(fmt.Sprintf("Subject: Your Login Token\n\nUse this token to authenticate: %s", token))
	return smtp.SendMail("smtp.mailtrap.io:587", auth, "no-reply@example.com", to, msg)
}

func (s *ApiServer) loginHandler(w http.ResponseWriter, r *http.Request) error {
	email := r.FormValue("email")

	// Check if email exists in auth_users
	exists, err := s.Store.IsUserExists(email)
	if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return nil
	}

	if !exists {
		http.Error(w, `{"error": "Email not registered"}`, http.StatusUnauthorized)
		return nil
	}

	// Generate a new OTP
	otp := generateUniqueKey()

	// Store OTP in DB
	err = s.Store.AddOtp(models.AuthUser{Email: email, OTP: otp})
	if err != nil {
		http.Error(w, `{"error": "Failed to store OTP"}`, http.StatusInternalServerError)
		return nil
	}

	// Send OTP via email
	err = sendMailtrap(email, otp)
	if err != nil {
		http.Error(w, `{"error": "Failed to send email"}`, http.StatusInternalServerError)
		return nil
	}

	w.Write([]byte(`{"message": "OTP sent to email"}`))
	return nil
}
