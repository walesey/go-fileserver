package main

import (
	"log"
	"os"

	"github.com/walesey/go-fileserver/client"
	"github.com/walesey/go-fileserver/server"
)

var serverAddr string

func main() {
	isServer := false
	if len(os.Args) >= 2 {
		isServer = os.Args[1] == "server"
	}

	if isServer {
		server.StartServer()
	} else {
		c := client.NewClient("http://localhost:3000")
		if err := c.SyncFiles("."); err != nil {
			log.Println(err)
		}
	}
}
