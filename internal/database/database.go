package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"time"
)

// DB represents the database connection

type DB struct {
	conn *pgx.Conn
}

type RequestData struct {
	ID  string `json:"correlation_id"`
	URL string `json:"original_url"`
}

// New creates a new database connection

func New(connString string) (*DB, error) {
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	return &DB{conn: conn}, nil
}

// Ping checks if the database connection is still alive

func (db *DB) Ping(ctx context.Context) error {
	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Ping the database
	err := db.conn.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	return nil
}

// Close closes the database connection

func (db *DB) Close(ctx context.Context) error {
	return db.conn.Close(ctx)
}

// CreateTable creates the URL table if it doesn't exist

func (db *DB) CreateTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,
			short_url VARCHAR(50) UNIQUE NOT NULL,
			long_url TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_short_url ON urls (short_url, long_url);
	`

	_, err := db.conn.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("error creating table: %v", err)
	}

	return nil
}

// InsertURL inserts a new short URL and its corresponding long URL into the database

func (db *DB) InsertURL(ctx context.Context, shortURL, longURL string) error {
	query := `
		INSERT INTO urls (short_url, long_url)
		VALUES ($1, $2)
		ON CONFLICT (short_url) DO UPDATE
		SET long_url = EXCLUDED.long_url
	`

	_, err := db.conn.Exec(ctx, query, shortURL, longURL)
	if err != nil {
		return fmt.Errorf("error inserting URL: %v", err)
	}

	return nil
}

// GetShortURL retrieves the short URL for a given long URL

func (db *DB) GetShortURL(ctx context.Context, longURL string) (string, error) {
	var shortURL string
	query := "SELECT short_url FROM urls WHERE long_url = $1"

	err := db.conn.QueryRow(ctx, query, longURL).Scan(&shortURL)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("long URL not found")
		}
		return "", fmt.Errorf("error retrieving short URL: %v", err)
	}

	return shortURL, nil
}

// GetLongURL retrieves the long URL for a given short URL

func (db *DB) GetLongURL(ctx context.Context, shortURL string) (string, error) {
	var longURL string
	query := "SELECT long_url FROM urls WHERE short_url = $1"

	err := db.conn.QueryRow(ctx, query, shortURL).Scan(&longURL)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("short URL not found")
		}
		return "", fmt.Errorf("error retrieving long URL: %v", err)
	}

	return longURL, nil
}

// LongURLExists checks if a long URL already exists in the database

func (db *DB) LongURLExists(ctx context.Context, longURL string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM urls WHERE long_url = $1)"

	err := db.conn.QueryRow(ctx, query, longURL).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking long URL existence: %v", err)
	}

	return exists, nil
}

// InsertURLs inserts multiple URL pairs into the database using a transaction

func (db *DB) InsertURLs(ctx context.Context, urlPairs []RequestData) error {
	// Start a transaction
	tx, err := db.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	// Ensure the transaction is rolled back if an error occurs
	defer tx.Rollback(ctx)

	// Prepare the query
	query := `
		INSERT INTO urls (short_url, long_url)
		VALUES ($1, $2)
		ON CONFLICT (short_url) DO UPDATE
		SET long_url = EXCLUDED.long_url
	`

	// Create a prepared statement
	stmt, err := tx.Prepare(ctx, "insert_urls", query)
	if err != nil {
		return fmt.Errorf("error preparing statement: %v", err)
	}

	// Insert each URL pair
	for _, pair := range urlPairs {
		_, err := tx.Exec(ctx, stmt.Name, pair.ID, pair.URL)
		if err != nil {
			return fmt.Errorf("error inserting URL pair (%s, %s): %v", pair.ID, pair.URL, err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}
