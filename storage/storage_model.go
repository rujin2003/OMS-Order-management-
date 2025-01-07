package storage

import (
	"AAHAOMS/OMS/models"
)

type Storage interface {
	CreateCustomer(name string, number int, email string, country string, address string) (int, error)
	CreateOrder(customerID int, customerName, orderDate, shipmentDue, shipmentAddress string) (int, error)

	AddItemToOrder(orderID int, name string, size, color *string, price float64, quantity int) error
	CreateShipment(orderID int, items []int) error
	GetOrderHistoryByCustomerName(customerName string) ([]models.Order, error)
	GetShipmentHistoryByName(customerName string) ([]models.Shipment, error)
	GetCustomerByID(id string) (*models.Customer, error)
	GetAllCustomers() ([]models.Customer, error)
	GetOrderByID(orderID int) (*models.Order, error)
	GetAllOrders() ([]models.Order, error)
}
