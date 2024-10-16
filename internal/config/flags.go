package config

import (
	"errors"
	"flag"
	"github.com/vkobazev/goShortenerUrl/internal/consts"
	"os"
	"strconv"
	"strings"
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

	a := flag.String("a", ":"+Options.DefPort,
		"Setup Listen address for your instance of `shortener`,"+
			"or example `-a :8080`,`-a localhost:8080` and `-a 192.168.0.1:8080`")
	b := flag.String("b", Options.DefHost+":"+Options.DefPort,
		"Setup hostname returning to the client in body,"+
			"it can be usefull if you have balancer(Nginx,HAproxy) and registered domain(exm.org)."+
			"Run `-b exm.org` to get in request body `http://exm.org/eAskfc`.")
	f := flag.String("f", "./data.json", "File storage path")
	d := flag.String("d", "", "DB storage connection")
	flag.Parse()

	// Parse string to Options struct
	as := strings.Split(*a, ":")
	if len(as) != 2 {
		return errors.New("need address in a form host:port")
	}
	// Check port for int convertation
	_, err := strconv.Atoi(as[1])
	if err != nil {
		return errors.New("cant parse port as string value")
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

	if ListenAddr := os.Getenv("SERVER_ADDRESS"); ListenAddr != "" {
		Options.ListenAddr = ListenAddr
		if *b != Options.DefHost+":"+Options.DefPort {
			Options.ReturnAddr = *b
		}
		ls := strings.Split(Options.ListenAddr, ":")
		if ls[0] == "" {
			Options.ReturnAddr = Options.DefHost + ":" + ls[1]
		} else {
			Options.ReturnAddr = ls[0] + ":" + ls[1]
		}
	}
	if ReturnAddr := os.Getenv("BASE_URL"); ReturnAddr != "" {
		Options.ReturnAddr = ReturnAddr
	}
	if FileStoragePath := os.Getenv("FILE_STORAGE_PATH"); FileStoragePath == "" {
		Options.FileStoragePath = *f
	} else {
		Options.FileStoragePath = FileStoragePath
	}

	if DataBaseConn := os.Getenv("DATABASE_DSN"); DataBaseConn == "" {
		Options.DataBaseConn = *d
	} else {
		Options.DataBaseConn = DataBaseConn
	}

	return nil
}
