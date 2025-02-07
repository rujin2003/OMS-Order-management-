package storage

import "AAHAOMS/OMS/models"

func (s *PostgresStorage) AddAuthUser(user models.AuthUser) error {
	query := `INSERT INTO auth_users (email) VALUES ($1) ON CONFLICT (email) DO NOTHING`
	_, err := s.DB.Exec(query, user.Email)
	return err
}

func (s *PostgresStorage) AddOtp(user models.AuthUser) error {
	query := `INSERT INTO otp (email, key) VALUES ($1, $2)`
	_, err := s.DB.Exec(query, user.Email, user.OTP)
	return err
}

func (s *PostgresStorage) IsUserExists(email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM auth_users WHERE email = $1)`
	var exists bool
	err := s.DB.QueryRow(query, email).Scan(&exists)
	return exists, err
}

func (s *PostgresStorage) VerifyOtp(user models.AuthUser) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM otp WHERE email = $1 AND key = $2)`
	var exists bool
	err := s.DB.QueryRow(query, user.Email, user.OTP).Scan(&exists)
	return exists, err
}

func (s *PostgresStorage) IsKeyInStorage(token string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM otp WHERE key = $1)`
	var exists bool
	err := s.DB.QueryRow(query, token).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
