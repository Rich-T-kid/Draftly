package internal

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	DbInstance *sql.DB
	logName    = "operations.log"
)

type WriteStore struct {
	db   *sql.DB
	file *os.File
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
func NewWriteStore() (*WriteStore, error) {
	if DbInstance == nil {
		DbInstance = Connect()
	}
	f, err := os.OpenFile(logName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, errors.New("failed to open operations log file: " + err.Error())
	}
	return &WriteStore{db: DbInstance, file: f}, nil
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

func (w *WriteStore) WriteOperation(roomID string, op Operation, timestamp time.Time) error {
	// TODO: replace with db later
	w.file.Write([]byte(fmt.Sprintf("RoomID: (%s), Operation: %s, %f, %s, Version: %d TimeStamp: %s\n", roomID, op.Kind, op.Position, op.Text, op.Version, timestamp.Format(time.RFC3339))))
	return nil
}
func (w *WriteStore) OperationsSince(roomID string, timestamp time.Time) ([]Operation, error) {
	content, err := os.ReadFile(w.file.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// TODO: Grab all the operations from the database since the given timestamp
	var operations []Operation
	lines := strings.Split(string(content), "\n")
	fmt.Println("lines:", lines)
	for _, line := range lines {
		if line == "" {
			continue
		}
		if !strings.Contains(line, fmt.Sprintf("RoomID: (%s)", roomID)) {
			continue
		}

		// Format: "RoomID: (roomID), Operation: kind, position, text, Version: version TimeStamp: time"
		parts := strings.Split(line, ", ") // Note: splitting on ", " instead of ","
		if len(parts) < 4 {
			continue
		}

		// Extract timestamp
		tsStr := strings.TrimPrefix(parts[len(parts)-1], "TimeStamp: ")
		ts, err := time.Parse(time.RFC3339, strings.TrimSpace(tsStr))
		if err != nil {
			continue
		}

		// Check if operation is after the given timestamp
		if ts.After(timestamp) {
			// Parse operation details
			opParts := strings.Split(parts[1], ": ")
			kind := strings.TrimSpace(opParts[1])
			position, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
			text := strings.TrimSpace(parts[3])
			versionPart := strings.Split(parts[4], "Version: ")
			version, _ := strconv.Atoi(strings.TrimSpace(versionPart[1]))

			operations = append(operations, Operation{
				Kind:     kind,
				Position: position,
				Text:     text,
				Version:  version,
			})
		}
	}

	return operations, nil
}
