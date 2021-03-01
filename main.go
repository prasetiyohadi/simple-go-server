package main

import (
	"log"
	"net/http"
)

func main() {
	// Create the default multiplexer (mux)
	mux := http.NewServeMux()

	// Handle the root (/) path using function handler
	mux.HandleFunc("/", helloHandler)

	// Handle /v1/func using function handler
	mux.HandleFunc("/v1/func", funcHandler)

	// Handle /v1/type using type handler
	tHandler := typeHandler{}
	mux.Handle("/v1/type", tHandler)

	// Create the server
	s := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	err := s.ListenAndServe()
	if err != nil {
		log.Printf("Cannot start server: %s/n", err)
	}
}

func helloHandler(rw http.ResponseWriter, r *http.Request) {
	data := []byte("Hello, World!\n")
	rw.WriteHeader(200)
	_, err := rw.Write(data)
	if err != nil {
		log.Printf("Error writing response data: %s/n", err)
	}
}

func funcHandler(rw http.ResponseWriter, r *http.Request) {
	data := []byte("v1 of func's called.\n")
	rw.WriteHeader(200)
	_, err := rw.Write(data)
	if err != nil {
		log.Printf("Error writing response data: %s/n", err)
	}
}

type typeHandler struct{}

func (h typeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	data := []byte("v1 of type's called.\n")
	rw.WriteHeader(200)
	_, err := rw.Write(data)
	if err != nil {
		log.Printf("Error writing response data: %s/n", err)
	}
}
