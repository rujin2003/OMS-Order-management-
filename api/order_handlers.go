package api

import (
	"AAHAOMS/OMS/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (s *ApiServer) handleOrders(w http.ResponseWriter, r *http.Request) {
	var order models.Order

	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	orderID, err := s.Store.CreateOrder(order.CustomerID, order.CustomerName, order.OrderDate, order.ShipmentDue, order.ShipmentAddress)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating order: %v", err), http.StatusInternalServerError)
		return
	}

	totalPrice := 0.0
	for _, item := range order.Items {
		itemTotalPrice := item.Price * float64(item.Quantity)
		totalPrice += itemTotalPrice

		if err := s.Store.AddItemToOrder(orderID, item.Name, item.Size, item.Color, item.Price, item.Quantity); err != nil {
			http.Error(w, fmt.Sprintf("Error adding item: %v", err), http.StatusInternalServerError)
			return
		}
	}

	response := map[string]interface{}{
		"order_id":         orderID,
		"customer_id":      order.CustomerID,
		"customer_name":    order.CustomerName,
		"order_date":       order.OrderDate,
		"shipment_due":     order.ShipmentDue,
		"shipment_address": order.ShipmentAddress,
		"no_of_items":      len(order.Items),
		"total_price":      totalPrice,
	}

	json.NewEncoder(w).Encode(response)
}

func (s *ApiServer) handleGetOrderByID(w http.ResponseWriter, r *http.Request) {
	orderID := mux.Vars(r)["id"]
	id, err := strconv.Atoi(orderID)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	order, err := s.Store.GetOrderByID(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving order: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(order)
}

func (s *ApiServer) handleGetAllOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := s.Store.GetAllOrders()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving orders: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(orders)
}
