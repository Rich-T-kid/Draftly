package internal

import (
	"database/sql"
	"fmt"
	"log"
)

var (
	DbInstance *sql.DB
)

type WriteStore struct {
	db *sql.DB
}
type Operation struct {
	Kind     string  `json:"kind"` // cant use type as field name because its a reserved word
	Position float64 `json:"position"`
	Text     string  `json:"text"`
	Version  int     `json:"version"`
}

func (o Operation) Validate() error {
	if o.Kind != "insert" && o.Kind != "delete" {
		return fmt.Errorf("invalid operation kind: %s", o.Kind)
	}
	if o.Position < 0 {
		return fmt.Errorf("position cannot be negative: %f", o.Position)
	}
	if o.Text == "" {
		return fmt.Errorf("text cannot be empty")
	}
	if o.Version < 0 {
		return fmt.Errorf("version cannot be negative: %d", o.Version)
	}
	return nil
}
func NewWriteStore() *WriteStore {
	if DbInstance == nil {
		DbInstance = Connect()
	}
	return &WriteStore{db: DbInstance}
}

// Postgress

func Connect() *sql.DB {
	if DbInstance != nil {
		return DbInstance
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

	DbInstance = db
	return db
}

func (w *WriteStore) WriteOperation(roomID string, op Operation, timestamp string) error {
	return nil
}
