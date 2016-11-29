package test

import (
	"crypto/md5"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/walesey/go-fileserver/client"
	"github.com/walesey/go-fileserver/server"
)

const testDir = "./test/testData"
const resultDir = "./test/testResult"

type fileHash struct {
	path, hash string
}

func runSync() error {
	os.RemoveAll(resultDir)
	os.Mkdir(resultDir, 0777)
	go server.NewServer(testDir).Start(3000)
	c := client.NewClient("http://127.0.0.1:3000")
	return c.SyncFiles(resultDir)
}

func summarizeFiles(dirPath string) []fileHash {
	files := []fileHash{}
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() {
			files = append(files, fileHash{path: path})
		} else {
			fileData, _ := ioutil.ReadFile(path)
			hashData := md5.Sum(fileData)
			files = append(files, fileHash{path: path, hash: string(hashData[:])})
		}
		return nil
	})
	return files
}

func TestFileSync(t *testing.T) {
	err := runSync()
	assert.Nil(t, err)

	// get expected content from the test directory
	expectedFiles := summarizeFiles(testDir)

	// // check the expected content against the destination directory
	actualFiles := summarizeFiles(resultDir)

	assert.EqualValues(t, expectedFiles, actualFiles, "Equal files")
}
