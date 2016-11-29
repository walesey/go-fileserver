package client

import (
	"encoding/json"
	"fmt"
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

func (c *Client) SyncFiles(basePath string) error {
	resp, err := http.Get(fmt.Sprint(c.serverAddr, "/files"))
	if err != nil {
		return err
	}

	filesData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var remoteFiles files.FileItems
	if err = json.Unmarshal(filesData, &remoteFiles); err != nil {
		return err
	}

	localFiles, err := files.AllFiles(basePath)
	if err != nil {
		return err
	}

	return c.syncFile(localFiles, remoteFiles, basePath)
}

func (c *Client) syncFile(localFiles, remoteFiles files.FileItems, path string) error {
	for name, file := range remoteFiles {
		var err error
		newPath := filepath.Join(path, name)
		if file.Directory {
			if localFile, ok := localFiles[name]; ok {
				err = c.syncFile(localFile.Items, file.Items, newPath)
			} else {
				os.Mkdir(filepath.Join(path, name), 0777)
				err = c.syncFile(make(map[string]files.FileItem), file.Items, newPath)
			}
		} else {
			if localFile, ok := localFiles[name]; !ok || localFile.Hash != file.Hash {
				err = c.downloadFile(newPath, file)
			}
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) downloadFile(path string, file files.FileItem) error {
	if _, err := os.Stat(path); os.IsExist(err) {
		os.Remove(path)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	for offset := 0; offset < file.Size; offset += c.ChunkSize {
		resp, err := http.Get(fmt.Sprintf("%v/download?path=%v&offset=%v&length=%v", c.serverAddr, path, offset, c.ChunkSize))
		if err != nil {
			return err
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if _, err := f.Write(data); err != nil {
			return err
		}
	}
	return nil
}
