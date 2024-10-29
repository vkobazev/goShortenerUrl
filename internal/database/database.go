package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/vkobazev/goShortenerUrl/internal/consts"
	"sync"
	"time"
)

// DB represents the database connection

type DB struct {
	conn *pgx.Conn
	mu   sync.RWMutex
}

type RequestData struct {
	ID     string `json:"correlation_id"`
	URL    string `json:"original_url"`
	UserID string `json:"user_id"`
}

type URLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func New(connString string) (*DB, error) {
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	return &DB{conn: conn}, nil
}

func (db *DB) Close(ctx context.Context) error {
	return db.conn.Close(ctx)
}

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

// Modified CreateTable function to include user_id

func (db *DB) CreateTable(ctx context.Context) error {
	query := `
        CREATE TABLE IF NOT EXISTS urls (
            id SERIAL PRIMARY KEY,
            short_url VARCHAR(50) UNIQUE NOT NULL,
            long_url TEXT NOT NULL,
            user_id VARCHAR(50) NOT NULL,
            deleted BOOLEAN DEFAULT FALSE,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );
        CREATE INDEX IF NOT EXISTS idx_short_url ON urls (short_url, long_url);
        CREATE INDEX IF NOT EXISTS idx_user_id ON urls (user_id);
    `

	_, err := db.conn.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("error creating table: %v", err)
	}

	return nil
}

// Modified InsertURL function to include user_id

func (db *DB) InsertURL(ctx context.Context, shortURL, longURL, userID string) error {
	query := `
        INSERT INTO urls (short_url, long_url, user_id)
        VALUES ($1, $2, $3)
        ON CONFLICT (short_url) DO UPDATE
        SET long_url = EXCLUDED.long_url,
            user_id = EXCLUDED.user_id
    `

	_, err := db.conn.Exec(ctx, query, shortURL, longURL, userID)
	if err != nil {
		return fmt.Errorf("error inserting URL: %v", err)
	}

	return nil
}

// Modified GetShortURL function to include user_id

func (db *DB) GetShortURL(ctx context.Context, longURL, userID string) (string, error) {
	var shortURL string
	query := `
		SELECT short_url 
		FROM urls 
		WHERE long_url = $1 AND user_id = $2
	`

	err := db.conn.QueryRow(ctx, query, longURL, userID).Scan(&shortURL)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("long URL not found for user")
		}
		return "", fmt.Errorf("error retrieving short URL: %v", err)
	}

	return shortURL, nil
}

// Modified GetLongURL function to include user_id check

func (db *DB) GetLongURL(ctx context.Context, shortURL string) (string, string, error) {
	var longURL, userID string
	query := `
		SELECT long_url, user_id 
		FROM urls 
		WHERE short_url = $1 AND deleted = FALSE
	`

	err := db.conn.QueryRow(ctx, query, shortURL).Scan(&longURL, &userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", "", fmt.Errorf("short URL not found или удален")
		}
		return "", "", fmt.Errorf("error retrieving long URL: %v", err)
	}

	return longURL, userID, nil
}

// Modified LongURLExists function to include user_id

func (db *DB) LongURLExists(ctx context.Context, longURL, userID string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM urls WHERE long_url = $1 AND user_id = $2)"

	err := db.conn.QueryRow(ctx, query, longURL, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking long URL existence: %v", err)
	}

	return exists, nil
}

func (db *DB) LongURLDeleted(ctx context.Context, shortURL string) (string, bool, error) {
	query := `
        SELECT long_url, deleted
        FROM urls
        WHERE short_url = $1
    `

	var originalURL string
	var deleted bool

	err := db.conn.QueryRow(ctx, query, shortURL).Scan(&originalURL, &deleted)
	if err != nil {
		return "", false, fmt.Errorf("error retrieving URL: %v", err)
	}

	return originalURL, deleted, nil
}

// Modified InsertURLs function to include user_id

func (db *DB) InsertURLs(ctx context.Context, urlPairs []RequestData) error {
	tx, err := db.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	query := `
        INSERT INTO urls (short_url, long_url, user_id)
        VALUES ($1, $2, $3)
        ON CONFLICT (short_url) DO UPDATE
        SET long_url = EXCLUDED.long_url,
            user_id = EXCLUDED.user_id
    `

	stmt, err := tx.Prepare(ctx, "insert_urls", query)
	if err != nil {
		return fmt.Errorf("error preparing statement: %v", err)
	}

	for _, pair := range urlPairs {
		_, err := tx.Exec(ctx, stmt.Name, pair.ID, pair.URL, pair.UserID)
		if err != nil {
			return fmt.Errorf("error inserting URL pair (%s, %s, %s): %v",
				pair.ID, pair.URL, pair.UserID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

func (db *DB) GetURLsByUser(ctx context.Context, userID string) ([]URLResponse, error) {
	query := `
		SELECT short_url, long_url 
		FROM urls 
		WHERE user_id = $1`

	rows, err := db.conn.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying URLs: %v", err)
	}
	defer rows.Close()

	var urls []URLResponse
	for rows.Next() {
		var shortURL, longURL string
		err := rows.Scan(&shortURL, &longURL)
		if err != nil {
			return nil, fmt.Errorf("error scanning URL row: %v", err)
		}

		urls = append(urls, URLResponse{
			ShortURL:    consts.BaseURL + shortURL,
			OriginalURL: longURL,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating URL rows: %v", err)
	}

	return urls, nil
}

func (db *DB) DeleteURLforUser(ctx context.Context, userID string, shortURLs []string) error {
	if len(shortURLs) == 0 {
		return nil
	}

	query := `
        UPDATE urls
        SET deleted = TRUE
        WHERE user_id = $1
          AND short_url = ANY($2::text[])
    `

	result, err := db.conn.Exec(ctx, query, userID, shortURLs)
	if err != nil {
		return fmt.Errorf("error marking URLs as deleted: %v", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no URLs were updated")
	}

	return nil
}
