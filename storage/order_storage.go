package storage

import (
	"AAHAOMS/OMS/models"
	"fmt"
	"log"
)

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

func (s *PostgresStorage) CreateOrder(order models.Order) (int, error) {
	if order.OrderStatus == "" {
		order.OrderStatus = "pending"
	}

	query := `
		INSERT INTO orders (id, customer_id, customer_name, order_date, shipment_due, shipment_address, order_status, total_price, no_of_items)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	var orderID int
	err := s.DB.QueryRow(
		query,
		order.ID, // Pass the provided ID
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

	// Insert items for this order
	for _, item := range order.Items {
		query = `
			INSERT INTO order_items (order_id, name, size, color, price, quantity)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = s.DB.Exec(query, orderID, item.Name, item.Size, item.Color, item.Price, item.Quantity)
		if err != nil {
			return 0, err
		}
	}

	return orderID, nil
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
