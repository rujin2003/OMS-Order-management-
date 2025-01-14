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
