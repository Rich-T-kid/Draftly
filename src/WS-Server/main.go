package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type config struct {
	Host     string
	Port     string
	User     string
	Password string
	DbName   string
	CrudPort string
	WSPort   string
}

var (
	cfg        *config
	dbInstance *sql.DB
	upgrader   = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file ", err)
	}
	cfg = &config{
		Host:     must("POSTGRESS_HOST"),
		Port:     must("POSTGRESS_PORT"),
		User:     must("POSTGRESS_USER"),
		Password: must("POSTGRESS_PASSWORD"),
		DbName:   must("POSTGRESS_DB_NAME"),
		CrudPort: must("CRUD_PORT"),
		WSPort:   must("WS_PORT"),
	}
}
func must(name string) string {
	val := os.Getenv(name)
	if val == "" {
		log.Fatalf("Environment variable %s not set", name)
	}
	return val
}
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	_ = Connect() //if this doesnt panic the server is fine
	w.WriteHeader(http.StatusOK)
	resp := map[string]interface{}{
		"status": "WS Server is Live",
		"time":   fmt.Sprint(time.Now().Format(time.RFC3339)),
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Error generating JSON response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}
func webSocketHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("WebSocket connection established")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()
	for {
		// Read message from client
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		// TODO: Need to decode the message for the operation and then sent a response
		// for now this is fine
		log.Printf("Received: %s", message)
		var op = Operation{
			Kind:     "insert",
			Position: 0,
			Text:     string(message),
		}
		conn.WriteJSON(op)
	}
}
func routes() *mux.Router {
	r := mux.NewRouter()
	// Define your routes here
	r.HandleFunc("/health", HealthCheckHandler)
	r.HandleFunc("/ws", webSocketHandler)
	return r
}

func main() {
	fmt.Printf("server running on port :%s\n", cfg.WSPort)
	if err := http.ListenAndServe(":"+cfg.WSPort, routes()); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func Connect() *sql.DB {
	if dbInstance != nil {
		return dbInstance
	}
	// Connection parameters
	host := cfg.Host
	port := cfg.Port
	user := cfg.User
	password := cfg.Password
	dbname := cfg.DbName

	// Build connection string
	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	// Open database
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}

	// Verify connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
	}

	dbInstance = db
	return db
}

type Operation struct {
	Kind     string  `json:"kind"` // cant use type as field name because its a reserved word
	Position float64 `json:"position"`
	Text     string  `json:"text"`
}
