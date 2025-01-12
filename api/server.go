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

	// MARK:Customers

	router.HandleFunc("/customers", s.handleCustomers).Methods("POST")
	router.HandleFunc("/customers", s.getAllCustomers).Methods("GET")
	router.HandleFunc("/customers/{id:[0-9]+}", s.getCustomerByID).Methods("GET")
	router.HandleFunc("/customer/totalCount", s.getCustumerCount).Methods("GET")
	router.HandleFunc("/customers/{id}", s.handleEditCustomers).Methods("PUT")

	//MARK:Order

	router.HandleFunc("/orders", s.handleCreateOrder).Methods("POST")
	router.HandleFunc("/orders/{id:[0-9]+}", s.handleGetOrderByID).Methods("GET")
	router.HandleFunc("/orders", s.handleGetAllOrders).Methods("GET")

	router.HandleFunc("/orders/{id:[0-9]+}/status", s.UpdateOrderStatusHandler).Methods("POST")

	router.HandleFunc("/orders/{id:[0-9]+}", s.handleDeleteOrder).Methods("DELETE")

	router.HandleFunc("/orders/total-value/{customer_name}", s.handleTotalOrderValueByCustomerName).Methods("GET")
	router.HandleFunc("/orders/history/{customer_name}", s.handleGetOrderHistoryByCustomerName).Methods("GET")
	router.HandleFunc("/orders/pending-count", s.handlePendingOrderCount).Methods("GET")
	router.HandleFunc("/orders/count/{customer_name}", s.handleOrderCountByCustomerName).Methods("GET")

	//MARK:Shipmemnt
	router.HandleFunc("/shipments", s.handlePostShipment).Methods("POST")
	router.HandleFunc("/shipments", s.handleGetAllShipments).Methods("GET")
	router.HandleFunc("/shipments/completed", s.handleGetCompletedShipments).Methods("GET")
	router.HandleFunc("/shipments/shipped-pending", s.handleGetShippedButPendingShipments).Methods("GET")
	router.HandleFunc("/shipments/{id}", s.handleDeleteShipment).Methods("DELETE")

	fmt.Printf("Server starting on %s...\n", s.Address)
	if err := http.ListenAndServe(s.Address, router); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
