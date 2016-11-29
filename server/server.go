package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"strconv"

	"github.com/gorilla/handlers"
	"github.com/walesey/go-fileserver/files"
)

const BASE_PATH = "."

func StartServer() {
	router := http.NewServeMux()
	router.HandleFunc("/", mainRoute)
	router.HandleFunc("/files", filesRoute)
	router.HandleFunc("/download", downloadRoute)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3000"
	}

	s := &http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: handlers.LoggingHandler(os.Stdout, router),
	}

	log.Printf("Listening on port: %v", port)
	log.Fatal(s.ListenAndServe())
}

func mainRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		writeMessage(w, http.StatusOK, "Go File Server")
	}
}

func filesRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if files, err := files.AllFiles(BASE_PATH); err == nil {
			writeMessage(w, http.StatusOK, files)
		} else {
			log.Println(err)
			writeMessage(w, http.StatusInternalServerError, "Internal Server Error")
		}
	}
}

func downloadRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		path := r.URL.Query().Get("path")
		offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
		length, _ := strconv.ParseInt(r.URL.Query().Get("length"), 10, 64)

		file, _ := os.Open(filepath.Join(BASE_PATH, path))
		defer file.Close()
		file.Seek(offset, os.SEEK_SET)
		n, _ := io.CopyN(w, file, length)
		w.Header().Set("Content-Length", strconv.FormatInt(n, 10))
		w.Header().Set("Content-Type", "application/octet-stream")
	}
}

func writeMessage(w http.ResponseWriter, status int, message interface{}) {
	switch t := message.(type) {
	case string:
		w.WriteHeader(status)
		w.Write([]byte(t))
	case *string:
		w.WriteHeader(status)
		w.Write([]byte(*t))
	case []byte:
		w.WriteHeader(status)
		w.Write(t)
	default:
		if data, err := json.Marshal(message); err == nil {
			w.WriteHeader(status)
			w.Write(data)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}
	}
}
