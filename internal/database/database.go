package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/vkobazev/goShortenerUrl/internal/config"
)

// PostgresConfig contains the configuration for connecting to a PostgreSQL database.

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// NewPostgresConnection creates a new SQL connection to a PostgreSQL database.

func NewDB() (*pgx.Conn, error) {
	connStr := config.Options.DataBaseConn //fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", config.Options.DBHost, config.Options.DBPort, config.Options.DBUser, config.Options.DBPassword, config.Options.DBName)

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	err = conn.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	fmt.Println("Successfully connected to the database!")
	return conn, nil
}
