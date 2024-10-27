package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"quetzalcoatl/models"
)

// Получение сообщений от клиента
func GetMessage(w http.ResponseWriter, r *http.Request) {
	var msg *models.Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("Error Decode msg")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	name, err := models.ValidTocken(msg.AuthToken)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_, err = models.SelectUserByName(name)
	if err != nil {
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

// Регистрация нового пользователя
func Registration(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		log.Println("user join to register")
		var usr *models.RegisterForm
		if err := json.NewDecoder(r.Body).Decode(&usr); err != nil {
			log.Printf("Error decoding registration form")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err := models.SelectUserByName(usr.Username)
		if err == nil {
			log.Printf("User already exists: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if usr.FirstName == "" || usr.LastName == "" || usr.Username == "" || usr.Password_1 == "" || usr.Password_1 != usr.Password_2 {
			w.WriteHeader(418)
			return
		}

		err = models.RegisterNewUser(*usr)
		if err != nil {
			log.Printf("Error insert reg-user: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// Выдача сообщений для клиента
func GetMessagesForClient(w http.ResponseWriter, r *http.Request) {
	var k *models.Key
	if err := json.NewDecoder(r.Body).Decode(&k); err != nil {
		log.Printf("Error decoding data for messages")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	name, err := models.ValidTocken(k.Token)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = models.SelectUserByName(name)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	k.Username = name

	ms, err := models.GetMessages(*k)
	if err != nil {
		log.Printf("Error select messages: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cs, err := models.GetContacts(*k)
	if err != nil {
		log.Printf("Error get contacts: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": ms,
		"contacts": cs,
	});
}

// Авторизация и выдача токена йоу
func Login(w http.ResponseWriter, r *http.Request) {
	var u *models.Key
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		log.Printf("Error decoding data for login")
		w.WriteHeader(http.StatusBadRequest)
		return 
	}

	user, err := models.SelectUserByName(u.Username)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if user.Password != u.Password {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token := models.NewJWT(user.Name)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
	});
}

// Добавление в Контакты
func NewContact(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var op *models.Key
		if err := json.NewDecoder(r.Body).Decode(&op); err != nil {
			log.Printf("Error decoding data for append contact: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		us, err := models.ValidTocken(op.Token)
		if err != nil {
			log.Printf("Error validation token: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		user, err := models.SelectUserByName(us)
		if err != nil {
			log.Printf("Not found user err: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = models.CheckContact(user.Name, op.Option)
		if err != nil {
			log.Println("contact already exists")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = models.AddContact(user.Name, op.Option) 
		if err != nil {
			log.Printf("Error add contact: %s", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
	}
}