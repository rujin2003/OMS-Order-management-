package storage

func (s *PostgresStorage) Init() error {
	_, err := s.DB.Exec(`
		CREATE TABLE IF NOT EXISTS customers (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL UNIQUE
		);

		CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			customer_id INT REFERENCES customers(id),
			order_date DATE,
			shipment_due DATE
		);

		CREATE TABLE IF NOT EXISTS items (
			id SERIAL PRIMARY KEY,
			order_id INT REFERENCES orders(id),
			name VARCHAR(100),
			size VARCHAR(50),
			color VARCHAR(50),
			price DECIMAL(10, 2),
			shipped BOOLEAN DEFAULT FALSE
		);

		CREATE TABLE IF NOT EXISTS shipments (
			id SERIAL PRIMARY KEY,
			order_id INT REFERENCES orders(id),
			shipped_date DATE,
			items TEXT[]
		);
	`)
	return err
}
