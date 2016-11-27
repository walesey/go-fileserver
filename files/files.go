package files

import (
	"os"
	"path/filepath"
)

type FileItems []FileItem

type FileItem struct {
	Name  string    `json:"name"`
	Hash  string    `json:"hash"`
	Items FileItems `json:"items"`
}

func AllFiles(path string) (FileItems, error) {
	fileNames, err := filepath.Glob(filepath.Join(path, "*"))
	if err != nil {
		return []FileItem{}, err
	}

	fileItems := make([]FileItem, len(fileNames))
	for i, name := range fileNames {
		fPath := filepath.Join(path, name)
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
		if fi.Mode().IsDir() {
			if items, err = AllFiles(fPath); err != nil {
				return fileItems, err
			}
		} else {
			items = []FileItem{}
			// md5.Sum() //TODO
		}
		fileItems[i] = FileItem{Name: name, Hash: hash, Items: items}
	}
	return fileItems, nil
}
