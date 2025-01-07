// i am making a go application i wanna make a order management system  in golang where there are three databases customer , order , shipments all of them interrelated here customer data base holds the data of customers and the orders and the shipment history , in order database i need to store the order date , shipment due of the order , items (within items i have name , optional size,color, req price), another database holds the shipment delivered how the system works is suppose a customers order a number of
// items with due on a certain date and then i the
//  manufacture wasn't able to complete the whole
//   order but delivered some items in the database
//   how it need to work is it needs to show the remaining due order and maybe
//   i can send it in next shipment you can improvise this system and make it better please use Postgres for this task make the code modules
//   it different packages and readable make the system better as required here is the base code package
//    api

package main

import (
	"AAHAOMS/OMS/api"
	"AAHAOMS/OMS/storage"
	"fmt"
)

func main() {
	store, err := storage.NewPostgresStorage()
	if err != nil {
		fmt.Println("Failed to initialize storage:", err)
		return
	}
	defer store.Close()

	if err := store.Init(); err != nil {
		fmt.Println("Failed to initialize database:", err)
		return
	}

	server := api.NewApiServer(":8080", store)
	server.Start()
}
