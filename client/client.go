package client

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/walesey/go-fileserver/files"
)

type Client struct {
	localBasePath              string
	serverAddr                 string
	bytesDone, totalDeltaBytes int64
	filesDone, totalDeltaFiles int
}

func New(localBasePath, serverAddr string) Client {
	return Client{
		localBasePath: localBasePath,
		serverAddr:    serverAddr,
	}
}

func (c *Client) SyncFiles(path string) error {
	remoteFiles, err := c.getRemoteFiles(path)
	if err != nil {
		return err
	}

	localFiles, err := files.GetFileItems(c.localBasePath)
	if err != nil {
		return err
	}

	c.filesDone, c.totalDeltaFiles = 0, remoteFiles.Count()
	c.bytesDone, c.totalDeltaBytes = 0, remoteFiles.Size(localFiles)

	return c.syncFiles(localFiles, remoteFiles, c.localBasePath, path)
}

func (c *Client) syncFiles(localFiles, remoteFiles files.FileItems, localPath, remotePath string) error {
	for name, remoteFile := range remoteFiles {
		var err error
		newLocalPath := filepath.Join(localPath, name)
		newRemotePath := filepath.Join(remotePath, name)
		localFile, localFileExists := localFiles[name]
		if remoteFile.Directory {
			if !localFileExists {
				os.Mkdir(newLocalPath, 0777)
				localFile = files.FileItem{Items: make(map[string]files.FileItem)}
			}
			if err = c.syncFiles(localFile.Items, remoteFile.Items, newLocalPath, newRemotePath); err != nil {
				return err
			}
		} else {
			if !localFileExists || localFile.Hash != remoteFile.Hash {
				if err = c.downloadFile(newLocalPath, newRemotePath, localFile, remoteFile); err != nil {
					return errors.Wrap(err, "unable to download file")
				}
			}

			c.filesDone++
			log.Infof("%v/%v --> %v\n", c.filesDone, c.totalDeltaFiles, name)
		}
	}
	return nil
}

func (c *Client) downloadFile(localPath, remotePath string, localFile, remoteFile files.FileItem) error {
	_, err := os.Stat(localPath)
	deltaFound := !os.IsExist(err)

	f, err := os.Create(localPath)
	if err != nil {
		return errors.Wrap(err, "unable to create file locally")
	}
	defer f.Close()

	offset := int64(0)
	var buf []byte
	for _, chunk := range remoteFile.Chunks {
		if !deltaFound {
			if int64(len(buf)) != chunk.Size {
				buf = make([]byte, chunk.Size)
			}

			n, err := f.Read(buf)
			if err != nil {
				return errors.Wrap(err, "unable to read file locally")
			}

			hashData := md5.Sum(buf[:n])
			localHash := base64.URLEncoding.EncodeToString(hashData[:])
			deltaFound = chunk.Hash != localHash
			if deltaFound {
				if _, err := f.Seek(offset, 0); err != nil {
					return errors.Wrap(err, "unable to seek file locally")
				}
			} else {
				continue
			}
		}

		query := url.Values{}
		query.Set("path", filepath.ToSlash(remotePath))
		query.Set("offset", fmt.Sprint(offset))
		query.Set("length", fmt.Sprint(chunk.Size))
		resp, err := http.Get(fmt.Sprintf("%v/download?%v", c.serverAddr, query.Encode()))
		if err != nil {
			return errors.Wrap(err, "unable to download file remotely")
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "unable to read http response")
		}

		if _, err := f.Write(data); err != nil {
			return errors.Wrap(err, "unable to write to local file")
		}
		f.Sync()

		c.bytesDone += chunk.Size
		log.Infof("%v/%v\n", formatBytes(c.bytesDone), formatBytes(c.totalDeltaBytes))

		offset += chunk.Size
	}

	return nil
}

func (c *Client) getRemoteFiles(path string) (files.FileItems, error) {
	query := url.Values{}
	query.Set("path", filepath.ToSlash(path))
	resp, err := http.Get(fmt.Sprint(c.serverAddr, "/files?", query.Encode()))
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch /files")
	}

	if resp.StatusCode >= 400 {
		return nil, errors.Errorf("unable to read to /files http body (%v)", resp.StatusCode)
	}

	filesData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read to /files http body")
	}

	var remoteFiles files.FileItems
	if err = json.Unmarshal(filesData, &remoteFiles); err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal /files http body")
	}

	return remoteFiles, nil
}

const (
	terabyte = 1000000000000
	gigabyte = 1000000000
	megabyte = 1000000
	kilobyte = 1000
)

func formatBytes(nbBytes int64) string {
	if nbBytes >= terabyte*10 {
		return fmt.Sprintf("%vTb", nbBytes/terabyte)
	}
	if nbBytes >= gigabyte*10 {
		return fmt.Sprintf("%vGb", nbBytes/gigabyte)
	}
	if nbBytes >= megabyte*10 {
		return fmt.Sprintf("%vMb", nbBytes/megabyte)
	}
	if nbBytes >= kilobyte*10 {
		return fmt.Sprintf("%vKb", nbBytes/kilobyte)
	}
	return fmt.Sprintf("%vb", nbBytes)
}
