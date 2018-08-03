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
	"strings"

	"github.com/walesey/go-fileserver/files"
)

type Server struct {
	BasePath string
}

func NewServer(basePath string) *Server {
	return &Server{
		BasePath: basePath,
	}
}

func (s *Server) Start(port int) {
	router := http.NewServeMux()
	router.HandleFunc("/", s.mainRoute)
	router.HandleFunc("/files", s.filesRoute)
	router.HandleFunc("/download", s.downloadRoute)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: router,
	}

	log.Printf("Listening on port: %v", port)
	log.Fatal(httpServer.ListenAndServe())
}

func (s *Server) mainRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		s.writeMessage(w, http.StatusOK, "Go File Server")
	}
}

func (s *Server) filesRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		path := s.parsePath(r)
		if files, err := files.GetFileItems(path); err == nil {
			s.writeMessage(w, http.StatusOK, files)
		} else {
			log.Println(err)
			s.writeMessage(w, http.StatusInternalServerError, "Internal Server Error")
		}
	}
}

func (s *Server) downloadRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
		length, _ := strconv.ParseInt(r.URL.Query().Get("length"), 10, 64)
		path := s.parsePath(r)

		file, _ := os.Open(path)
		defer file.Close()
		file.Seek(offset, os.SEEK_SET)
		n, _ := io.CopyN(w, file, length)
		w.Header().Set("Content-Length", strconv.FormatInt(n, 10))
		w.Header().Set("Content-Type", "application/octet-stream")
	}
}

func (s *Server) parsePath(r *http.Request) string {
	path := r.URL.Query().Get("path")
	path = strings.Replace(path, "..", ".", -1)
	path = filepath.Join(s.BasePath, path)
	return path
}

func (s *Server) writeMessage(w http.ResponseWriter, status int, message interface{}) {
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
