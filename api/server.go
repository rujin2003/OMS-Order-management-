// package api

// import (
// 	"fmt"
// 	"net/http"

// 	"AAHAOMS/OMS/storage"

// 	"github.com/gorilla/mux"
// )

// type ApiServer struct {
// 	Address string
// 	Store   storage.Storage
// }

// func NewApiServer(address string, store storage.Storage) *ApiServer {
// 	return &ApiServer{Address: address, Store: store}
// }
// func (s *ApiServer) Start() {
// 	router := mux.NewRouter()

// 	// MARK: Customers
// 	router.HandleFunc("/customers", makeHandler(wrapHandler(s.handleCustomers))).Methods("POST")
// 	router.HandleFunc("/customers", makeHandler(wrapHandler(s.getAllCustomers))).Methods("GET")
// 	router.HandleFunc("/customers/{id:[0-9]+}", makeHandler(wrapHandler(s.getCustomerByID))).Methods("GET")
// 	router.HandleFunc("/customer/totalCount", makeHandler(wrapHandler(s.getCustumerCount))).Methods("GET")
// 	router.HandleFunc("/customers/{id}", makeHandler(wrapHandler(s.handleEditCustomers))).Methods("PUT")

// 	// MARK: Orders
// 	router.HandleFunc("/orders", makeHandler(wrapHandler(s.handleCreateOrder))).Methods("POST")
// 	router.HandleFunc("/orders/{id:[0-9]+}", makeHandler(wrapHandler(s.handleGetOrderByID))).Methods("GET")
// 	router.HandleFunc("/orders", makeHandler(wrapHandler(s.handleGetAllOrders))).Methods("GET")

// 	router.HandleFunc("/orders/{id:[0-9]+}/status", makeHandler(wrapHandler(s.UpdateOrderStatusHandler))).Methods("POST")
// 	router.HandleFunc("/orders/{id:[0-9]+}", makeHandler(wrapHandler(s.handleDeleteOrder))).Methods("DELETE")
// 	router.HandleFunc("/orders/total-value/{customer_name}", makeHandler(wrapHandler(s.handleTotalOrderValueByCustomerName))).Methods("GET")

// 	router.HandleFunc("/orders/history/{customer_name}", makeHandler(wrapHandler(s.handleGetOrderHistoryByCustomerName))).Methods("GET")
// 	router.HandleFunc("/orders/pending-count", makeHandler(wrapHandler(s.handlePendingOrderCount))).Methods("GET")
// 	router.HandleFunc("/orders/count/{customer_name}", makeHandler(wrapHandler(s.handleOrderCountByCustomerName))).Methods("GET")

// 	// MARK: Shipments
// 	router.HandleFunc("/shipments", makeHandler(wrapHandler(s.handlePostShipment))).Methods("POST")
// 	router.HandleFunc("/shipments", makeHandler(wrapHandler(s.handleGetAllShipments))).Methods("GET")
// 	router.HandleFunc("/shipments/completed", makeHandler(wrapHandler(s.handleGetCompletedShipments))).Methods("GET")
// 	router.HandleFunc("/shipments/shipped-pending", makeHandler(wrapHandler(s.handleGetShippedButPendingShipments))).Methods("GET")
// 	router.HandleFunc("/shipments/{id}", makeHandler(wrapHandler(s.handleDeleteShipment))).Methods("DELETE")
// 	router.HandleFunc("/due_items/{order_id}", makeHandler(wrapHandler(s.handleGetDueItems))).Methods("GET")
// 	router.HandleFunc("/items/{id}", makeHandler(wrapHandler(s.handleGetItemByID))).Methods("GET")
// 	router.HandleFunc("/totalSales", makeHandler(wrapHandler(s.handleGetTotalSales))).Methods("GET")

//		fmt.Printf("Server starting on %s...\n", s.Address)
//		if err := http.ListenAndServe(s.Address, router); err != nil {
//			fmt.Printf("Error starting server: %v\n", err)
//		}
//	}
package api

//MARK: TODO: Delete Customer function

import (
	"fmt"
	"net/http"

	"AAHAOMS/OMS/storage"

	"github.com/gorilla/mux"
)

// ApiServer struct
type ApiServer struct {
	Address string
	Store   storage.Storage
}

// NewApiServer creates a new server instance
func NewApiServer(address string, store storage.Storage) *ApiServer {
	return &ApiServer{Address: address, Store: store}
}

// CORS Middleware
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from React frontend on port 8082
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight OPTIONS request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Start initializes the server
func (s *ApiServer) Start() {
	router := mux.NewRouter()

	// MARK: Customers
	router.HandleFunc("/customers", makeHandler(wrapHandler(s.handleCustomers))).Methods("POST")
	router.HandleFunc("/customers", makeHandler(wrapHandler(s.getAllCustomers))).Methods("GET")
	router.HandleFunc("/customers/{id:[0-9]+}", makeHandler(wrapHandler(s.getCustomerByID))).Methods("GET")
	router.HandleFunc("/customer/totalCount", makeHandler(wrapHandler(s.getCustumerCount))).Methods("GET")
	router.HandleFunc("/customers/{id}", makeHandler(wrapHandler(s.handleEditCustomers))).Methods("PUT")
	router.HandleFunc("/customers/{id}", makeHandler(wrapHandler(s.handleDeleteCustomer))).Methods("DELETE")

	// MARK: Orders
	

	router.HandleFunc("/orders", makeHandler(wrapHandler(s.handleCreateOrder))).Methods("POST")
	router.HandleFunc("/orders/{id:[0-9]+}", makeHandler(wrapHandler(s.handleGetOrderByID))).Methods("GET")
	router.HandleFunc("/orders", makeHandler(wrapHandler(s.handleGetAllOrders))).Methods("GET")
	router.HandleFunc("/orders/{id:[0-9]+}/status", makeHandler(wrapHandler(s.UpdateOrderStatusHandler))).Methods("POST")
	router.HandleFunc("/orders/{id:[0-9]+}", makeHandler(wrapHandler(s.handlerDeleteOrder))).Methods("DELETE")
	router.HandleFunc("/orders/total-value/{customer_name}", makeHandler(wrapHandler(s.handleTotalOrderValueByCustomerName))).Methods("GET")
	router.HandleFunc("/order/totalordercount", makeHandler(wrapHandler(s.handleTotalOrderCount))).Methods("GET")
	router.HandleFunc("/orders/recentorders", makeHandler(wrapHandler(s.handlerRecentOrders))).Methods("GET")

	router.HandleFunc("/orders/history/{customer_name}", makeHandler(wrapHandler(s.handleGetOrderHistoryByCustomerName))).Methods("GET")
	router.HandleFunc("/orders/pending-count", makeHandler(wrapHandler(s.handlePendingOrderCount))).Methods("GET")
	router.HandleFunc("/orders/count/{customer_name}", makeHandler(wrapHandler(s.handleOrderCountByCustomerName))).Methods("GET")
	router.HandleFunc("/orders/latestOrderId", makeHandler(wrapHandler(s.handleGetLatestOrderID))).Methods("GET")
	router.HandleFunc("/orders/{customer_name}/{order_date}", makeHandler(wrapHandler(s.handleOrderByDateAndName))).Methods("GET")
	router.HandleFunc("/due_items/{order_id}", makeHandler(wrapHandler(s.handleGetDueItems))).Methods("GET")

	// MARK: Shipments
	router.HandleFunc("/shipments", makeHandler(wrapHandler(s.handlePostShipment))).Methods("POST")
	router.HandleFunc("/shipments", makeHandler(wrapHandler(s.handleGetAllShipments))).Methods("GET")
	router.HandleFunc("/shipments/completed", makeHandler(wrapHandler(s.handleGetCompletedShipments))).Methods("GET")

	router.HandleFunc("/shipments/shipped-pending", makeHandler(wrapHandler(s.handleGetShippedButPendingShipments))).Methods("GET")
	router.Handle("/shipments/{customer_name}", makeHandler(wrapHandler(s.handleGetShipmentHistoryByCustomerName))).Methods("GET")

	router.HandleFunc("/shipments/{id}", makeHandler(wrapHandler(s.handleDeleteShipment))).Methods("DELETE")

	router.HandleFunc("/items/{id}", makeHandler(wrapHandler(s.handleGetItemByID))).Methods("GET")
	router.HandleFunc("/totalSales", makeHandler(wrapHandler(s.handleGetTotalSales))).Methods("GET")
	router.HandleFunc("/totalSales/{customerName}", makeHandler(wrapHandler(s.handleGetTotalSalesByCustomer))).Methods("GET")

	// Apply CORS middleware to all routes
	corsRouter := enableCORS(router)

	fmt.Printf("Server starting on %s...\n", s.Address)
	if err := http.ListenAndServe(s.Address, corsRouter); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
