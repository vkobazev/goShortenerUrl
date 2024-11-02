package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vkobazev/goShortenerUrl/internal/consts"
	"time"
)

// DB представляет пул соединений с базой данных
type DB struct {
	pool *pgxpool.Pool
}

// RequestData представляет данные запроса для вставки URL
type RequestData struct {
	ID     string `json:"correlation_id"`
	URL    string `json:"original_url"`
	UserID string `json:"user_id"`
}

// URLResponse представляет ответ с коротким и оригинальным URL
type URLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func New(connString string) (*DB, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("не удалось разобрать строку подключения: %v", err)
	}

	// Опциональная настройка параметров пула соединений
	config.MaxConns = 25 // Максимальное количество соединений в пуле
	config.MinConns = 5  // Минимальное количество соединений в пуле

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать пул соединений: %v", err)
	}

	return &DB{pool: pool}, nil
}

func (db *DB) Close() {
	db.pool.Close()
}

func (db *DB) Ping(ctx context.Context) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := db.pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("не удалось пинговать базу данных: %v", err)
	}

	return nil
}

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

	_, err := db.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("ошибка при создании таблицы: %v", err)
	}

	return nil
}

func (db *DB) InsertURL(ctx context.Context, shortURL, longURL, userID string) error {
	query := `
        INSERT INTO urls (short_url, long_url, user_id)
        VALUES ($1, $2, $3)
        ON CONFLICT (short_url) DO UPDATE
        SET long_url = EXCLUDED.long_url,
            user_id = EXCLUDED.user_id
    `

	_, err := db.pool.Exec(ctx, query, shortURL, longURL, userID)
	if err != nil {
		return fmt.Errorf("ошибка при вставке URL: %v", err)
	}

	return nil
}

func (db *DB) GetShortURL(ctx context.Context, longURL, userID string) (string, error) {
	var shortURL string
	query := `
		SELECT short_url 
		FROM urls 
		WHERE long_url = $1 AND user_id = $2
	`

	err := db.pool.QueryRow(ctx, query, longURL, userID).Scan(&shortURL)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return "", fmt.Errorf("длинный URL не найден для пользователя")
		}
		return "", fmt.Errorf("ошибка при получении короткого URL: %v", err)
	}

	return shortURL, nil
}

func (db *DB) GetLongURL(ctx context.Context, shortURL string) (string, string, error) {
	var longURL, userID string
	query := `
		SELECT long_url, user_id 
		FROM urls 
		WHERE short_url = $1 AND deleted = FALSE
	`

	err := db.pool.QueryRow(ctx, query, shortURL).Scan(&longURL, &userID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return "", "", fmt.Errorf("короткий URL не найден или удален")
		}
		return "", "", fmt.Errorf("ошибка при получении длинного URL: %v", err)
	}

	return longURL, userID, nil
}

func (db *DB) LongURLExists(ctx context.Context, longURL, userID string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM urls WHERE long_url = $1 AND user_id = $2)"

	err := db.pool.QueryRow(ctx, query, longURL, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("ошибка при проверке существования длинного URL: %v", err)
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

	err := db.pool.QueryRow(ctx, query, shortURL).Scan(&originalURL, &deleted)
	if err != nil {
		return "", false, fmt.Errorf("ошибка при получении URL: %v", err)
	}

	return originalURL, deleted, nil
}

func (db *DB) InsertURLs(ctx context.Context, urlPairs []RequestData) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ошибка при начале транзакции: %v", err)
	}
	defer tx.Rollback(ctx)

	query := `
        INSERT INTO urls (short_url, long_url, user_id)
        VALUES ($1, $2, $3)
        ON CONFLICT (short_url) DO UPDATE
        SET long_url = EXCLUDED.long_url,
            user_id = EXCLUDED.user_id
    `

	for _, pair := range urlPairs {
		_, err := tx.Exec(ctx, query, pair.ID, pair.URL, pair.UserID)
		if err != nil {
			return fmt.Errorf("ошибка при вставке пары URL (%s, %s, %s): %v",
				pair.ID, pair.URL, pair.UserID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("ошибка при коммите транзакции: %v", err)
	}

	return nil
}

func (db *DB) GetURLsByUser(ctx context.Context, userID string) ([]URLResponse, error) {
	query := `
		SELECT short_url, long_url 
		FROM urls 
		WHERE user_id = $1
	`

	rows, err := db.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе URL: %v", err)
	}
	defer rows.Close()

	var urls []URLResponse
	for rows.Next() {
		var shortURL, longURL string
		err := rows.Scan(&shortURL, &longURL)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании строки URL: %v", err)
		}

		urls = append(urls, URLResponse{
			ShortURL:    consts.BaseURL + shortURL,
			OriginalURL: longURL,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации строк URL: %v", err)
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

	result, err := db.pool.Exec(ctx, query, userID, shortURLs)
	if err != nil {
		return fmt.Errorf("ошибка при пометке URL как удаленных: %v", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("не было обновлено ни одного URL")
	}

	return nil
}
