package main

import (
	"github.com/vkobazev/goShortenerUrl/config"
	"log"
)

var (
	Urls = map[string]string{}
)

func main() {

	// Parse Flags to set up Server
	err := config.ParseFlags()
	if err != nil {
		log.Fatal("Can't parse port as string value")
	}

	// Start Web Server
	WebServer()
}
