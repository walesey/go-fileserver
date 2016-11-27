package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
)

func StartServer() {
	router := http.NewServeMux()
	router.HandleFunc("/", mainRoute)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "80"
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
