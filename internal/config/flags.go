package config

import (
	"flag"
	"github.com/vkobazev/goShortenerUrl/internal/consts"
	"os"
)

var Options struct {
	DefScheme       string
	DefHost         string
	DefPort         string
	ListenAddr      string
	ReturnAddr      string
	FileStoragePath string
	DataBaseConn    string
	//DBHost          string
	//DBPort          int
	//DBUser          string
	//DBPassword      string
	//DBName          string
}

func ConfigService() error {

	// Init flag strings
	Options.DefHost = consts.HTTPMethod + "://" + "localhost"
	Options.DefPort = "8080"

	flag.StringVar(&Options.ListenAddr, "a", ":8080", "Listen address")
	flag.StringVar(&Options.ReturnAddr, "b", "http://localhost:8080", "Return address")
	flag.StringVar(&Options.FileStoragePath, "f", "./data.json", "File storage path")
	flag.StringVar(&Options.DataBaseConn, "d", "", "Database connection string")
	flag.Parse()

	if addr := os.Getenv("SERVER_ADDRESS"); addr != "" {
		Options.ListenAddr = addr
	}
	if ReturnAddr := os.Getenv("BASE_URL"); ReturnAddr != "" {
		Options.ReturnAddr = ReturnAddr
	}
	if FileStoragePath := os.Getenv("FILE_STORAGE_PATH"); FileStoragePath != "" {
		Options.FileStoragePath = FileStoragePath
	}
	if DataBaseConn := os.Getenv("DATABASE_DSN"); DataBaseConn != "" {
		Options.DataBaseConn = DataBaseConn
	}
	return nil
}
