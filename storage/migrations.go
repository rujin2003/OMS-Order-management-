package storage

import (
	"database/sql"
	"fmt"
	"os"
)

func (s *PostgresStorage) Init() error {
	_, err := s.DB.Exec(`
			CREATE TABLE IF NOT EXISTS customers (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL UNIQUE,
		number VARCHAR(20),
		email VARCHAR(150),
		country VARCHAR(100),
		address TEXT
	);


	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL PRIMARY KEY,
		customer_id INT REFERENCES customers(id) ON DELETE CASCADE,
		customer_name VARCHAR(100),
		order_date DATE NOT NULL,
		shipment_due DATE,
		shipment_address TEXT,
		order_status VARCHAR(50) DEFAULT 'pending',
		total_price DECIMAL(10, 2),
		no_of_items INT DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS order_items (
		id SERIAL PRIMARY KEY,
		order_id INT REFERENCES orders(id) ON DELETE CASCADE,
		name VARCHAR(100),
		size VARCHAR(50),
		color VARCHAR(50),
		price DECIMAL(10, 2),
		quantity INT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS due_orders (
		id SERIAL PRIMARY KEY,
		order_id INT REFERENCES orders(id) ON DELETE CASCADE,
		item_id INT REFERENCES order_items(id) ON DELETE CASCADE,
		quantity INT NOT NULL,
		UNIQUE(order_id, item_id)
	);

	CREATE TABLE IF NOT EXISTS shipments (
		id SERIAL PRIMARY KEY,
		order_id INT REFERENCES orders(id) ON DELETE CASCADE,
		shipped_date DATE,
		due_order_type BOOLEAN,
		items INT[]
	);
	`)

	return err
}

//	func NewPostgresStorage() (*PostgresStorage, error) {
//		psqlInfo := "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"
//		db, err := sql.Open("postgres", psqlInfo)
//		if err != nil {
//			return nil, err
//		}
//		return &PostgresStorage{DB: db}, nil

//	}

func NewPostgresStorage() (*PostgresStorage, error) {

	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DBNAME")
	sslmode := os.Getenv("POSTGRES_SSLMODE")

	if host == "" || port == "" || user == "" || password == "" || dbname == "" || sslmode == "" {
		return nil, fmt.Errorf("one or more required environment variables are not set")
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("could not ping database: %v", err)
	}

	return &PostgresStorage{DB: db}, nil
}
