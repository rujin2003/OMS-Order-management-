package models

type Shipment struct {
	ID          int    `json:"id"`
	ShippedDate string `json:"shipped_date"`
	OrderID     int    `json:"order_id"`
	Items       []int  `json:"items"`
}
