package api

import (
	"AAHAOMS/OMS/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/xuri/excelize/v2"
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
func (s *ApiServer) handleGetDueItems(w http.ResponseWriter, r *http.Request) {
	// Extracting the order_id from the URL path
	vars := mux.Vars(r)
	orderIDStr, ok := vars["order_id"]
	if !ok {
		http.Error(w, "Missing order_id path parameter", http.StatusBadRequest)
		return
	}

	// Converting order_id to an integer
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid order_id, must be an integer", http.StatusBadRequest)
		return
	}

	// Fetching due items from the database
	dueItems, err := s.Store.GetDueItems(orderID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving due items: %v", err), http.StatusInternalServerError)
		return
	}

	// Sending response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dueItems)
}

func (s *ApiServer) handleGetShipmentHistoryByCustomerName(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	customerName, ok := vars["customer_name"]
	if !ok || customerName == "" {
		http.Error(w, "Customer name is required", http.StatusBadRequest)
		return
	}

	// Fetch the order history from the store
	shipment, err := s.Store.GetShipmentByName(customerName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching order history: %v", err), http.StatusInternalServerError)
		return
	}

	// Encode the orders as JSON and send as response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shipment)
}

func (s *ApiServer) handleGetShipmentByID(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	shipmentIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Missing shipment ID in the URL", http.StatusBadRequest)
		return
	}

	shipmentID, err := strconv.Atoi(shipmentIDStr)
	if err != nil {
		http.Error(w, "Invalid shipment ID, must be an integer", http.StatusBadRequest)
		return
	}

	shipment, err := s.Store.GetShipmentByID(shipmentID)
	if err != nil {
		if err.Error() == "shipment not found" {
			http.Error(w, "Shipment not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Error fetching shipment: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Encode the shipment as JSON and send as response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shipment)
}

func (s *ApiServer) handleDownloadShipmentExcel(w http.ResponseWriter, r *http.Request) {
	// Retrieve the shipment ID from URL parameters.
	vars := mux.Vars(r)
	shipmentIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Missing shipment ID in the URL", http.StatusBadRequest)
		return
	}

	shipmentID, err := strconv.Atoi(shipmentIDStr)
	if err != nil {
		http.Error(w, "Invalid shipment ID, must be an integer", http.StatusBadRequest)
		return
	}

	// Get the shipment details.
	shipment, err := s.Store.GetShipmentByID(shipmentID)
	if err != nil {
		if err.Error() == "shipment not found" {
			http.Error(w, "Shipment not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Error fetching shipment: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Since shipment does not have customer details, fetch the order using shipment.OrderID.
	order, err := s.Store.GetOrderByID(shipment.OrderID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching order details: %v", err), http.StatusInternalServerError)
		return
	}

	f := excelize.NewFile()
	sheetName := f.GetSheetName(0)

	centerStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Font: &excelize.Font{
			Bold: true,
		},
	})

	centerStyleNoBold, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})

	f.SetCellValue(sheetName, "A1", "AAHA FELT")
	f.SetCellValue(sheetName, "A2", "Bhanyatar, 8 Tokha")
	f.SetCellValue(sheetName, "A3", "Kathmandu, Nepal")
	f.SetCellValue(sheetName, "A4", "email-aahafelt@gmail.com")
	f.SetCellValue(sheetName, "A5", "Ph-015159015, 9851043414")

	for i := 1; i <= 5; i++ {
		startCell := fmt.Sprintf("A%d", i)
		endCell := fmt.Sprintf("E%d", i)
		f.MergeCell(sheetName, startCell, endCell)
		if i <= 3 {
			f.SetCellStyle(sheetName, startCell, endCell, centerStyle)
		} else {
			f.SetCellStyle(sheetName, startCell, endCell, centerStyleNoBold)
		}
	}

	orderDate := order.OrderDate[:10]
	shippedDate := shipment.ShippedDate[:10]

	f.SetCellValue(sheetName, "A7", "Customer Name:")
	f.SetCellValue(sheetName, "B7", order.CustomerName)
	f.SetCellValue(sheetName, "D7", "Order Date:")
	f.SetCellValue(sheetName, "E7", orderDate)

	f.SetCellValue(sheetName, "A8", "Shipment Date:")
	f.SetCellValue(sheetName, "B8", shippedDate)

	// --- Table Header for Shipment Items ---
	tableHeaderRow := 10
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", tableHeaderRow), "S.n.")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", tableHeaderRow), "Particular")
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", tableHeaderRow), "Qty")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", tableHeaderRow), "Rate")
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", tableHeaderRow), "Total")

	currentRow := tableHeaderRow + 1
	var grandTotal float64 = 0

	for i, item := range shipment.Items {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), i+1)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", currentRow), item.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", currentRow), item.Quantity)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", currentRow), item.Price)
		total := float64(item.Quantity) * item.Price
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", currentRow), total)
		grandTotal += total
		currentRow++
	}

	f.SetCellValue(sheetName, fmt.Sprintf("D%d", currentRow), "Grand Total")
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", currentRow), grandTotal)

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=shipment.xlsx")

	if err := f.Write(w); err != nil {
		http.Error(w, fmt.Sprintf("Error writing excel file: %v", err), http.StatusInternalServerError)
		return
	}
}
