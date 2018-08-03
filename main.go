package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/walesey/go-fileserver/client"
	"github.com/walesey/go-fileserver/server"
)

var serverAddr string

func main() {
	isServer := flag.Bool("server", false, "Server Mode")
	host := flag.String("host", "localhost", "Hostname of server")
	port := flag.Int64("port", 11862, "Port to listen on")
	remotePath := flag.String("path", ".", "Remote path to download")
	localpath := flag.String("out", ".", "Local destination path")
	quiet := flag.Bool("quiet", false, "run in quiet mode")
	flag.Parse()

	if *isServer {
		log.Println("Running Server")
		server.NewServer(*remotePath).Start(int(*port))
	} else {
		log.Println("Running Client")
		c := client.New(*localpath, fmt.Sprintf("http://%v:%v", *host, *port))
		c.Quiet = *quiet
		if err := c.SyncFiles(*remotePath); err != nil {
			log.Fatal(err)
		}
	}
}
