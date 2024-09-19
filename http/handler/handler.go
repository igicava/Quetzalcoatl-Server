package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"quetzalcoatl/models"
)

func GetMessage(w http.ResponseWriter, r *http.Request) {
	// Get message by client
	var msg *models.Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("Error Decode msg")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := models.NewMessage(*msg); err != nil {
		log.Printf("By handler; %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func OnlineDevice(w http.ResponseWriter, r *http.Request) {
	// Set online user device
	var device *models.Info
	if err := json.NewDecoder(r.Body).Decode(&device); err != nil {
		log.Printf("Error decode info device")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	models.OnlineDevices[device.Username] = *device
	w.WriteHeader(http.StatusOK)
}

func Registration(w http.ResponseWriter, r *http.Request) {
	// Registration new  user
	if r.Method == "POST" {
		var usr *models.RegisterForm
		if err := json.NewDecoder(r.Body).Decode(&usr); err != nil {
			log.Printf("Error decoding registration form")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err := models.RegisterNewUser(*usr)
		if err != nil {
			log.Printf("Error insert reg-user: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func GetMessagesForClient(w http.ResponseWriter, r *http.Request) {
	// Get messages for client
	var k *models.KeyForMessages
	if err := json.NewDecoder(r.Body).Decode(&k); err != nil {
		log.Printf("Error decoding data for messages")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ms, err := models.GetMessages(*k)
	if err != nil {
		log.Printf("Error select messages: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": ms,
	});
}
