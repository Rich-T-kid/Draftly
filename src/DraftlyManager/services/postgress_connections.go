package services

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// DatabaseService handles PostgreSQL connections and operations
type DatabaseService struct {
	db *sql.DB
}

// NewDatabaseService creates a new database service instance
func NewDatabaseService() (*DatabaseService, error) {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get environment variables
	host := os.Getenv("POSTGRESS_HOST")
	port := os.Getenv("POSTGRESS_PORT")
	user := os.Getenv("POSTGRESS_USER")
	password := os.Getenv("POSTGRESS_PASSWORD")
	dbname := os.Getenv("POSTGRESS_DB_NAME")

	missing := []string{}

	if host == "" {
		missing = append(missing, "host")
	}
	if port == "" {
		missing = append(missing, "port")
	}
	if user == "" {
		missing = append(missing, "user")
	}
	if password == "" {
		missing = append(missing, "password")
	}
	if dbname == "" {
		missing = append(missing, "dbname")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %v", missing)
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Open db connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")

	return &DatabaseService{db: db}, nil
}

// Close closes connection
func (ds *DatabaseService) Close() error {
	if ds.db != nil {
		return ds.db.Close()
	}
	return nil
}

// ExecuteQuery executes a SELECT query and returns results as a slice of maps
func (ds *DatabaseService) ExecuteQuery(query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := ds.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}

	for rows.Next() {
		// Create a slice of interface{} to hold column values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create a map for this row
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// Handle byte slices (common with PostgreSQL)
			if b, ok := val.([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}

		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// ExecuteNonQuery executes INSERT, UPDATE, DELETE queries
func (ds *DatabaseService) ExecuteNonQuery(query string, args ...interface{}) (int64, error) {
	result, err := ds.db.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute non-query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// ExecuteQueryRow executes a query that returns a single row
func (ds *DatabaseService) ExecuteQueryRow(query string, dest []interface{}, args ...interface{}) error {
	row := ds.db.QueryRow(query, args...)
	if err := row.Scan(dest...); err != nil {
		return fmt.Errorf("failed to scan row: %w", err)
	}
	return nil
}

// BeginTransaction starts a new transaction
func (ds *DatabaseService) BeginTransaction() (*sql.Tx, error) {
	tx, err := ds.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, nil
}
