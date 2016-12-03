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

func (c *Client) SyncFiles(path string) error {
	query := url.Values{}
	query.Set("path", filepath.ToSlash(path))
	resp, err := http.Get(fmt.Sprint(c.serverAddr, "/files?", query.Encode()))
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

	return c.syncFile(localFiles, remoteFiles, ".", path)
}

func (c *Client) syncFile(localFiles, remoteFiles files.FileItems, path, remotepath string) error {
	for name, file := range remoteFiles {
		var err error
		newPath := filepath.Join(path, name)
		newRemotePath := filepath.Join(remotepath, name)
		if file.Directory {
			if localFile, ok := localFiles[name]; ok {
				err = c.syncFile(localFile.Items, file.Items, newPath, newRemotePath)
			} else {
				os.Mkdir(filepath.Join(c.basePath, newPath), 0777)
				err = c.syncFile(make(map[string]files.FileItem), file.Items, newPath, newRemotePath)
			}
		} else {
			if localFile, ok := localFiles[name]; !ok || localFile.Hash != file.Hash {
				err = c.downloadFile(newPath, newRemotePath, file)
			}

			select { // Don't block when channel is full
			case c.Complete <- name:
			default:
			}
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) downloadFile(path, remotepath string, file files.FileItem) error {
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
		query.Set("path", filepath.ToSlash(remotepath))
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
