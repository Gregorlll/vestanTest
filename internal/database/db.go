package database

import (
	"database/sql"
	"log"
	"os"

	models "vestantest/internal/models"

	_ "github.com/lib/pq"
)

type DB struct {
	db     *sql.DB
	logger *log.Logger
}

func NewDB(connStr string) (*DB, error) {
	logger := log.New(os.Stdout, "[DB] ", log.LstdFlags)
	logger.Printf("Connecting to database...")

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Printf("Database connection failed: %v", err)
		return nil, err
	}

	if err = db.Ping(); err != nil {
		logger.Printf("Database ping failed: %v", err)
		return nil, err
	}

	logger.Println("Successfully connected to database")

	// Create tables if they don't exist
	if err = createTables(db); err != nil {
		logger.Printf("Failed to create tables: %v", err)
		return nil, err
	}

	return &DB{
		db:     db,
		logger: logger,
	}, nil
}

func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS messages (
            id SERIAL PRIMARY KEY,
            username VARCHAR(10) NOT NULL,
            message TEXT NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT NOW()
        )`,
		`CREATE TABLE IF NOT EXISTS connection_logs (
            id SERIAL PRIMARY KEY,
            username VARCHAR(10) NOT NULL,
            event_type VARCHAR(20) NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT NOW()
        )`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) SaveMessage(username, message string) error {
	db.logger.Printf("Saving message from %s", username)
	query := `INSERT INTO messages (username, message) VALUES ($1, $2)`
	_, err := db.db.Exec(query, username, message)
	if err != nil {
		db.logger.Printf("Error saving message: %v", err)
	}
	return err
}

func (db *DB) LogConnection(username, eventType string) error {
	db.logger.Printf("Logging connection event: %s - %s", username, eventType)
	query := `INSERT INTO connection_logs (username, event_type) VALUES ($1, $2)`
	_, err := db.db.Exec(query, username, eventType)
	if err != nil {
		db.logger.Printf("Error logging connection: %v", err)
	}
	return err
}

func (db *DB) GetMessages(page, pageSize int) ([]models.Message, int, error) {
	var total int
	err := db.db.QueryRow("SELECT COUNT(*) FROM messages").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `
        SELECT username, message, created_at 
        FROM messages 
        ORDER BY created_at DESC 
        LIMIT $1 OFFSET $2
    `

	rows, err := db.db.Query(query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(&msg.User, &msg.Message, &msg.Time)
		if err != nil {
			return nil, 0, err
		}
		messages = append(messages, msg)
	}

	return messages, total, nil
}

func (db *DB) GetConnectionHistory() ([]models.ConnectionEvent, error) {
	query := `
        SELECT username, event_type, created_at 
        FROM connection_logs 
        ORDER BY created_at DESC
    `

	rows, err := db.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.ConnectionEvent
	for rows.Next() {
		var event models.ConnectionEvent
		err := rows.Scan(&event.User, &event.EventType, &event.Time)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}
