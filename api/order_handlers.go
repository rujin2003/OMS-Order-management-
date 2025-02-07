package api

import (
	"AAHAOMS/OMS/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (s *ApiServer) handleTotalOrderCount(w http.ResponseWriter, r *http.Request) {
	totalOrderCount, err := s.Store.TotalOrderCount()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching total order count: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"total_order_count": totalOrderCount})
}
func (s *ApiServer) handlerRecentOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := s.Store.GetRecentOrders(5)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching recent orders: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// CreateOrderHandler handles order creation
func (s *ApiServer) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	var order models.Order

	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	orderID, err := s.Store.CreateOrder(order)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating order: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]int{"order_id": orderID})
}
func (s *ApiServer) handlerDeleteOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	err = s.Store.DeleteOrder(orderID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting order: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Order deleted successfully"})
}

func (s *ApiServer) handleGetTotalSales(w http.ResponseWriter, r *http.Request) {
	totalSales, err := s.Store.GetTotalSalesForShippedOrders()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching total sales: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]float64{"total_sales": totalSales})
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}
func (s *ApiServer) handleGetTotalSalesByCustomer(w http.ResponseWriter, r *http.Request) {
	// Extract the customer name from the route parameters
	vars := mux.Vars(r)
	customerName, ok := vars["customerName"]
	if !ok || customerName == "" {
		http.Error(w, "Customer name is required", http.StatusBadRequest)
		return
	}

	// Fetch total sales for the customer
	totalSales, err := s.Store.GetTotalSalesForShippedOrdersByCustomer(customerName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching total sales for customer %s: %v", customerName, err), http.StatusInternalServerError)
		return
	}

	// Return the result as JSON
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]float64{"total_sales": totalSales})
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func (s *ApiServer) handleGetOrderByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	order, err := s.Store.GetOrderByID(orderID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching order: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(order)
}

func (s *ApiServer) handleGetOrderHistoryByCustomerName(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	customerName, ok := vars["customer_name"]
	if !ok || customerName == "" {
		http.Error(w, "Customer name is required", http.StatusBadRequest)
		return
	}

	// Fetch the order history from the store
	orders, err := s.Store.GetOrderHistoryByCustomerName(customerName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching order history: %v", err), http.StatusInternalServerError)
		return
	}

	// Encode the orders as JSON and send as response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (s *ApiServer) handleGetLatestOrderID(w http.ResponseWriter, r *http.Request) {
	latestOrderID, err := s.Store.GetLatestOrderID()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching latest order ID: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]int{"latest_order_id": latestOrderID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *ApiServer) handleGetAllOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := s.Store.GetAllOrders()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching orders: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(orders)
}

func (s *ApiServer) UpdateOrderStatusHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	orderID, err := strconv.Atoi(vars["id"])
	if err != nil || orderID <= 0 {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var payload struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Status == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = s.Store.UpdateOrderStatus(orderID, payload.Status)
	if err != nil {
		http.Error(w, "Failed to update order status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Order status updated successfully",
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
	log.Println(pendingCount)
	if err != nil {
		http.Error(w, "Failed to fetch pending order count", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"pending_order_count": pendingCount,
	}
	json.NewEncoder(w).Encode(response)
}

func (s *ApiServer) handleOrderByDateAndName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerName := vars["customer_name"]
	orderDate := vars["order_date"]

	orders, err := s.Store.GetOrdersByNameAndDate(customerName, orderDate)
	if err != nil {
		http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(orders)
}
