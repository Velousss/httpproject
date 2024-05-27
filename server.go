package main

import (
	"encoding/json"
	"fmt"
	"io"
	"main/data"
	"main/handler"
	"net/http"
)

func middleware(method string, handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handlerFunc(w, r)
	}
}

func getMessageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello")
}

func sendFileHandler(w http.ResponseWriter, r *http.Request) {
	jsonData := r.FormValue("Person")
	var person data.Person
	err := json.Unmarshal([]byte(jsonData), &person)
	handler.HandleError(err)
	fmt.Println("JSON:", person)

	file, header, err := r.FormFile("file")
	handler.HandleError(err)
	fmt.Printf("Recieved File: %s\n", header.Filename)

	fileContent, err := io.ReadAll(file)
	handler.HandleError(err)

	fmt.Printf("File content: \n%s\n", fileContent)
	fmt.Fprintln(w, "Successfully received data")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", middleware(http.MethodGet, getMessageHandler))
	mux.HandleFunc("/sendFile", middleware(http.MethodPost, sendFileHandler))

	server := http.Server{
		Addr:    "localhost:9876",
		Handler: mux,
	}

	err := server.ListenAndServe()
	handler.HandleError(err)
}
