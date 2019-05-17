package main

import (
	"flag"
	"io/ioutil"
	"os"
	"strings"

	"githib.com/A-UNDERSCORE-D/goRRProxy/internal/client"
	"githib.com/A-UNDERSCORE-D/goRRProxy/internal/server"
)

var (
	isServer         bool
	clientFile       string
	apiKey           string
	validAPIKeysFile string
	bindAddr         string
	bindPort         string
)

func init() {
	flag.BoolVar(&isServer, "server", false, "runs as a server. Servers use clients to send messages")
	flag.StringVar(&clientFile, "clients", "./clients", "sets the file containing a client list to use (server only)")
	flag.StringVar(&apiKey, "apikey", "test", "sets the API key to use when contacting clients (server only)")
	flag.StringVar(&validAPIKeysFile, "validAPIKeys", "./apiKeys", "sets the API keys that this client will accept proxy requests from")
	flag.StringVar(&bindAddr, "bind-addr", "127.0.0.1", "sets the address that the client or server will accept requests on")
	flag.StringVar(&bindPort, "bind-port", "1337", "sets the port on which the client and server will accept requests on")
}

func main() {
	flag.Parse()
	var err error
	if isServer {
		err = StartServer(bindAddr, bindPort)
	} else {
		err = StartClient(bindAddr, bindPort)
	}
	if err != nil {
		panic(err)
	}
}

func readFromFile(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(f)
	f.Close()
	if err != nil {
		return nil, err
	}

	return strings.Split(string(data), "\n"), nil
}

func StartServer(bindAddr, bindPort string) error {
	clients, err := readFromFile(clientFile)
	if err != nil {
		return err
	}

	return server.ListenForRequests(apiKey, clients, bindAddr, bindPort)
}

func StartClient(bindAddr, bindPort string) error {
	return client.ListenForConnections(bindAddr, bindPort)
}
