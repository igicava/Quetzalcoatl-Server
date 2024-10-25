package service

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Время, разрешенное для отправки сообщения к партнеру.
	writeWait = 10 * time.Second

	// Время, разрешенное для чтения следующего сообщения pong от партнера.
	pongWait = 60 * time.Second

	// Отправлять пинги партнеру с этой периодичностью. Должно быть меньше pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Максимальный размер сообщения, разрешенный от партнера.
	maxMessageSize = 250000
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client - посредник между веб-сокетным соединением и хабом.
type Client struct {
	hub *Hub

	// Веб-сокетное соединение.
	conn *websocket.Conn

	// Буферизованный канал исходящих сообщений.
	send chan []byte

	// Юзернэйм клиента
	us string

	// Список получателей
	rcs map[string]*Client
}

// readPump отправляет сообщения из веб-сокетного соединения в хаб.
//
// Приложение запускает readPump в отдельной горутине для каждого соединения.
// Приложение гарантирует, что на соединении будет не более одного читателя,
// выполняя все чтения из этой горутины.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ошибка: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		var mdl *struct {
			Receiver string `json:"receiver"`
			Sender string `json:"sender"`
			Text string `json:"text"`
		}

		err = json.Unmarshal(message, &mdl)
		if err != nil {
			log.Printf("ошибка десериализации: %s", err)
		}

		if _, ok := c.hub.clientsID[mdl.Receiver]; ok {
			c.hub.clientsID[mdl.Receiver].send <- message
		} else {
			log.Println("Получатель не подключен")
		}
		c.send <- message
	}
}

// writePump отправляет сообщения из хаба в веб-сокетное соединение.
//
// Горутина, выполняющая writePump, запускается для каждого соединения.
// Приложение гарантирует, что на соединении будет не более одного писателя,
// выполняя все записи из этой горутины.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Хаб закрыл канал.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Добавляем ожидания чат-сообщения к текущему веб-сокетному сообщению.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	usr := r.URL.Query().Get("id")
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), us: usr, rcs: map[string]*Client{}}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
