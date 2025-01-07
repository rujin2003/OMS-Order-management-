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

	//MARK:Shipment

	router.HandleFunc("/orders", s.handleOrders).Methods("POST")
	router.HandleFunc("/shipments", s.handleShipments).Methods("POST")
	router.HandleFunc("/order-history/{name}", s.handleOrderHistoryByName).Methods("GET")
	router.HandleFunc("/shipment-history/{name}", s.handleShipmentHistoryByName).Methods("GET")

	fmt.Printf("Server starting on %s...\n", s.Address)
	if err := http.ListenAndServe(s.Address, router); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
