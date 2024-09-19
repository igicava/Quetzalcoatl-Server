package main

import (
	"log"
	
	"quetzalcoatl/http/server"
	"quetzalcoatl/models"
)

func main() {
	go func() { // Init DB
		models.OpenDB()
		models.CreateTables()
		models.OnlineDevices = make(map[string]models.Info)
		log.Println("DB is start")
	}()
	
	log.Println("Server is start on port 8888") 
	server.RunHTTPServer() // Run server
}