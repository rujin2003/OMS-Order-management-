package models

type Order struct {
	ID              int    `json:"id"`
	CustomerID      int    `json:"customer_id"`
	CustomerName    string `json:"customer_name"`
	OrderDate       string `json:"order_date"`
	ShipmentDue     string `json:"shipment_due"`
	ShipmentAddress string `json:"shipment_address"`
	// OrderStatus can be one of: "pending", "shipped", "shipped but due"
	OrderStatus string  `json:"order_status"`
	Items       []Item  `json:"items"`
	TotalPrice  float64 `json:"total_price"`
	NoOfItems   int     `json:"no_of_items"`
}

type Item struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Size     *string `json:"size,omitempty"`
	Color    *string `json:"color,omitempty"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}
