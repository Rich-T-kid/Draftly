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
	Kind           string `json:"kind"` // cant use type as field name because its a reserved word
	Position       int    `json:"position"`
	Text           string `json:"text"`
	SequenceNumber int    `json:"sequence_number"`
	CursorPosition int    `json:"cursor_position"`
	Version        int    `json:"version"`
}

func (o Operation) Validate() error {
	if o.Kind != "insert" && o.Kind != "delete" {
		return fmt.Errorf("invalid operation kind: %s", o.Kind)
	}
	if o.Position < 0 {
		return fmt.Errorf("position cannot be negative: %d", o.Position)
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
	w.file.Write([]byte(fmt.Sprintf("RoomID:(%s),%s,%d,%s,%d,%d,%d,%s\n",
		roomID, op.Kind, op.Position, op.Text, op.Version, op.SequenceNumber, op.CursorPosition, timestamp.Format(time.RFC3339))))
	return nil
	// roomID, kind, position, text, version, sequence_number, cursor_position, timestamp
}
func (w *WriteStore) OperationsSince(roomID string, timestamp time.Time) ([]Operation, error) {
	content, err := os.ReadFile(w.file.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	var operations []Operation
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		fmt.Printf("processing line: '%s'\n", line)
		if line == "" {
			continue
		}
		if !strings.Contains(line, fmt.Sprintf("RoomID:(%s)", roomID)) {
			continue
		}

		// Split the line into parts
		parts := strings.Split(line, ",")
		if len(parts) < 7 {
			fmt.Printf("skipping malformed line: '%s'\n", line)
			continue
		}
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
			fmt.Printf("part[%d]: '%s'\n", i, parts[i])
		}

		// Parse operation details
		opType := parts[1]
		pos, _ := strconv.Atoi(parts[2])
		text := parts[3]
		version, _ := strconv.Atoi(parts[4])
		sequenceNumber, _ := strconv.Atoi(parts[5])
		cursorPosition, _ := strconv.Atoi(parts[6])

		operations = append(operations, Operation{
			Kind:           opType,
			Position:       int(pos),
			Text:           text,
			Version:        version,
			SequenceNumber: sequenceNumber,
			CursorPosition: cursorPosition,
		})
	}
	// this is shorted by time by default because we read the file top to bottom
	return operations, nil
}
