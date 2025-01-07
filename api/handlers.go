package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

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
