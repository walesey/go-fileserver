package main

import (
	"fmt"
	"os"

	"net/http"

	"io/ioutil"

	"encoding/json"

	"github.com/walesey/go-fileserver/files"
	"github.com/walesey/go-fileserver/server"
)

func main() {
	isServer := false
	if len(os.Args) >= 2 {
		isServer = os.Args[1] == "server"
	}

	if isServer {
		server.StartServer()
	} else {
		syncFiles("http://localhost:3000", ".")
	}
}

func syncFiles(serverAddr, basePath string) {
	resp, err := http.Get(fmt.Sprint(serverAddr, "/files"))
	if err != nil {
		panic(err)
	}

	filesData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var remoteFiles files.FileItems
	if err = json.Unmarshal(filesData, &filesData); err != nil {
		panic(err)
	}

	localFiles, err := files.AllFiles(basePath)
	if err != nil {
		panic(err)
	}

	syncFile(localFiles, remoteFiles)
}

func syncFile(localFiles, remoteFiles files.FileItems) {
	for _, file := range remoteFiles {
		if file.Directory {

		} else {

		}
	}
}
