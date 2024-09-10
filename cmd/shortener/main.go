package main

import (
	"errors"
	"github.com/vkobazev/goShortenerUrl/config"
)

var (
	Urls = map[string]string{}
)

func main() {

	// Parse Flags to set up Server
	err := config.ParseFlags()
	if err != nil {
		errors.New("Can't parse port as string value")
	}

	// Start Web Server
	WebServer()

}
