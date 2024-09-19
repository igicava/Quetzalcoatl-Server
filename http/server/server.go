package server

import (
	"net/http"

	"quetzalcoatl/http/handler"
)

func RunHTTPServer() {
	r := http.NewServeMux() // Main router

	r.HandleFunc("/getmsg", handler.GetMessage)
	r.HandleFunc("/setonline", handler.OnlineDevice)
	r.HandleFunc("/reg", handler.Registration)
	r.HandleFunc("/msgs", handler.GetMessagesForClient)

	http.ListenAndServe(":8888", r)
}