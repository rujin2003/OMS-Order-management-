package storage

import (
	"AAHAOMS/OMS/models"
)

type Storage interface {
	CreateCustomer(name string, number int, email string, country string, address string) (int, error)
	CreateOrder(customerID int, orderDate, dueDate string) (int, error)
	AddItemToOrder(orderID int, name string, size *string, color *string, price float64) error
	CreateShipment(orderID int, items []int) error
	GetOrderHistoryByName(customerName string) ([]models.Order, error)
	GetShipmentHistoryByName(customerName string) ([]models.Shipment, error)
	GetCustomerByID(id string) (*models.Customer, error)
	GetAllCustomers() ([]models.Customer, error)
}
