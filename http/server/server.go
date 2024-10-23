package server

import (
	"log"
	"net/http"

	"quetzalcoatl/http/handler"
	"quetzalcoatl/models"
	"quetzalcoatl/service"
)

func RunHTTPServer() {
	r := http.NewServeMux() // Роутер
	hub := service.NewHub()
	go hub.Run()

	// API
	r.HandleFunc("/getmsg", handler.GetMessage)
	r.HandleFunc("/reg", handler.Registration)
	r.HandleFunc("/login", handler.Login)
	r.HandleFunc("/msgs", handler.GetMessagesForClient)
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")

		if token == "" {
			log.Println("Unvailable token")
		}

		name, err := models.ValidTocken(token)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err = models.SelectUserByName(name)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		service.ServeWs(hub, w, r)
	})

	http.ListenAndServe(":80", r)
}
