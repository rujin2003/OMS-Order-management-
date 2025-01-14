package models

type Shipment struct {
	ID           int    `json:"id"`
	ShippedDate  string `json:"shipped_date"`
	OrderID      int    `json:"order_id"`
	Items        []Item `json:"items"`
	DueOrderType bool   `json:"due_order_type"`
}
type ShipmentResponse struct {
	ShipmentID   int       `json:"shipment_id"`
	OrderID      int       `json:"order_id"`
	ShipmentDate string    `json:"shipment_date"`
	Status       string    `json:"status"`
	DueItems     []DueItem `json:"due_items"`
}
type DueItem struct {
	ItemID      int    `json:"item_id"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
}
