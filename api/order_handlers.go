package api

import (
	"AAHAOMS/OMS/models"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// CreateOrderHandler handles order creation
func (s *ApiServer) handleOrders(w http.ResponseWriter, r *http.Request) {
	var order models.Order
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	orderID, err := s.Store.CreateOrder(order.CustomerID, order.CustomerName, order.OrderDate, order.ShipmentDue, order.ShipmentAddress, "pending")
	if err != nil {
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"order_id":         orderID,
		"customer_id":      order.CustomerID,
		"customer_name":    order.CustomerName,
		"order_date":       order.OrderDate,
		"shipment_due":     order.ShipmentDue,
		"shipment_address": order.ShipmentAddress,
		"order_status":     "Pending",
	}
	json.NewEncoder(w).Encode(response)
}

func (s *ApiServer) handleGetOrderByID(w http.ResponseWriter, r *http.Request) {
	orderIDStr := r.URL.Query().Get("id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	order, err := s.Store.GetOrderByID(orderID)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(order)
}

func (s *ApiServer) handleGetAllOrders(w http.ResponseWriter, r *http.Request) {
	customerName := r.URL.Query().Get("customer_name")
	if customerName == "" {
		http.Error(w, "Customer name is required", http.StatusBadRequest)
		return
	}

	orders, err := s.Store.GetOrderHistoryByCustomerName(customerName)
	if err != nil {
		http.Error(w, "Failed to fetch order history", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(orders)
}

// GetAllOrdersHandler fetches all orders
func (s *ApiServer) handleGetOrderHistoryByCustomerName(w http.ResponseWriter, r *http.Request) {
	orders, err := s.Store.GetAllOrders()
	if err != nil {
		http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(orders)
}

func (s *ApiServer) UpdateOrderStatusHandler(w http.ResponseWriter, r *http.Request) {

	var payload struct {
		OrderID int    `json:"order_id"`
		Status  string `json:"status"`
	}

	// Decode the JSON request body
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil || payload.OrderID <= 0 || payload.Status == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Update the order status in the database
	err = s.Store.UpdateOrderStatus(payload.OrderID, payload.Status)
	if err != nil {
		http.Error(w, "Failed to update order status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Order status updated successfully",
	})
}

func (s *ApiServer) handleDeleteOrders(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		OrderID int `json:"order_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil || payload.OrderID <= 0 {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = s.Store.DeleteOrder(payload.OrderID)
	if err != nil {
		http.Error(w, "Failed to delete order", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Order deleted successfully",
	})
}

func (s *ApiServer) handleTotalOrderValueByCustomerName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerName := vars["customer_name"]

	totalValue, err := s.Store.GetTotalOrderValueByCustomerName(customerName)
	if err != nil {
		http.Error(w, "Failed to calculate total order value", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"customer_name": customerName,
		"total_value":   totalValue,
	}
	json.NewEncoder(w).Encode(response)
}

func (s *ApiServer) handleOrderCountByCustomerName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerName := vars["customer_name"]

	if customerName == "" {
		http.Error(w, "Customer name is required", http.StatusBadRequest)
		return
	}

	orderCount, err := s.Store.GetOrderCountByCustomerName(customerName)
	if err != nil {
		http.Error(w, "Failed to fetch order count", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"customer_name": customerName,
		"order_count":   orderCount,
	}
	json.NewEncoder(w).Encode(response)
}

func (s *ApiServer) handlePendingOrderCount(w http.ResponseWriter, r *http.Request) {
	pendingCount, err := s.Store.GetPendingOrderCount()
	if err != nil {
		http.Error(w, "Failed to fetch pending order count", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"pending_order_count": pendingCount,
	}
	json.NewEncoder(w).Encode(response)
}
