package main

import (
	"log"
	"os"
	"strconv"

	"github.com/walesey/go-fileserver/client"
	"github.com/walesey/go-fileserver/server"
)

var serverAddr string

func main() {
	isServer := false
	if len(os.Args) >= 2 {
		isServer = os.Args[1] == "server"
	}

	if isServer { // server
		var port int64 = 3000
		if len(os.Args) >= 3 {
			port, _ = strconv.ParseInt(os.Args[2], 10, 64)
		}

		path := "."
		if len(os.Args) >= 4 {
			path = os.Args[3]
		}

		server.NewServer(path).Start(int(port))

	} else { // client
		addr := "http://localhost:3000"
		if len(os.Args) >= 2 {
			addr = os.Args[1]
		}
		c := client.NewClient(addr)

		path := "."
		if len(os.Args) >= 3 {
			path = os.Args[2]
		}
		if err := c.SyncFiles(path); err != nil {
			log.Println(err)
		}
	}
}
