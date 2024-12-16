package main

import (
	"log"
	
	"quetzalcoatl/http/server"
	"quetzalcoatl/models"
)

func main() {
	go func() { // Инициализация БД
		models.OpenDB()
		models.CreateTables()
		log.Println("DB is start")
	}()
	
	log.Println("Server (ver. 16.12) is start on port 8888") 
	server.RunHTTPServer() // Запуск сервера
}