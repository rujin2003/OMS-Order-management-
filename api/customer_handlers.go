package api

import (
	"AAHAOMS/OMS/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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

func (s *ApiServer) handleEditCustomers(w http.ResponseWriter, r *http.Request) {
	var customer models.Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}
	customer.ID = id
	err = s.Store.EditCustumerDetails(customer)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error updating customer: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Customer updated successfully"})
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

	if customers == nil {
		customers = []models.Customer{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customers)
}

func (s *ApiServer) getCustumerCount(w http.ResponseWriter, r *http.Request) {
	count, err := s.Store.CountCustumer()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching customer: %v", err), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(count)
}
func (s *ApiServer) handleDeleteCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	err = s.Store.DeleteCustomer(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting customer: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Customer deleted successfully"})
}
