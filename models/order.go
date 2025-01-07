package models

type Order struct {
	ID          int    `json:"id"`
	CustomerID  int    `json:"customer_id"`
	OrderDate   string `json:"order_date"`
	ShipmentDue string `json:"shipment_due"`
	Items       []Item `json:"items"`
}

type Item struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Size    *string `json:"size,omitempty"`
	Color   *string `json:"color,omitempty"`
	Price   float64 `json:"price"`
	Shipped bool    `json:"shipped"`
}
