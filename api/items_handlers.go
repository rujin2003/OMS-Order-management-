package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (s *ApiServer) handleGetItemByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	item, err := s.Store.GetItemByID(itemID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching item: %v", err), http.StatusInternalServerError)
		return
	}

	if item.ID == 0 {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}
