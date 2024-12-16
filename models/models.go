package models

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/mattn/go-sqlite3"
)

const SUPERSECRET = "ILOVEFEMBOYS"

// Для авторизации
type Key struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
	Option   string `json:"option"`
}

// Модель контакта
type Contact struct {
	Name string `json:"name"`
	Contact string `json:"contact"`
}

// Модель пользователя
type UserModel struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// Сообщение
type Message struct {
	Sender    string `json:"sender"`
	Receiver  string `json:"receiver"`
	Text      string `json:"text"`
	AuthToken string `json:"token"`
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

// Открытие БД
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

// Создание таблиц БД
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

		contactsTable = `
		CREATE TABLE IF NOT EXISTS contacts(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			contact TEXT NOT NULL
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

	_, err = DB.ExecContext(ctx, contactsTable)
	if err != nil {
		log.Printf("Error create table contacts: %s", err)
	}
}

// Создание нового сообщения в БД
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

// Регистрация нового пользователя
func RegisterNewUser(usr RegisterForm) error {
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

// Получение сообщений для клиента
func GetMessages(key Key) ([]Message, error) {
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

func GetContacts(key Key) ([]Contact, error) {
	var (
		contacts []Contact
	)
	const q = "SELECT name, contact FROM contacts WHERE name = $1"
	ctx := context.TODO()

	r, err := DB.QueryContext(ctx, q, key.Username)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	for r.Next() {
		c := Contact{}
		err := r.Scan(&c.Name, &c.Contact)
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, c)
	}
	return contacts, nil
}

// Получение юзера по нику
func SelectUserByName(name string) (UserModel, error) {
	u := UserModel{}
	ctx := context.TODO()
	var q = "SELECT name, password FROM users WHERE name = $1"
	err := DB.QueryRowContext(ctx, q, name).Scan(&u.Name, &u.Password)
	if err != nil {
		return u, err
	}

	return u, nil
}

// Проверка есть ли такой контакт
func CheckContact(name string, contact string) (error) {
	ctx := context.TODO()
	var q = "SELECT name, contact FROM contacts WHERE name = $1 AND contact = $2"
	err := DB.QueryRowContext(ctx, q, name, contact)
	if err != nil {
		return fmt.Errorf("not found")
	}

	return nil
}

// Добавление пользователя в контакты
func AddContact(name string, contact string) (error) {
	q := "INSERT INTO contacts (name, contact) values ($1, $2)"
	ctx := context.TODO()
	ts, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	r, err := DB.ExecContext(ctx, q, name, contact)
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

// Создание JWT токена
func NewJWT(us string) string {
	tm := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name": us,
		"nbf":  tm.Unix(),
		"now":  tm.Add(672 * time.Hour).Unix(),
		"iat":  tm.Unix(),
	})

	tokenString, err := token.SignedString([]byte(SUPERSECRET))
	if err != nil {
		log.Printf("Error token gen for '%s': %s", us, err)
	}

	return tokenString
}

// Валидация токена
func ValidTocken(token string) (string, error) {
	tokenFromString, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			panic(fmt.Errorf("unexpected signing method: %v", token.Header["alg"]))
		}

		return []byte(SUPERSECRET), nil
	})

	if err != nil {
		log.Printf("Error on validation jwt token: %s", err)
		return "", err
	}

	if claims, ok := tokenFromString.Claims.(jwt.MapClaims); ok {
		return fmt.Sprint(claims["name"]), nil
	} else {
		return "", fmt.Errorf("error models.go : %s", err)
	}
}
