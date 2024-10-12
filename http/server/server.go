package server

import (
	"net/http"

	"quetzalcoatl/http/handler"
	"quetzalcoatl/service"
)

func RunHTTPServer() {
	r := http.NewServeMux() // Main router
	hub := service.NewHub()
	go hub.Run()

	r.HandleFunc("/getmsg", handler.GetMessage)
	r.HandleFunc("/setonline", handler.OnlineDevice)
	r.HandleFunc("/reg", handler.Registration)
	r.HandleFunc("/msgs", handler.GetMessagesForClient)
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		service.ServeWs(hub, w, r)
	})

	http.ListenAndServe(":8888", r)
}