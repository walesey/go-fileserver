package files

import (
	"crypto/md5"
	"encoding/base64"
	"io/ioutil"
	"os"
	"path/filepath"
)

type FileItems []FileItem

type FileItem struct {
	Name      string    `json:"name"`
	Size      int       `json:"size"`
	Hash      string    `json:"hash"`
	Directory bool      `json:"directory"`
	Items     FileItems `json:"items"`
}

func AllFiles(path string) (FileItems, error) {
	filePaths, err := filepath.Glob(filepath.Join(path, "*"))
	if err != nil {
		return []FileItem{}, err
	}

	fileItems := make([]FileItem, len(filePaths))
	for i, fPath := range filePaths {
		name := filepath.Base(fPath)
		f, err := os.Open(fPath)
		if err != nil {
			return fileItems, err
		}

		defer f.Close()
		fi, err := f.Stat()
		if err != nil {
			return fileItems, err
		}

		var items FileItems
		var hash string
		var size int

		isDirectory := fi.Mode().IsDir()
		if isDirectory {
			if items, err = AllFiles(fPath); err != nil {
				return fileItems, err
			}
		} else {
			items = []FileItem{}
			fileData, err := ioutil.ReadAll(f)
			if err != nil {
				return fileItems, err
			}

			size = len(fileData)
			hashData := md5.Sum(fileData)
			hash = base64.URLEncoding.EncodeToString(hashData[:])
		}

		fileItems[i] = FileItem{
			Name:      name,
			Hash:      hash,
			Size:      size,
			Directory: isDirectory,
			Items:     items,
		}
	}
	return fileItems, nil
}