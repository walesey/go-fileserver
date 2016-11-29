# Go File Server

This is a file server that allows clients to sync a filesystem with the server

[![Build Status](https://travis-ci.org/walesey/go-fileserver.svg?branch=master)](https://travis-ci.org/walesey/go-fileserver)

Usage:

``` go
import (
	"github.com/walesey/go-fileserver/client"
	"github.com/walesey/go-fileserver/server"
)

// start server
	server.NewServer(".").Start(3000)

// sync directory with server
c := client.NewClient("http://localhost:3000")
if err := c.SyncFiles("."); err != nil {
	log.Println(err)
}

```