package storage

import (
	"AAHAOMS/OMS/models"
)

type Storage interface {
	CreateCustomer(name string, number string, email string, country string, address string) (int, error)

	EditCustumerDetails(custumer models.Customer) error

	GetCustomerByID(id string) (*models.Customer, error)
	GetAllCustomers() ([]models.Customer, error)
	CountCustumer() (int, error)
	DeleteCustomer(id int) error

	///Order
	CreateOrder(order models.Order) (int, error)
	GetOrderHistoryByCustomerName(customerName string) ([]models.Order, error)
	GetOrderByID(orderID int) (models.Order, error)
	GetAllOrders() ([]models.Order, error)
	UpdateOrderStatus(orderID int, status string) error
	DeleteOrder(orderID int) error
	GetTotalOrderValueByCustomerName(customerName string) (float64, error)
	GetOrderCountByCustomerName(customerName string) (int, error)
	GetPendingOrderCount() (int, error)
	GetLatestOrderID() (int, error)
	GetOrdersByNameAndDate(customerName string, orderDate string) ([]models.Order, error)
	TotalOrderCount() (int, error)
	GetRecentOrders(limit int) ([]models.Order, error)

	//Shipement
	DeleteShipment(shipmentID int) error
	HandleShipment(shipment models.Shipment) error
	GetAllShipments() ([]models.Shipment, error)
	GetCompletedShipments() ([]models.Shipment, error)
	GetShippedButPendingShipments() ([]models.Shipment, error)
	GetDueItems(orderID int) ([]DueItem, error)
	GetItemByID(itemID int) (models.Item, error)
	GetTotalSalesForShippedOrders() (float64, error)
	GetTotalSalesForShippedOrdersByCustomer(customerName string) (float64, error)
	GetShipmentByName(customerName string) ([]models.Shipment, error)
	GetShipmentByID(shipmentID int) (*models.Shipment, error)

	// Auth user
	VerifyOtp(user models.AuthUser) (bool, error)
	IsUserExists(email string) (bool, error)
	AddOtp(user models.AuthUser) error
	AddAuthUser(user models.AuthUser) error
	IsKeyInStorage(token string) (bool, error)
}
