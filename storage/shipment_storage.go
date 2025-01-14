package storage

import (
	"AAHAOMS/OMS/models"
	"database/sql"
	"encoding/json"

	"fmt"
	"log"
)

func (s *PostgresStorage) HandleShipment(shipment models.Shipment) error {
	// Start a database transaction
	tx, err := s.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Step 1: Validate Order ID
	var orderStatus string
	err = tx.QueryRow(`SELECT order_status FROM orders WHERE id = $1`, shipment.OrderID).Scan(&orderStatus)
	if err != nil {
		log.Printf("Invalid order ID: %v", err)
		return fmt.Errorf("invalid order ID: %v", err)
	}
	log.Printf("Initial order status: %s", orderStatus)

	// Step 2: Fetch Order Items
	orderItemsQuery := `
		SELECT id, name, quantity
		FROM order_items
		WHERE order_id = $1
	`
	rows, err := tx.Query(orderItemsQuery, shipment.OrderID)
	if err != nil {
		log.Printf("Error fetching order items: %v", err)
		return fmt.Errorf("failed to fetch order items: %v", err)
	}
	defer rows.Close()

	orderItems := map[int]models.Item{}
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Quantity); err != nil {
			log.Printf("Error scanning order item: %v", err)
			return fmt.Errorf("failed to scan order item: %v", err)
		}
		orderItems[item.ID] = item
	}

	// Step 3: Handle Shipment Logic
	if shipment.DueOrderType {
		// Ensure the order is already "shipped and due"
		if orderStatus != "shipped and due" {
			log.Printf("Order status mismatch: expected 'shipped and due', got '%s'", orderStatus)
			return fmt.Errorf("order is not in 'shipped and due' status")
		}

		// Fetch due items
		dueItemsQuery := `
			SELECT item_id, quantity
			FROM due_orders
			WHERE order_id = $1
		`
		dueRows, err := tx.Query(dueItemsQuery, shipment.OrderID)
		if err != nil {
			log.Printf("Error fetching due items: %v", err)
			return fmt.Errorf("failed to fetch due items: %v", err)
		}
		defer dueRows.Close()

		dueItems := map[int]int{}
		for dueRows.Next() {
			var itemID, quantity int
			if err := dueRows.Scan(&itemID, &quantity); err != nil {
				log.Printf("Error scanning due item: %v", err)
				return fmt.Errorf("failed to scan due item: %v", err)
			}
			dueItems[itemID] = quantity
		}

		// Validate shipment items match due items exactly
		for _, shippedItem := range shipment.Items {
			dueQuantity, exists := dueItems[shippedItem.ID]
			if !exists || shippedItem.Quantity != dueQuantity {
				log.Printf("Shipment item %d does not match due quantity", shippedItem.ID)
				return fmt.Errorf("shipment item %d does not match due quantity", shippedItem.ID)
			}
		}

		// Remove due items after successful shipment
		_, err = tx.Exec(`
			DELETE FROM due_orders
			WHERE order_id = $1
		`, shipment.OrderID)
		if err != nil {
			log.Printf("Error clearing due orders: %v", err)
			return fmt.Errorf("failed to clear due orders: %v", err)
		}

		// Update order status to "shipped"
		orderStatus = "shipped"
	} else {
		// Validate partial shipment and update due items
		hasDueItems := false
		for _, shippedItem := range shipment.Items {
			orderItem, exists := orderItems[shippedItem.ID]
			if !exists {
				log.Printf("Item ID %d does not exist in the order", shippedItem.ID)
				return fmt.Errorf("item ID %d does not exist in the order", shippedItem.ID)
			}
			if shippedItem.Quantity > orderItem.Quantity {
				log.Printf("Shipped quantity for item %d exceeds order quantity", shippedItem.ID)
				return fmt.Errorf("shipped quantity for item %d exceeds order quantity", shippedItem.ID)
			}

			// Calculate due quantity
			dueQuantity := orderItem.Quantity - shippedItem.Quantity
			if dueQuantity > 0 {
				hasDueItems = true // Mark that there are due items
				_, err = tx.Exec(`
					INSERT INTO due_orders (order_id, item_id, quantity)
					VALUES ($1, $2, $3)
					ON CONFLICT (order_id, item_id)
					DO UPDATE SET quantity = EXCLUDED.quantity
				`, shipment.OrderID, orderItem.ID, dueQuantity)
				if err != nil {
					log.Printf("Error updating due orders: %v", err)
					return fmt.Errorf("failed to update due orders: %v", err)
				}
			}
		}

		// Determine order status based on whether there are due items
		if hasDueItems {
			orderStatus = "shipped and due"
		} else {
			orderStatus = "shipped"
		}
	}
	log.Printf("Final order status: %s", orderStatus)

	// Step 4: Update Order Status
	_, err = tx.Exec(`
		UPDATE orders
		SET order_status = $1
		WHERE id = $2
	`, orderStatus, shipment.OrderID)
	if err != nil {
		log.Printf("Error updating order status: %v", err)
		return fmt.Errorf("failed to update order status: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("Shipment processed successfully for order ID %d", shipment.OrderID)
	return nil
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

func (s *PostgresStorage) GetShipmentByName(customerName string) ([]models.Shipment, error) {

	var orderID int
	err := s.DB.QueryRow(`
		SELECT o.id
		FROM orders o
		INNER JOIN customers c ON o.customer_id = c.id
		WHERE c.name = $1
	`, customerName).Scan(&orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no orders found for customer: %s", customerName)
		}
		return nil, fmt.Errorf("failed to fetch order ID for customer %s: %v", customerName, err)
	}

	// Fetch Shipments for the Order ID
	rows, err := s.DB.Query(`
		SELECT id, shipped_date, order_id, items, due_order_type
		FROM shipments
		WHERE order_id = $1
	`, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shipments for order ID %d: %v", orderID, err)
	}
	defer rows.Close()

	return s.parseShipments(rows)
}

//HELPER FUNCTION

func (s *PostgresStorage) parseShipments(rows *sql.Rows) ([]models.Shipment, error) {
	var shipments []models.Shipment

	for rows.Next() {
		var shipment models.Shipment
		var itemsSQL sql.NullString

		err := rows.Scan(&shipment.ID, &shipment.ShippedDate, &shipment.OrderID, &itemsSQL, &shipment.DueOrderType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shipment row: %v", err)
		}

		// Handle NULL items
		if itemsSQL.Valid {
			err = json.Unmarshal([]byte(itemsSQL.String), &shipment.Items)
			if err != nil {
				return nil, fmt.Errorf("failed to decode shipment items: %v", err)
			}
		} else {
			shipment.Items = []models.Item{} // Default to an empty slice
		}

		shipments = append(shipments, shipment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
	}

	return shipments, nil
}

func (s *PostgresStorage) GetDueItems(orderID int) ([]DueItem, error) {
	var dueItems []DueItem
	query := `
        SELECT item_id, quantity
        FROM due_orders
        WHERE order_id = $1;
    `
	rows, err := s.DB.Query(query, orderID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving due items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var dueItem DueItem
		if err := rows.Scan(&dueItem.ItemID, &dueItem.Quantity); err != nil {
			return nil, fmt.Errorf("error scanning due item: %w", err)
		}
		dueItems = append(dueItems, dueItem)
	}

	return dueItems, nil
}

// DueItem struct
type DueItem struct {
	ItemID   int `json:"item_id"`
	Quantity int `json:"quantity"`
}
