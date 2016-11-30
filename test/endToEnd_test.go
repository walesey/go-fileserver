package test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
	"github.com/walesey/go-fileserver/client"
	"github.com/walesey/go-fileserver/server"
)

const testDir = "./testData"
const resultDir = "./testResult"

type fileHash struct {
	path, content string
}

func TestMain(m *testing.M) {
	//setup
	os.RemoveAll(resultDir)
	os.Mkdir(resultDir, 0777)
	go server.NewServer(testDir).Start(3000)

	time.Sleep(1 * time.Second)

	//run tests
	ret := m.Run()

	//cleanup
	os.RemoveAll(resultDir)

	//exit
	os.Exit(ret)
}

func summarizeFiles(dirPath string) ([]fileHash, error) {
	files := []fileHash{}
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(dirPath, path)
		if info == nil || info.IsDir() {
			files = append(files, fileHash{path: relPath})
		} else {
			fileData, _ := ioutil.ReadFile(path)
			files = append(files, fileHash{path: relPath, content: string(fileData[:])})
		}
		return nil
	})
	return files, err
}

func TestFileSync(t *testing.T) {
	c := client.NewClient(resultDir, "http://127.0.0.1:3000")
	err := c.SyncFiles()
	assert.Nil(t, err)

	// get expected content from the test directory
	expectedFiles, err := summarizeFiles(testDir)
	assert.Nil(t, err)

	// // check the expected content against the destination directory
	actualFiles, err := summarizeFiles(resultDir)
	assert.Nil(t, err)

	assert.EqualValues(t, expectedFiles, actualFiles, "Equal files")
}
