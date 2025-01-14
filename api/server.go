package api

import (
	"fmt"
	"net/http"

	"AAHAOMS/OMS/storage"

	"github.com/gorilla/mux"
)

type ApiServer struct {
	Address string
	Store   storage.Storage
}

func NewApiServer(address string, store storage.Storage) *ApiServer {
	return &ApiServer{Address: address, Store: store}
}
func (s *ApiServer) Start() {
	router := mux.NewRouter()

	// MARK: Customers
	router.HandleFunc("/customers", makeHandler(wrapHandler(s.handleCustomers))).Methods("POST")
	router.HandleFunc("/customers", makeHandler(wrapHandler(s.getAllCustomers))).Methods("GET")
	router.HandleFunc("/customers/{id:[0-9]+}", makeHandler(wrapHandler(s.getCustomerByID))).Methods("GET")
	router.HandleFunc("/customer/totalCount", makeHandler(wrapHandler(s.getCustumerCount))).Methods("GET")
	router.HandleFunc("/customers/{id}", makeHandler(wrapHandler(s.handleEditCustomers))).Methods("PUT")

	// MARK: Orders
	router.HandleFunc("/orders", makeHandler(wrapHandler(s.handleCreateOrder))).Methods("POST")
	router.HandleFunc("/orders/{id:[0-9]+}", makeHandler(wrapHandler(s.handleGetOrderByID))).Methods("GET")
	router.HandleFunc("/orders", makeHandler(wrapHandler(s.handleGetAllOrders))).Methods("GET")

	router.HandleFunc("/orders/{id:[0-9]+}/status", makeHandler(wrapHandler(s.UpdateOrderStatusHandler))).Methods("POST")
	router.HandleFunc("/orders/{id:[0-9]+}", makeHandler(wrapHandler(s.handleDeleteOrder))).Methods("DELETE")
	router.HandleFunc("/orders/total-value/{customer_name}", makeHandler(wrapHandler(s.handleTotalOrderValueByCustomerName))).Methods("GET")

	router.HandleFunc("/orders/history/{customer_name}", makeHandler(wrapHandler(s.handleGetOrderHistoryByCustomerName))).Methods("GET")
	router.HandleFunc("/orders/pending-count", makeHandler(wrapHandler(s.handlePendingOrderCount))).Methods("GET")
	router.HandleFunc("/orders/count/{customer_name}", makeHandler(wrapHandler(s.handleOrderCountByCustomerName))).Methods("GET")

	// MARK: Shipments
	router.HandleFunc("/shipments", makeHandler(wrapHandler(s.handlePostShipment))).Methods("POST")
	router.HandleFunc("/shipments", makeHandler(wrapHandler(s.handleGetAllShipments))).Methods("GET")
	router.HandleFunc("/shipments/completed", makeHandler(wrapHandler(s.handleGetCompletedShipments))).Methods("GET")
	router.HandleFunc("/shipments/shipped-pending", makeHandler(wrapHandler(s.handleGetShippedButPendingShipments))).Methods("GET")
	router.HandleFunc("/shipments/{id}", makeHandler(wrapHandler(s.handleDeleteShipment))).Methods("DELETE")
	router.HandleFunc("/due_items/{order_id}", makeHandler(wrapHandler(s.handleGetDueItems))).Methods("GET")
	router.HandleFunc("/items/{id}", makeHandler(wrapHandler(s.handleGetItemByID))).Methods("GET")
	router.HandleFunc("/totalSales", makeHandler(wrapHandler(s.handleGetTotalSales))).Methods("GET")

	fmt.Printf("Server starting on %s...\n", s.Address)
	if err := http.ListenAndServe(s.Address, router); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
