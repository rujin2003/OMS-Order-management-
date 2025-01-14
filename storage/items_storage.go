package storage

import (
	"AAHAOMS/OMS/models"
	"database/sql"
)

func (s *PostgresStorage) GetItemByID(itemID int) (models.Item, error) {
	var item models.Item
	query := `
		SELECT id, name, size, color, price, quantity
		FROM order_items
		WHERE id = $1
	`
	err := s.DB.QueryRow(query, itemID).Scan(
		&item.ID, &item.Name, &item.Size, &item.Color, &item.Price, &item.Quantity,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Item{}, nil
		}
		return models.Item{}, err
	}
	return item, nil
}
