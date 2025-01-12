package api

import (
	"AAHAOMS/OMS/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// PostShipmentHandler handles the creation of a new shipment
func (s *ApiServer) handlePostShipment(w http.ResponseWriter, r *http.Request) {
	var shipment models.Shipment

	if err := json.NewDecoder(r.Body).Decode(&shipment); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err := s.Store.HandleShipment(shipment)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error processing shipment: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "shipment processed successfully"})
}

func (s *ApiServer) handleGetAllShipments(w http.ResponseWriter, r *http.Request) {
	shipments, err := s.Store.GetAllShipments()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving all shipments: %v", err), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(shipments)
}

func (s *ApiServer) handleGetCompletedShipments(w http.ResponseWriter, r *http.Request) {
	shipments, err := s.Store.GetCompletedShipments()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving completed shipments: %v", err), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(shipments)
}
func (s *ApiServer) handleGetShippedButPendingShipments(w http.ResponseWriter, r *http.Request) {
	shipments, err := s.Store.GetShippedButPendingShipments()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving shipped but pending shipments: %v", err), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(shipments)
}
func (s *ApiServer) handleDeleteShipment(w http.ResponseWriter, r *http.Request) {
	// Parse shipment ID from URL parameters
	vars := mux.Vars(r)
	shipmentID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid shipment ID", http.StatusBadRequest)
		return
	}

	// Call the DeleteShipment function
	err = s.Store.DeleteShipment(shipmentID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting shipment: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Shipment deleted successfully"})
}
