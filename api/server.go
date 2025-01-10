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

	//MARK:Order

	router.HandleFunc("/orders", s.handleOrders).Methods("POST")
	router.HandleFunc("/orders/{id}", s.handleGetOrderByID).Methods("GET")
	router.HandleFunc("/orders", s.handleGetAllOrders).Methods("GET")
	router.HandleFunc("/orders/history/{name}", s.handleGetOrderHistoryByCustomerName).Methods("GET")
	router.HandleFunc("update_order_status", s.UpdateOrderStatusHandler).Methods("POST")
	router.HandleFunc("/orders", s.handleDeleteOrders).Methods("DELETE")
	router.HandleFunc("/orders/total-value/{customer_name}", s.handleOrderCountByCustomerName).Methods("GET")
	router.HandleFunc("/orders/history/{customer_name}", s.handleGetOrderHistoryByCustomerName).Methods("GET")
	router.HandleFunc("/orders/{pending-count}", s.handlePendingOrderCount).Methods("GET")

	//MARK:Shipmemnt

	router.HandleFunc("/shipments", s.handleShipments).Methods("POST")
	router.HandleFunc("/shipment-history/{name}", s.handleShipmentHistoryByName).Methods("GET")

	fmt.Printf("Server starting on %s...\n", s.Address)
	if err := http.ListenAndServe(s.Address, router); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
