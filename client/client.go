package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"strconv"

	"github.com/walesey/go-fileserver/files"
)

type Client struct {
	basePath   string
	serverAddr string
	ChunkSize  int
	TotalFiles int
	Complete   chan string
}

func NewClient(basePath, serverAddr string) *Client {
	return &Client{
		basePath:   basePath,
		serverAddr: serverAddr,
		ChunkSize:  100000,
		Complete:   make(chan string, 32),
	}
}

func (c *Client) SyncFiles() error {
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
	c.TotalFiles = remoteFiles.Count()

	localFiles, err := files.AllFiles(c.basePath)
	if err != nil {
		return err
	}

	return c.syncFile(localFiles, remoteFiles, ".")
}

func (c *Client) syncFile(localFiles, remoteFiles files.FileItems, path string) error {
	for name, file := range remoteFiles {
		var err error
		newPath := filepath.Join(path, name)
		if file.Directory {
			if localFile, ok := localFiles[name]; ok {
				err = c.syncFile(localFile.Items, file.Items, newPath)
			} else {
				os.Mkdir(filepath.Join(c.basePath, newPath), 0777)
				err = c.syncFile(make(map[string]files.FileItem), file.Items, newPath)
			}
		} else {
			if localFile, ok := localFiles[name]; !ok || localFile.Hash != file.Hash {
				err = c.downloadFile(newPath, file)
			}
			c.Complete <- name
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) downloadFile(path string, file files.FileItem) error {
	localPath := filepath.Join(c.basePath, path)
	if _, err := os.Stat(localPath); os.IsExist(err) {
		os.Remove(localPath)
	}

	f, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	for offset := 0; offset < file.Size; offset += c.ChunkSize {
		query := url.Values{}
		query.Set("path", path)
		query.Set("offset", strconv.Itoa(offset))
		query.Set("length", strconv.Itoa(c.ChunkSize))
		resp, err := http.Get(fmt.Sprintf("%v/download?%v", c.serverAddr, query.Encode()))
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
