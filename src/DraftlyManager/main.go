package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

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

var cfg *config

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

func main() {
	fmt.Println("Hello, World! WS")
	Connect()
	healthHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("CRUD Server is Live"))
	}
	http.HandleFunc("/health", healthHandler)
	log.Println("Starting CRUD server on :" + cfg.CrudPort)
	if err := http.ListenAndServe(":"+cfg.CrudPort, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func Connect() *sql.DB {
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

	log.Println("Connected to Postgres!")
	return db
}
