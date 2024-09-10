package main

import (
	"github.com/vkobazev/goShortenerUrl/config"
	"github.com/vkobazev/goShortenerUrl/webserver"
	"log"
)

func main() {

	// Parse Flags to set up Server
	err := config.ConfigService()
	if err != nil {
		log.Fatal("Can't parse port as string value")
	}

	// Start Web Server
	webserver.WebServer()
}
