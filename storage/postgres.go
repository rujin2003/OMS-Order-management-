package storage

import (
	
	"database/sql"

	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	DB *sql.DB
}

func (s *PostgresStorage) Close() {
	s.DB.Close()
}
