package service

type Hub struct {
	// Подключенные пользователи
	clients map[*Client]bool

	// Хэш мапа вебсокетов пользователей по их никам
	clientsID map[string]*Client

	// Регистрация клиента вебсокета
	register chan *Client

	// Отключение пользователя
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clientsID:  make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.clientsID[client.us] = client
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.clientsID, client.us)
				close(client.send)
			}
		}
	}
}
