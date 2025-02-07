package models

type AuthUser struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}
