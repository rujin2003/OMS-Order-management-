package storage

import (
	"AAHAOMS/OMS/models"
	"database/sql"

	_ "github.com/lib/pq"
)

func NewPostgresStorage() (*PostgresStorage, error) {
	psqlInfo := "host=localhost port=5432 user=postgres password=password dbname=order_management sslmode=disable"
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	return &PostgresStorage{DB: db}, nil
}

func (s *PostgresStorage) GetAllCustomers() ([]models.Customer, error) {
	rows, err := s.DB.Query("SELECT id, name, number, email, country, address FROM customers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var customer models.Customer
		if err := rows.Scan(&customer.ID, &customer.Name, &customer.Number, &customer.Email, &customer.Country, &customer.Address); err != nil {
			return nil, err
		}
		customers = append(customers, customer)
	}
	return customers, nil
}

func (s *PostgresStorage) GetCustomerByID(id string) (*models.Customer, error) {
	var customer models.Customer
	err := s.DB.QueryRow(
		"SELECT id, name, number, email, country, address FROM customers WHERE id = $1",
		id,
	).Scan(&customer.ID, &customer.Name, &customer.Number, &customer.Email, &customer.Country, &customer.Address)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &customer, nil
}

func (s *PostgresStorage) CreateCustomer(name string, number int, email string, country string, address string) (int, error) {
	var id int
	query := `
		INSERT INTO customers (name, number, email, country, address)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	err := s.DB.QueryRow(query, name, number, email, country, address).Scan(&id)
	return id, err
}
