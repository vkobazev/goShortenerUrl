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
	flagAddr := flag.String("a", ":8080",
		"Setup Listen address for your instance of `shortener`,"+
			"or example `-a :8080`,`-a localhost:8080` and `-a 192.168.0.1:8080`")
	flag.StringVar(&Options.ReturnAddr, "b", Options.Host+":"+Options.Port,
		"Setup hostname returning to the client in body,"+
			"it can be usefull if you have balancer(Nginx,HAproxy) and registered domain(exm.org)."+
			"Run `-b exm.org` to get in request body `http://exm.org/eAskfc`.")
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
