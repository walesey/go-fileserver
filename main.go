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
	port := flag.Int64("port", 3000, "Port to listen on")
	path := flag.String("path", ".", "Remote path to download")
	localpath := flag.String("out", ".", "Local destination path")
	flag.Parse()

	if *isServer { // server
		server.NewServer(*path).Start(int(*port))

	} else { // client
		c := client.NewClient(*localpath, fmt.Sprintf("http://%v:%v", *host, *port))
		inProgress := true
		go func() {
			if err := c.SyncFiles(*path); err != nil {
				log.Println(err)
			}
			inProgress = false
		}()

		completed := 0
		for inProgress {
			complete := <-c.Complete
			completed++
			fmt.Printf("%v/%v --> %v\n", completed, c.TotalFiles, complete)
		}
	}
}
