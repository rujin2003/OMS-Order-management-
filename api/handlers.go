package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *ApiServer) handleOrders(w http.ResponseWriter, r *http.Request) {
	var order struct {
		CustomerID int    `json:"customer_id"`
		OrderDate  string `json:"order_date"`
		DueDate    string `json:"shipment_due"`
		Items      []struct {
			Name  string  `json:"name"`
			Size  *string `json:"size"`
			Color *string `json:"color"`
			Price float64 `json:"price"`
		} `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	orderID, err := s.Store.CreateOrder(order.CustomerID, order.OrderDate, order.DueDate)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating order: %v", err), http.StatusInternalServerError)
		return
	}
	for _, item := range order.Items {
		if err := s.Store.AddItemToOrder(orderID, item.Name, item.Size, item.Color, item.Price); err != nil {
			http.Error(w, fmt.Sprintf("Error adding item: %v", err), http.StatusInternalServerError)
			return
		}
	}
	json.NewEncoder(w).Encode(map[string]int{"order_id": orderID})
}

func (s *ApiServer) handleShipments(w http.ResponseWriter, r *http.Request) {
	var shipment struct {
		OrderID int   `json:"order_id"`
		Items   []int `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&shipment); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if err := s.Store.CreateShipment(shipment.OrderID, shipment.Items); err != nil {
		http.Error(w, fmt.Sprintf("Error creating shipment: %v", err), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "shipment created"})
}

func (s *ApiServer) handleOrderHistoryByName(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	history, err := s.Store.GetOrderHistoryByName(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving order history: %v", err), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(history)
}

func (s *ApiServer) handleShipmentHistoryByName(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	history, err := s.Store.GetShipmentHistoryByName(name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving shipment history: %v", err), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(history)
}
