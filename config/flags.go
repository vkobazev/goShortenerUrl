package config

import (
	"errors"
	"flag"
	"strconv"
	"strings"
)

var Options struct {
	DefScheme  string
	DefHost    string
	DefPort    string
	ListenAddr string
	ReturnAddr string
}

func ParseFlags() error {

	// Init flag strings
	Options.DefScheme = "http"
	Options.DefHost = Options.DefScheme + "://" + "localhost"
	Options.DefPort = "8080"

	a := flag.String("a", ":"+Options.DefPort,
		"Setup Listen address for your instance of `shortener`,"+
			"or example `-a :8080`,`-a localhost:8080` and `-a 192.168.0.1:8080`")
	b := flag.String("b", Options.DefHost+":"+Options.DefPort,
		"Setup hostname returning to the client in body,"+
			"it can be usefull if you have balancer(Nginx,HAproxy) and registered domain(exm.org)."+
			"Run `-b exm.org` to get in request body `http://exm.org/eAskfc`.")
	flag.Parse()

	// Parse string to Options struct
	as := strings.Split(*a, ":")
	if len(as) != 2 {
		return errors.New("Need address in a form host:port ")
	}
	// Check port for int convertation
	_, err := strconv.Atoi(as[1])
	if err != nil {
		return errors.New("Can't parse port as string value ")
	}
	if (as[0] != "") && (*b == Options.DefHost+":"+Options.DefPort) {
		Options.ListenAddr = as[0] + ":" + as[1]
		Options.ReturnAddr = Options.DefScheme + "://" + as[0] + ":" + as[1]
	}
	if (as[0] == "") && (*b == Options.DefHost+":"+Options.DefPort) {
		Options.ListenAddr = as[0] + ":" + as[1]
		Options.ReturnAddr = Options.DefHost + ":" + as[1]
	}
	if *b != Options.DefHost+":"+Options.DefPort {
		Options.ListenAddr = as[0] + ":" + as[1]
		Options.ReturnAddr = *b
	}

	return nil
}
