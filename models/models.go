package models

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// Для авторизации
type KeyForMessages struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Сообщение 
type Message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Text     string `json:"text"`
}

// Для регистрации
type RegisterForm struct {
	Username   string `json:"username"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Password_1 string `json:"password_1"`
	Password_2 string `json:"password_2"`
}

var DB *sql.DB // БД

func OpenDB() {
	var err error
	ctx := context.TODO()

	DB, err = sql.Open("sqlite3", "data.db")
	if err != nil {
		log.Fatalf("Error open DB: %s", err)
	}

	err = DB.PingContext(ctx)
	if err != nil {
		log.Fatalf("Error ping DB: %s", err)
	}
}

func CreateTables() {
	ctx := context.TODO()
	const (
		messageTable = `
		CREATE TABLE IF NOT EXISTS messages(
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			receiver TEXT NOT NULL,
			sender TEXT NOT NULL,
			text TEXT NOT NULL
		);`

		userTable = `
		CREATE TABLE IF NOT EXISTS users(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			password TEXT NOT NULL,
			first_name TEXT, 
			last_name TEXT
		);`
	)

	_, err := DB.ExecContext(ctx, messageTable)
	if err != nil {
		log.Printf("Error create table messages: %s", err)
	}

	_, err = DB.ExecContext(ctx, userTable)
	if err != nil {
		log.Printf("Error create table users: %s", err)
	}
}

func NewMessage(msg Message) error {
	ctx := context.TODO()
	q := "INSERT INTO messages (receiver, sender, text) values ($1, $2, $3)"

	ts, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	r, err := DB.ExecContext(ctx, q, msg.Receiver, msg.Sender, msg.Text)
	if err != nil {
		return err
	}

	_, err = r.LastInsertId()
	if err != nil {
		return err
	}

	if err = ts.Commit(); err != nil {
		return err
	}

	return nil
}

func RegisterNewUser(usr RegisterForm) (error) {
	ctx := context.TODO()
	ts, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := "INSERT INTO users(name, first_name, last_name, password) values ($1, $2, $3, $4)"
	r, err := DB.ExecContext(ctx, q, usr.Username, usr.FirstName, usr.LastName, usr.Password_1)
	if err != nil {
		return err
	}

	_, err = r.LastInsertId()
	if err != nil {
		return err
	}

	if err = ts.Commit(); err != nil {
		return err
	}

	return nil
}

func GetMessages(key KeyForMessages) ([]Message, error) {
	var (
		messages []Message
	)
	const q = "SELECT sender, receiver, text FROM messages WHERE sender = $1 OR receiver = $1"
	ctx := context.TODO()

	r, err := DB.QueryContext(ctx, q, key.Username)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	for r.Next() {
		m := Message{}
		err := r.Scan(&m.Sender, &m.Receiver, &m.Text)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}