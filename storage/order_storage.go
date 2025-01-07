package storage

import "AAHAOMS/OMS/models"

func (s *PostgresStorage) CreateOrder(customerID int, customerName, orderDate, shipmentDue, shipmentAddress string) (int, error) {
	query := `
		INSERT INTO orders (customer_id, customer_name, order_date, shipment_due, shipment_address)
		VALUES ($1, $2, $3, $4, $5) RETURNING id
	`
	var orderID int
	err := s.DB.QueryRow(query, customerID, customerName, orderDate, shipmentDue, shipmentAddress).Scan(&orderID)
	if err != nil {
		return 0, err
	}
	return orderID, nil

}
func (s *PostgresStorage) GetOrderHistoryByCustomerName(customerName string) ([]models.Order, error) {
	query := `
		SELECT o.id, o.customer_id, o.customer_name, o.order_date, o.shipment_due, o.shipment_address, 
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

func (s *PostgresStorage) AddItemToOrder(orderID int, name string, size, color *string, price float64, quantity int) error {
	query := `
		INSERT INTO items (order_id, name, size, color, price, quantity)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := s.DB.Exec(query, orderID, name, size, color, price, quantity)
	return err
}

func (s *PostgresStorage) GetOrderByID(orderID int) (*models.Order, error) {
	var order models.Order
	query := `
		SELECT o.id, o.customer_id, o.customer_name, o.order_date, o.shipment_due, o.shipment_address, COALESCE(SUM(i.price * i.quantity), 0) AS total_price, COUNT(i.id) AS no_of_items
		FROM orders o
		LEFT JOIN items i ON o.id = i.order_id
		WHERE o.id = $1
		GROUP BY o.id
	`
	err := s.DB.QueryRow(query, orderID).Scan(
		&order.ID,
		&order.CustomerID,
		&order.CustomerName,
		&order.OrderDate,
		&order.ShipmentDue,
		&order.ShipmentAddress,
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
		SELECT o.id, o.customer_id, o.customer_name, o.order_date, o.shipment_due, o.shipment_address, COALESCE(SUM(i.price * i.quantity), 0) AS total_price, COUNT(i.id) AS no_of_items
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
			&order.CustomerName, // Fetch customer name
			&order.OrderDate,
			&order.ShipmentDue,
			&order.ShipmentAddress,
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
