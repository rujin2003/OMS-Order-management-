package storage

import (
	"AAHAOMS/OMS/models"
	"database/sql"
	"fmt"
	"log"
)

func (s *PostgresStorage) getOrderStatus(tx *sql.Tx, orderID int) (string, error) {
	var orderStatus string
	err := tx.QueryRow(`SELECT order_status FROM orders WHERE id = $1`, orderID).Scan(&orderStatus)
	if err != nil {
		log.Printf("Invalid order ID: %v", err)
		return "", fmt.Errorf("invalid order ID: %v", err)
	}
	return orderStatus, nil
}

func (s *PostgresStorage) GetTotalSalesForShippedOrders() (float64, error) {
	query := `
		SELECT COALESCE(SUM(o.total_price), 0) AS total_sales
		FROM orders o
		WHERE TRIM(o.order_status) = 'shipped'
	`
	var totalSales float64
	err := s.DB.QueryRow(query).Scan(&totalSales)
	if err != nil {
		log.Printf("Error calculating total sales for shipped orders: %v", err)
		return 0, err
	}
	log.Printf("Total sales for shipped orders: %f", totalSales)
	return totalSales, nil
}

func (s *PostgresStorage) GetTotalSalesForShippedOrdersByCustomer(customerName string) (float64, error) {
	query := `
		SELECT COALESCE(SUM(o.total_price), 0) AS total_sales
		FROM orders o
		WHERE TRIM(o.order_status) = 'shipped' AND TRIM(o.customer_name) ILIKE $1
	`
	var totalSales float64
	err := s.DB.QueryRow(query, "%"+customerName+"%").Scan(&totalSales)
	if err != nil {
		log.Printf("Error calculating total sales for shipped orders by customer %s: %v", customerName, err)
		return 0, err
	}
	log.Printf("Total sales for shipped orders by customer %s: %f", customerName, totalSales)
	return totalSales, nil
}

func (s *PostgresStorage) TotalOrderCount() (int, error) {
	query := `
		SELECT COUNT(*)
		FROM orders
	`
	var totalCount int
	err := s.DB.QueryRow(query).Scan(&totalCount)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return 0, err
	}
	log.Printf("Query executed successfully. Total order count: %d", totalCount)
	return totalCount, nil
}

func (s *PostgresStorage) CreateOrder(order models.Order) (int, error) {
	if order.OrderStatus == "" {
		order.OrderStatus = "pending"
	}

	query := `
		INSERT INTO orders (customer_id, customer_name, order_date, shipment_due, shipment_address, order_status, total_price, no_of_items)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	var orderID int
	err := s.DB.QueryRow(
		query,
		order.CustomerID,
		order.CustomerName,
		order.OrderDate,
		order.ShipmentDue,
		order.ShipmentAddress,
		order.OrderStatus,
		order.TotalPrice,
		order.NoOfItems,
	).Scan(&orderID)
	if err != nil {
		return 0, err
	}

	for _, item := range order.Items {
		itemQuery := `
			INSERT INTO order_items (order_id, name, size, color, price, quantity)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`
		var itemID int
		err = s.DB.QueryRow(itemQuery, orderID, item.Name, item.Size, item.Color, item.Price, item.Quantity).Scan(&itemID)
		if err != nil {
			return 0, err
		}
		item.ID = itemID // Assign the auto-generated ID back to the item struct
	}

	return orderID, nil
}

func (s *PostgresStorage) GetTotalOrderCount() (int, error) {
	query := `
		SELECT COUNT(*)
		FROM orders
	`
	var totalCount int
	err := s.DB.QueryRow(query).Scan(&totalCount)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return 0, err
	}
	log.Printf("Query executed successfully. Total order count: %d", totalCount)
	return totalCount, nil
}

func (s *PostgresStorage) GetLatestOrderID() (int, error) {
	query := `
		SELECT id
		FROM orders
		ORDER BY order_date DESC
		LIMIT 1
	`
	var latestOrderID int
	err := s.DB.QueryRow(query).Scan(&latestOrderID)

	// If no rows are found, return 0 instead of an error
	if err == sql.ErrNoRows {
		log.Println("No orders found. Returning 0 as latest order ID.")
		return 0, nil
	}

	// Handle other possible errors
	if err != nil {
		log.Printf("Error fetching latest order ID: %v", err)
		return 0, err
	}

	log.Printf("Latest order ID: %d", latestOrderID)
	return latestOrderID, nil
}

func (s *PostgresStorage) GetOrderHistoryByCustomerName(name string) ([]models.Order, error) {
	query := `
		SELECT id, customer_id, customer_name, order_date, shipment_due, shipment_address, order_status, total_price, no_of_items
		FROM orders
		WHERE customer_name = $1
	`
	rows, err := s.DB.Query(query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err = rows.Scan(&order.ID, &order.CustomerID, &order.CustomerName, &order.OrderDate, &order.ShipmentDue, &order.ShipmentAddress, &order.OrderStatus, &order.TotalPrice, &order.NoOfItems)
		if err != nil {
			return nil, err
		}

		// Fetch items for this order
		itemQuery := `
			SELECT id, name, size, color, price, quantity
			FROM order_items
			WHERE order_id = $1
		`
		itemRows, err := s.DB.Query(itemQuery, order.ID)
		if err != nil {
			return nil, err
		}
		defer itemRows.Close()

		var items []models.Item
		for itemRows.Next() {
			var item models.Item
			err = itemRows.Scan(&item.ID, &item.Name, &item.Size, &item.Color, &item.Price, &item.Quantity)
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

func (s *PostgresStorage) GetOrderByID(orderID int) (models.Order, error) {
	var order models.Order
	query := `
		SELECT id, customer_id, customer_name, order_date, shipment_due, shipment_address, order_status, total_price, no_of_items
		FROM orders
		WHERE id = $1
	`
	err := s.DB.QueryRow(query, orderID).Scan(
		&order.ID, &order.CustomerID, &order.CustomerName, &order.OrderDate, &order.ShipmentDue, &order.ShipmentAddress,
		&order.OrderStatus, &order.TotalPrice, &order.NoOfItems,
	)
	if err != nil {
		return models.Order{}, err
	}

	itemQuery := `
		SELECT id, name, size, color, price, quantity
		FROM order_items
		WHERE order_id = $1
	`
	itemRows, err := s.DB.Query(itemQuery, order.ID)
	if err != nil {
		return models.Order{}, err
	}
	defer itemRows.Close()

	var items []models.Item
	for itemRows.Next() {
		var item models.Item
		err = itemRows.Scan(&item.ID, &item.Name, &item.Size, &item.Color, &item.Price, &item.Quantity)
		if err != nil {
			return models.Order{}, err
		}
		items = append(items, item)
	}
	order.Items = items

	return order, nil
}

func (s *PostgresStorage) GetAllOrders() ([]models.Order, error) {
	query := `
		SELECT id, customer_id, customer_name, order_date, shipment_due, shipment_address, order_status, total_price, no_of_items
		FROM orders
	`
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err = rows.Scan(
			&order.ID, &order.CustomerID, &order.CustomerName, &order.OrderDate, &order.ShipmentDue, &order.ShipmentAddress,
			&order.OrderStatus, &order.TotalPrice, &order.NoOfItems,
		)
		if err != nil {
			return nil, err
		}

		// Fetch items for each order
		itemQuery := `
			SELECT id, name, size, color, price, quantity
			FROM order_items
			WHERE order_id = $1
		`
		itemRows, err := s.DB.Query(itemQuery, order.ID)
		if err != nil {
			return nil, err
		}
		defer itemRows.Close()

		var items []models.Item
		for itemRows.Next() {
			var item models.Item
			err = itemRows.Scan(&item.ID, &item.Name, &item.Size, &item.Color, &item.Price, &item.Quantity)
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

func (s *PostgresStorage) DeleteOrder(orderID int) error {

	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	shipmentQuery := `DELETE FROM shipments WHERE order_id = $1`
	_, err = tx.Exec(shipmentQuery, orderID)
	if err != nil {
		return fmt.Errorf("failed to delete related shipments: %v", err)
	}

	itemQuery := `DELETE FROM order_items WHERE order_id = $1`
	_, err = tx.Exec(itemQuery, orderID)
	if err != nil {
		return fmt.Errorf("failed to delete related order items: %v", err)
	}

	orderQuery := `DELETE FROM orders WHERE id = $1`
	_, err = tx.Exec(orderQuery, orderID)
	if err != nil {
		return fmt.Errorf("failed to delete order: %v", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// // *****************************************************************************///
func (s *PostgresStorage) UpdateOrderStatus(orderID int, status string) error {
	query := `
		UPDATE orders
		SET order_status = $1
		WHERE id = $2
	`
	_, err := s.DB.Exec(query, status, orderID)
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
	WHERE TRIM(order_status) = 'pending'
`
	var pendingCount int
	err := s.DB.QueryRow(query).Scan(&pendingCount)

	if err != nil {

		log.Printf("Error executing query: %v", err)
		return 0, err
	}
	log.Printf("Query executed successfully. Pending order count: %d", pendingCount)
	return pendingCount, nil
}

func (s *PostgresStorage) GetOrdersByNameAndDate(customerName, orderDate string) ([]models.Order, error) {
	query := `
		SELECT id, customer_id, customer_name, order_date, shipment_due, shipment_address, order_status, total_price, no_of_items
		FROM orders
		WHERE customer_name ILIKE $1 AND order_date = $2
	`

	rows, err := s.DB.Query(query, "%"+customerName+"%", orderDate)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err = rows.Scan(
			&order.ID, &order.CustomerID, &order.CustomerName, &order.OrderDate, &order.ShipmentDue,
			&order.ShipmentAddress, &order.OrderStatus, &order.TotalPrice, &order.NoOfItems,
		)
		if err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}

		itemQuery := `
			SELECT id, name, size, color, price, quantity
			FROM order_items
			WHERE order_id = $1
		`
		itemRows, err := s.DB.Query(itemQuery, order.ID)
		if err != nil {
			return nil, fmt.Errorf("item query error: %w", err)
		}
		defer itemRows.Close()

		var items []models.Item
		for itemRows.Next() {
			var item models.Item
			err = itemRows.Scan(&item.ID, &item.Name, &item.Size, &item.Color, &item.Price, &item.Quantity)
			if err != nil {
				return nil, fmt.Errorf("item scan error: %w", err)
			}
			items = append(items, item)
		}
		order.Items = items
		orders = append(orders, order)
	}

	return orders, nil
}

func (s *PostgresStorage) GetRecentOrders(limit int) ([]models.Order, error) {
	query := `
		SELECT id, customer_id, customer_name, order_date, shipment_due, shipment_address, order_status, total_price, no_of_items
		FROM orders
		ORDER BY order_date DESC
		LIMIT $1
	`
	rows, err := s.DB.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recent orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err = rows.Scan(
			&order.ID, &order.CustomerID, &order.CustomerName, &order.OrderDate, &order.ShipmentDue,
			&order.ShipmentAddress, &order.OrderStatus, &order.TotalPrice, &order.NoOfItems,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		// Fetch items for the order
		itemQuery := `
			SELECT id, name, size, color, price, quantity
			FROM order_items
			WHERE order_id = $1
		`
		itemRows, err := s.DB.Query(itemQuery, order.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch order items: %w", err)
		}
		defer itemRows.Close()

		var items []models.Item
		for itemRows.Next() {
			var item models.Item
			err = itemRows.Scan(&item.ID, &item.Name, &item.Size, &item.Color, &item.Price, &item.Quantity)
			if err != nil {
				return nil, fmt.Errorf("failed to scan order item: %w", err)
			}
			items = append(items, item)
		}
		order.Items = items
		orders = append(orders, order)
	}

	return orders, nil
}

func (s *PostgresStorage) GetCompletedShipments() ([]models.Shipment, error) {
	query := `
		SELECT s.id, s.order_id, s.shipped_date::DATE, s.items::int[], s.due_order_type
		FROM shipments s
		INNER JOIN orders o ON s.order_id = o.id
		WHERE o.order_status = 'shipped'
	`

	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve completed shipments: %v", err)
	}
	defer rows.Close()

	return s.parseShipments(rows)
}
func (s *PostgresStorage) DeleteShipment(shipmentID int) error {

	var orderID int
	getOrderQuery := `SELECT order_id FROM shipments WHERE id = $1`
	err := s.DB.QueryRow(getOrderQuery, shipmentID).Scan(&orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("shipment ID %d not found", shipmentID)
		}
		return fmt.Errorf("failed to fetch order ID for shipment ID %d: %v", shipmentID, err)
	}

	deleteShipmentQuery := `DELETE FROM shipments WHERE id = $1`
	_, err = s.DB.Exec(deleteShipmentQuery, shipmentID)
	if err != nil {
		return fmt.Errorf("failed to delete shipment ID %d: %v", shipmentID, err)
	}

	updateOrderQuery := `UPDATE orders SET order_status = 'pending' WHERE id = $1`
	_, err = s.DB.Exec(updateOrderQuery, orderID)
	if err != nil {
		return fmt.Errorf("failed to reset order status for order ID %d: %v", orderID, err)
	}

	return nil
}
func (s *PostgresStorage) GetShippedButPendingShipments() ([]models.Shipment, error) {
	query := `
		SELECT s.id, s.order_id, s.shipped_date::DATE, s.items::int[], s.due_order_type
		FROM shipments s
		INNER JOIN orders o ON s.order_id = o.id
		WHERE o.order_status = 'shipped and due'
	`

	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve shipped but pending shipments: %v", err)
	}
	defer rows.Close()

	return s.parseShipments(rows)
}

func (s *PostgresStorage) getOrderItems(tx *sql.Tx, orderID int) (map[int]models.Item, error) {
	rows, err := tx.Query(`SELECT id, name, quantity FROM order_items WHERE order_id = $1`, orderID)
	if err != nil {
		log.Printf("Error fetching order items: %v", err)
		return nil, fmt.Errorf("failed to fetch order items: %v", err)
	}
	defer rows.Close()

	orderItems := make(map[int]models.Item)
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Quantity); err != nil {
			log.Printf("Error scanning order item: %v", err)
			return nil, fmt.Errorf("failed to scan order item: %v", err)
		}
		orderItems[item.ID] = item
	}
	return orderItems, nil
}
