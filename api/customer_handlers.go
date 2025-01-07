package api

import (
	"AAHAOMS/OMS/models"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *ApiServer) handleCustomers(w http.ResponseWriter, r *http.Request) {
	var customer models.Customer

	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	id, err := s.Store.CreateCustomer(customer.Name, customer.Number, customer.Email, customer.Country, customer.Address)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating customer: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]int{"customer_id": id})
}

func (s *ApiServer) getCustomerByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	customer, err := s.Store.GetCustomerByID(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching customer: %v", err), http.StatusInternalServerError)
		return
	}
	if customer == nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(customer)
}

func (s *ApiServer) getAllCustomers(w http.ResponseWriter, r *http.Request) {
	customers, err := s.Store.GetAllCustomers()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching customers: %v", err), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(customers)
}
