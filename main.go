package main

import (
	"os"

	"github.com/walesey/go-fileserver/server"
)

func main() {
	isServer := false
	if len(os.Args) >= 2 {
		isServer = os.Args[1] == "server"
	}

	if isServer {
		server.StartServer()
	}
}
