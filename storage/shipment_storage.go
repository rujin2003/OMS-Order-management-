package storage

import (
	"AAHAOMS/OMS/models"
	"database/sql"
	"encoding/json"

	"fmt"
)

func (s *PostgresStorage) HandleShipment(shipment models.Shipment) error {
	// Start a database transaction
	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Step 1: Validate Order ID and Fetch Order Status
	var orderStatus string
	err = tx.QueryRow(`SELECT order_status FROM orders WHERE id = $1`, shipment.OrderID).Scan(&orderStatus)
	if err != nil {
		return fmt.Errorf("invalid order ID: %v", err)
	}

	// Step 2: Fetch Original Order Items
	orderItemsQuery := `
		SELECT id, name, size, color, price, quantity
		FROM order_items
		WHERE order_id = $1
	`
	rows, err := tx.Query(orderItemsQuery, shipment.OrderID)
	if err != nil {
		return fmt.Errorf("failed to fetch order items: %v", err)
	}
	defer rows.Close()

	originalItems := map[int]models.Item{}
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Size, &item.Color, &item.Price, &item.Quantity); err != nil {
			return fmt.Errorf("failed to scan order item: %v", err)
		}
		originalItems[item.ID] = item
	}

	// Step 3: Handle Shipment Based on `dueOrderType`
	if shipment.DueOrderType {
		// Fetch due items
		dueItems := map[int]int{}
		dueRows, err := tx.Query(`
			SELECT item_id, quantity
			FROM due_orders
			WHERE order_id = $1
		`, shipment.OrderID)
		if err != nil {
			return fmt.Errorf("failed to fetch due items: %v", err)
		}
		defer dueRows.Close()

		for dueRows.Next() {
			var itemID, quantity int
			if err := dueRows.Scan(&itemID, &quantity); err != nil {
				return fmt.Errorf("failed to scan due item: %v", err)
			}
			dueItems[itemID] = quantity
		}

		// Validate shipment items
		for _, shippedItemID := range shipment.Items {
			if dueItems[shippedItemID] == 0 {
				return fmt.Errorf("item ID %d is not due for shipment", shippedItemID)
			}
		}
	} else {
		// Validate partial shipment
		for _, shippedItemID := range shipment.Items {
			originalItem, exists := originalItems[shippedItemID]
			if !exists {
				return fmt.Errorf("item ID %d does not exist in the order", shippedItemID)
			}

			// Update due items
			if originalItem.Quantity > 1 {
				_, err = tx.Exec(`
					INSERT INTO due_orders (order_id, item_id, quantity)
					VALUES ($1, $2, $3)
					ON CONFLICT (order_id, item_id)
					DO UPDATE SET quantity = due_orders.quantity + EXCLUDED.quantity
				`, shipment.OrderID, originalItem.ID, originalItem.Quantity-1)
				if err != nil {
					return fmt.Errorf("failed to update due orders: %v", err)
				}
			}
		}
	}

	// Step 4: Insert Shipment Record
	_, err = tx.Exec(`
		INSERT INTO shipments (order_id, shipped_date, items)
		VALUES ($1, $2, $3)
	`, shipment.OrderID, shipment.ShippedDate, shipment.Items)
	if err != nil {
		return fmt.Errorf("failed to insert shipment: %v", err)
	}

	// Step 5: Update Order Status
	newStatus := "shipped"
	if len(shipment.Items) < len(originalItems) {
		newStatus = "shipped and due"
	}
	_, err = tx.Exec(`
		UPDATE orders
		SET order_status = $1
		WHERE id = $2
	`, newStatus, shipment.OrderID)
	if err != nil {
		return fmt.Errorf("failed to update order status: %v", err)
	}

	// Commit transaction
	return tx.Commit()
}

func (s *PostgresStorage) GetAllShipments() ([]models.Shipment, error) {
	query := `
		SELECT id, shipped_date, order_id, items, due_order_type
		FROM shipments
	`
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve all shipments: %v", err)
	}
	defer rows.Close()

	return s.parseShipments(rows)
}

func (s *PostgresStorage) GetCompletedShipments() ([]models.Shipment, error) {
	query := `
		SELECT s.id, s.shipped_date, s.order_id, s.items, s.due_order_type
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

func (s *PostgresStorage) GetShippedButPendingShipments() ([]models.Shipment, error) {
	query := `
		SELECT s.id, s.shipped_date, s.order_id, s.items, s.due_order_type
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

func (s *PostgresStorage) DeleteShipment(shipmentID int) error {
	// Step 1: Fetch the Order ID related to the Shipment
	var orderID int
	getOrderIDQuery := `SELECT order_id FROM shipments WHERE id = $1`
	err := s.DB.QueryRow(getOrderIDQuery, shipmentID).Scan(&orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("shipment ID %d not found", shipmentID)
		}
		return fmt.Errorf("failed to fetch order ID for shipment ID %d: %v", shipmentID, err)
	}

	// Step 2: Delete the Shipment
	deleteShipmentQuery := `DELETE FROM shipments WHERE id = $1`
	_, err = s.DB.Exec(deleteShipmentQuery, shipmentID)
	if err != nil {
		return fmt.Errorf("failed to delete shipment ID %d: %v", shipmentID, err)
	}

	// Step 3: Revert the Order Status to "pending"
	updateOrderStatusQuery := `UPDATE orders SET order_status = 'pending' WHERE id = $1`
	_, err = s.DB.Exec(updateOrderStatusQuery, orderID)
	if err != nil {
		return fmt.Errorf("failed to update order status for order ID %d: %v", orderID, err)
	}

	return nil
}

//HELPER FUNCTION

func (s *PostgresStorage) parseShipments(rows *sql.Rows) ([]models.Shipment, error) {
	shipments := []models.Shipment{}

	for rows.Next() {
		var shipment models.Shipment
		var itemsJSON []byte

		err := rows.Scan(&shipment.ID, &shipment.ShippedDate, &shipment.OrderID, &itemsJSON, &shipment.DueOrderType)
		if err != nil {
			return nil, fmt.Errorf("failed to parse shipment row: %v", err)
		}

		// Decode JSON-encoded items
		if err := json.Unmarshal(itemsJSON, &shipment.Items); err != nil {
			return nil, fmt.Errorf("failed to decode shipment items: %v", err)
		}

		shipments = append(shipments, shipment)
	}

	return shipments, nil
}
