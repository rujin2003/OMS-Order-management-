package storage

import (
	"AAHAOMS/OMS/models"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
)



// Handle a new shipment transactionally
func (s *PostgresStorage) HandleShipment(shipment models.Shipment) error {
	tx, err := s.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Validate order existence and get status
	orderStatus, err := s.getOrderStatus(tx, shipment.OrderID)
	if err != nil {
		return err
	}

	// Fetch order items
	orderItems, err := s.getOrderItems(tx, shipment.OrderID)
	if err != nil {
		return err
	}

	// Handle due order shipment
	if shipment.DueOrderType {
		err := s.handleDueOrderShipment(tx, shipment, orderStatus)
		if err != nil {
			return err
		}
		orderStatus = "shipped"
	} else {
		orderStatus, err = s.handlePartialShipment(tx, shipment, orderItems)
		if err != nil {
			return err
		}
	}

	// Update order status
	if err := s.updateOrderStatus(tx, shipment.OrderID, orderStatus); err != nil {
		return err
	}

	// Insert shipment record
	if err := s.insertShipment(tx, shipment); err != nil {
		return err
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("Shipment processed successfully for order ID %d", shipment.OrderID)
	return nil
}

// Handle due order shipment
func (s *PostgresStorage) handleDueOrderShipment(tx *sql.Tx, shipment models.Shipment, orderStatus string) error {
	if orderStatus != "shipped and due" {
		log.Printf("Order status mismatch: expected 'shipped and due', got '%s'", orderStatus)
		return fmt.Errorf("order is not in 'shipped and due' status")
	}

	rows, err := tx.Query(`SELECT item_id, quantity FROM due_orders WHERE order_id = $1`, shipment.OrderID)
	if err != nil {
		return fmt.Errorf("failed to fetch due items: %v", err)
	}
	defer rows.Close()

	dueItems := make(map[int]int)
	for rows.Next() {
		var itemID, quantity int
		if err := rows.Scan(&itemID, &quantity); err != nil {
			return fmt.Errorf("failed to scan due item: %v", err)
		}
		dueItems[itemID] = quantity
	}

	for _, shippedItem := range shipment.Items {
		if dueQuantity, exists := dueItems[shippedItem.ID]; !exists || shippedItem.Quantity != dueQuantity {
			return fmt.Errorf("shipment item %d does not match due quantity", shippedItem.ID)
		}
	}

	_, err = tx.Exec(`DELETE FROM due_orders WHERE order_id = $1`, shipment.OrderID)
	if err != nil {
		return fmt.Errorf("failed to clear due orders: %v", err)
	}
	return nil
}

// Handle partial shipment
func (s *PostgresStorage) handlePartialShipment(tx *sql.Tx, shipment models.Shipment, orderItems map[int]models.Item) (string, error) {
	hasDueItems := false

	for _, shippedItem := range shipment.Items {
		orderItem, exists := orderItems[shippedItem.ID]
		if !exists {
			return "", fmt.Errorf("item ID %d does not exist in the order", shippedItem.ID)
		}
		if shippedItem.Quantity > orderItem.Quantity {
			return "", fmt.Errorf("shipped quantity for item %d exceeds order quantity", shippedItem.ID)
		}

		dueQuantity := orderItem.Quantity - shippedItem.Quantity
		if dueQuantity > 0 {
			hasDueItems = true
			_, err := tx.Exec(`
				INSERT INTO due_orders (order_id, item_id, quantity)
				VALUES ($1, $2, $3)
				ON CONFLICT (order_id, item_id) DO UPDATE SET quantity = EXCLUDED.quantity
			`, shipment.OrderID, orderItem.ID, dueQuantity)
			if err != nil {
				return "", fmt.Errorf("failed to update due orders: %v", err)
			}
		}
	}

	if hasDueItems {
		return "shipped and due", nil
	}
	return "shipped", nil
}

// Update order status
func (s *PostgresStorage) updateOrderStatus(tx *sql.Tx, orderID int, status string) error {
	_, err := tx.Exec(`UPDATE orders SET order_status = $1 WHERE id = $2`, status, orderID)
	return err
}

// Insert shipment into the database
func (s *PostgresStorage) insertShipment(tx *sql.Tx, shipment models.Shipment) error {
	var itemIDs []int
	for _, item := range shipment.Items {
		itemIDs = append(itemIDs, item.ID)
	}

	shippedDate, err := time.Parse("2006-01-02", shipment.ShippedDate)
	if err != nil {
		return fmt.Errorf("invalid date format: %v", err)
	}

	_, err = tx.Exec(`
		INSERT INTO shipments (order_id, shipped_date, items, due_order_type)
		VALUES ($1, $2, $3::int[], $4)
	`, shipment.OrderID, shippedDate, pq.Array(itemIDs), shipment.DueOrderType)

	return err
}

// Retrieve all shipments
func (s *PostgresStorage) GetAllShipments() ([]models.Shipment, error) {
	rows, err := s.DB.Query(`
		SELECT s.id, s.order_id, s.shipped_date::DATE, s.items::int[], s.due_order_type
		FROM shipments s
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.parseShipments(rows)
}

func (s *PostgresStorage) GetShipmentByName(customerName string) ([]models.Shipment, error) {
	var orderIDs []int

	// Fetch all orders associated with the given customer name
	orderQuery := `
		SELECT o.id 
		FROM orders o
		INNER JOIN customers c ON o.customer_id = c.id
		WHERE c.name = $1
	`

	rows, err := s.DB.Query(orderQuery, customerName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order IDs for customer %s: %v", customerName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var orderID int
		if err := rows.Scan(&orderID); err != nil {
			return nil, fmt.Errorf("error scanning order ID: %v", err)
		}
		orderIDs = append(orderIDs, orderID)
	}

	// If no orders are found, return an empty array
	if len(orderIDs) == 0 {
		return []models.Shipment{}, nil
	}

	// Fetch shipments for the retrieved order IDs
	shipmentsQuery := `
		SELECT s.id, s.order_id, s.shipped_date::DATE, s.items::int[], s.due_order_type
		FROM shipments s
		WHERE s.order_id = ANY($1)
	`

	rows, err = s.DB.Query(shipmentsQuery, pq.Array(orderIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shipments: %v", err)
	}
	defer rows.Close()

	return s.parseShipments(rows)
}


// Parse shipments from SQL rows
func (s *PostgresStorage) parseShipments(rows *sql.Rows) ([]models.Shipment, error) {
	var shipments []models.Shipment

	for rows.Next() {
		var shipment models.Shipment
		var shippedDate string
		var itemIDs pq.Int64Array

		err := rows.Scan(&shipment.ID, &shipment.OrderID, &shippedDate, &itemIDs, &shipment.DueOrderType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shipment row: %v", err)
		}

		shipment.ShippedDate = shippedDate

		// Fetch full item details
		shipment.Items, err = s.getItemDetails(itemIDs)
		if err != nil {
			return nil, err
		}

		shipments = append(shipments, shipment)
	}

	return shipments, nil
}

// Fetch full item details
func (s *PostgresStorage) getItemDetails(itemIDs pq.Int64Array) ([]models.Item, error) {
	if len(itemIDs) == 0 {
		return []models.Item{}, nil
	}

	query := `
		SELECT id, name, price, quantity 
		FROM order_items 
		WHERE id = ANY($1)
	`
	rows, err := s.DB.Query(query, pq.Array(itemIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		err := rows.Scan(&item.ID, &item.Name, &item.Price, &item.Quantity)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
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
