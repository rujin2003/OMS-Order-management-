package models

type Customer struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Number  int    `json:"number"`
	Email   string `json:"email"`
	Country string `json:"country"`
	Address string `json:"address"`
	
}
