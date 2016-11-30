package test

import (
	"os"

	"github.com/walesey/go-fileserver/server"
)

func RunTestServer(testDir, resultDir string) {
	os.RemoveAll(resultDir)
	os.Mkdir(resultDir, 0777)
	go server.NewServer(testDir).Start(3000)
}
