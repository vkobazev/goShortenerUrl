package config

import (
	"errors"
	"flag"
	"strconv"
	"strings"
)

var Options struct {
	Host       string
	Port       string
	ReturnAddr string
}

func ParseFlags() error {

	// Init flag strings
	Options.Host = "localhost"
	Options.Port = "8080"
	flagAddr := flag.String("a", ":8080", "Net listen addr")
	flag.StringVar(&Options.ReturnAddr, "b", Options.Host+":"+Options.Port, "destination folder")
	flag.Parse()

	// Parse string to Options struct
	addrString := strings.Split(*flagAddr, ":")
	if len(addrString) != 2 {
		return errors.New("Need address in a form host:port")
	}
	// Check port for int convertation
	_, err := strconv.Atoi(addrString[1])
	if err != nil {
		return errors.New("Can't parse port as string value")
	}

	// Setup Host and Port
	Options.Host = addrString[0]
	Options.Port = addrString[1]

	if Options.ReturnAddr == "" {
		Options.ReturnAddr = Options.Host + ":" + Options.Port
	}

	return nil
}
