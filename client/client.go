package client

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/walesey/go-fileserver/files"
)

type Client struct {
	serverAddr string
	ChunkSize  int
}

func NewClient(serverAddr string) *Client {
	return &Client{
		serverAddr: serverAddr,
		ChunkSize:  100000,
	}
}

func (c *Client) SyncFiles(basePath string) {
	resp, err := http.Get(fmt.Sprint(c.serverAddr, "/files"))
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

	c.syncFile(localFiles, remoteFiles, basePath)
}

func (c *Client) syncFile(localFiles, remoteFiles files.FileItems, path string) {
	for name, file := range remoteFiles {
		newPath := filepath.Join(path, name)
		if file.Directory {
			if localFile, ok := localFiles[name]; ok {
				c.syncFile(file.Items, localFile.Items, newPath)
			} else {
				os.Mkdir(filepath.Join(path, name), 0777)
				c.syncFile(file.Items, make(map[string]files.FileItem), newPath)
			}
		} else {
			if localFile, ok := localFiles[name]; !ok || localFile.Hash != file.Hash {
				c.downloadFile(newPath, file)
			}
		}
	}
}

func (c *Client) downloadFile(path string, file files.FileItem) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	for offset := 0; offset < file.Size; offset += c.ChunkSize {
		resp, err := http.Get(fmt.Sprintf("%v/download?path=%v&offset=%v&limit=%v", c.serverAddr, path, offset, c.ChunkSize))
		if err != nil {
			panic(err)
		}

		_, err = io.CopyN(f, resp.Body, int64(c.ChunkSize))
		if err != nil {
			panic(err)
		}

	}
}
