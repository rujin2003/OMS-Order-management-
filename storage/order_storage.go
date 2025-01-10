package storage

import (
	"AAHAOMS/OMS/models"
)

func (s *PostgresStorage) CreateOrder(customerID int, customerName, orderDate, shipmentDue, shipmentAddress, orderStatus string) (int, error) {
	query := `
		INSERT INTO orders (customer_id, customer_name, order_date, shipment_due, shipment_address, order_status)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
	`
	var orderID int
	err := s.DB.QueryRow(query, customerID, customerName, orderDate, shipmentDue, shipmentAddress, orderStatus).Scan(&orderID)
	return orderID, err
}

func (s *PostgresStorage) GetOrderHistoryByCustomerName(customerName string) ([]models.Order, error) {
	query := `
		SELECT o.id, o.customer_id, o.customer_name, o.order_date, o.shipment_due, o.shipment_address, o.order_status,
		       COALESCE(SUM(i.price * i.quantity), 0) AS total_price, COUNT(i.id) AS no_of_items
		FROM orders o
		LEFT JOIN items i ON o.id = i.order_id
		WHERE o.customer_name ILIKE $1
		GROUP BY o.id
		ORDER BY o.order_date DESC
	`
	rows, err := s.DB.Query(query, "%"+customerName+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID,
			&order.CustomerID,
			&order.CustomerName,
			&order.OrderDate,
			&order.ShipmentDue,
			&order.ShipmentAddress,
			&order.OrderStatus,
			&order.TotalPrice,
			&order.NoOfItems,
		)
		if err != nil {
			return nil, err
		}

		// Fetch items for this order
		itemsQuery := `
			SELECT id, name, size, color, price, quantity, shipped
			FROM items
			WHERE order_id = $1
		`
		itemRows, err := s.DB.Query(itemsQuery, order.ID)
		if err != nil {
			return nil, err
		}
		defer itemRows.Close()

		var items []models.Item
		for itemRows.Next() {
			var item models.Item
			err := itemRows.Scan(&item.ID, &item.Name, &item.Size, &item.Color, &item.Price, &item.Quantity, &item.Shipped)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
		order.Items = items

		orders = append(orders, order)
	}

	return orders, nil
}

func (s *PostgresStorage) GetOrderByID(orderID int) (*models.Order, error) {
	query := `
		SELECT o.id, o.customer_id, o.customer_name, o.order_date, o.shipment_due, o.shipment_address, o.order_status,
		       COALESCE(SUM(i.price * i.quantity), 0) AS total_price, COUNT(i.id) AS no_of_items
		FROM orders o
		LEFT JOIN items i ON o.id = i.order_id
		WHERE o.id = $1
		GROUP BY o.id
	`
	var order models.Order
	err := s.DB.QueryRow(query, orderID).Scan(
		&order.ID,
		&order.CustomerID,
		&order.CustomerName,
		&order.OrderDate,
		&order.ShipmentDue,
		&order.ShipmentAddress,
		&order.OrderStatus,
		&order.TotalPrice,
		&order.NoOfItems,
	)
	if err != nil {
		return nil, err
	}

	// Fetch items
	itemsQuery := `
		SELECT id, name, size, color, price, quantity, shipped
		FROM items
		WHERE order_id = $1
	`
	rows, err := s.DB.Query(itemsQuery, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		err := rows.Scan(&item.ID, &item.Name, &item.Size, &item.Color, &item.Price, &item.Quantity, &item.Shipped)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	order.Items = items

	return &order, nil
}

func (s *PostgresStorage) GetAllOrders() ([]models.Order, error) {
	query := `
		SELECT o.id, o.customer_id, o.customer_name, o.order_date, o.shipment_due, o.shipment_address, o.order_status,
		       COALESCE(SUM(i.price * i.quantity), 0) AS total_price, COUNT(i.id) AS no_of_items
		FROM orders o
		LEFT JOIN items i ON o.id = i.order_id
		GROUP BY o.id
		ORDER BY o.order_date DESC
	`
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID,
			&order.CustomerID,
			&order.CustomerName,
			&order.OrderDate,
			&order.ShipmentDue,
			&order.ShipmentAddress,
			&order.OrderStatus,
			&order.TotalPrice,
			&order.NoOfItems,
		)
		if err != nil {
			return nil, err
		}

		// Fetch items
		itemsQuery := `
			SELECT id, name, size, color, price, quantity, shipped
			FROM items
			WHERE order_id = $1
		`
		itemRows, err := s.DB.Query(itemsQuery, order.ID)
		if err != nil {
			return nil, err
		}
		defer itemRows.Close()

		var items []models.Item
		for itemRows.Next() {
			var item models.Item
			err := itemRows.Scan(&item.ID, &item.Name, &item.Size, &item.Color, &item.Price, &item.Quantity, &item.Shipped)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
		order.Items = items

		orders = append(orders, order)
	}

	return orders, nil
}
func (s *PostgresStorage) UpdateOrderStatus(orderID int, status string) error {
	query := `
		UPDATE orders
		SET order_status = $1
		WHERE id = $2
	`
	_, err := s.DB.Exec(query, status, orderID)
	return err
}


func (s *PostgresStorage) DeleteOrder(order int) error {
	query := `DELETE FROM order WHERE id = $1`
	_, err := s.DB.Exec(query, order)
	return err
}

func (s *PostgresStorage) GetTotalOrderValueByCustomerName(customerName string) (float64, error) {
	query := `
		SELECT COALESCE(SUM(i.price * i.quantity), 0) AS total_value
		FROM orders o
		LEFT JOIN items i ON o.id = i.order_id
		WHERE o.customer_name ILIKE $1
	`
	var totalValue float64
	err := s.DB.QueryRow(query, "%"+customerName+"%").Scan(&totalValue)
	return totalValue, err
}

func (s *PostgresStorage) GetOrderCountByCustomerName(customerName string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM orders
		WHERE customer_name ILIKE $1
	`
	var orderCount int
	err := s.DB.QueryRow(query, "%"+customerName+"%").Scan(&orderCount)
	return orderCount, err
}

func (s *PostgresStorage) GetPendingOrderCount() (int, error) {
	query := `
		SELECT COUNT(*)
		FROM orders
		WHERE order_status = 'pending'
	`
	var pendingCount int
	err := s.DB.QueryRow(query).Scan(&pendingCount)
	return pendingCount, err
}
