package storage

import (
	"AAHAOMS/OMS/models"
	"database/sql"

	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	DB *sql.DB
}

func (s *PostgresStorage) CreateShipment(orderID int, itemIDs []int) error {
	_, err := s.DB.Exec("INSERT INTO shipments (order_id, shipped_date, item_ids) VALUES ($1, CURRENT_DATE, $2)", orderID, itemIDs)
	if err != nil {
		return err
	}
	for _, itemID := range itemIDs {
		if _, err := s.DB.Exec("UPDATE items SET shipped = TRUE WHERE id = $1", itemID); err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStorage) GetOrderHistoryByName(customerName string) ([]models.Order, error) {
	var orders []models.Order

	// Query to fetch orders for the given customer
	queryOrders := `
		SELECT o.id, o.customer_id, o.order_date, o.shipment_due
		FROM orders o
		JOIN customers c ON o.customer_id = c.id
		WHERE c.name = $1
	`

	rows, err := s.DB.Query(queryOrders, customerName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate over orders
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.ID, &order.CustomerID, &order.OrderDate, &order.ShipmentDue); err != nil {
			return nil, err
		}

		// Fetch items associated with the current order
		queryItems := `
			SELECT id, name, size, color, price
			FROM items
			WHERE order_id = $1
		`

		itemRows, err := s.DB.Query(queryItems, order.ID)
		if err != nil {
			return nil, err
		}

		var items []models.Item
		for itemRows.Next() {
			var item models.Item
			if err := itemRows.Scan(&item.ID, &item.Name, &item.Size, &item.Color, &item.Price); err != nil {
				itemRows.Close()
				return nil, err
			}
			items = append(items, item)
		}
		itemRows.Close()

		// Attach items to the order
		order.Items = items
		orders = append(orders, order)
	}

	return orders, nil
}

func (s *PostgresStorage) GetShipmentHistoryByName(name string) ([]models.Shipment, error) {
	query := `
		SELECT s.id, s.shipped_date, s.order_id
		FROM shipments s
		JOIN orders o ON s.order_id = o.id
		JOIN customers c ON o.customer_id = c.id
		WHERE c.name = $1
	`

	rows, err := s.DB.Query(query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shipments []models.Shipment
	for rows.Next() {
		var shipment models.Shipment
		if err := rows.Scan(&shipment.ID, &shipment.ShippedDate, &shipment.OrderID); err != nil {
			return nil, err
		}

		// Fetch the item IDs for the shipment
		itemQuery := `
			SELECT item_ids
			FROM shipments
			WHERE id = $1
		`
		var itemIDs []int
		err = s.DB.QueryRow(itemQuery, shipment.ID).Scan(&itemIDs)
		if err != nil {
			return nil, err
		}
		shipment.Items = itemIDs

		shipments = append(shipments, shipment)
	}

	return shipments, nil
}

func (s *PostgresStorage) Close() {
	s.DB.Close()
}
