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

// Server ...
type Server struct {
	BasePath string
	Quiet    bool
	manifest files.FileItems
}

// NewServer ...
func NewServer(basePath string) *Server {
	return &Server{
		BasePath: basePath,
	}
}

// Start - start the file server
func (s *Server) Start(port int) error {
	// pre calculate the file manifest
	var err error
	if s.manifest, err = files.GetFileItems(s.BasePath); err != nil {
		return err
	}

	router := http.NewServeMux()
	router.HandleFunc("/", s.mainRoute)
	router.HandleFunc("/files", s.filesRoute)
	router.HandleFunc("/download", s.downloadRoute)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: router,
	}

	log.Printf("Listening on port: %v", port)
	return httpServer.ListenAndServe()
}

func (s *Server) mainRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		s.writeMessage(w, http.StatusOK, "Go File Server")
	}
}

func (s *Server) filesRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		path := r.URL.Query().Get("path")
		if !s.Quiet {
			log.Printf("getting files for path='%v'\n", path)
		}

		pathParts := strings.Split(path, "/")
		files := s.manifest
		if path != "" {
			for _, part := range pathParts {
				if part == "." {
					continue
				}
				if file, ok := files[part]; ok {
					files = file.Items
				} else {
					s.writeMessage(w, http.StatusNotFound, "not found")
					return
				}
			}
		}
		s.writeMessage(w, http.StatusOK, files)
	}
}

func (s *Server) downloadRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
		length, _ := strconv.ParseInt(r.URL.Query().Get("length"), 10, 64)
		path := s.parsePath(r)

		if !s.Quiet {
			log.Printf("serving file - path='%v'\n", path)
		}

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
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(status)
			w.Write(data)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}
	}
}
