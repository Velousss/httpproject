package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type User struct {
	First string `json:"first"`
	Last  string `json:"last"`
}

func handlePostUser(w http.ResponseWriter, r *http.Request) {
	defer func(r io.ReadCloser) {
		_, _ = io.Copy(ioutil.Discard, r)
		_ = r.Close()
	}(r.Body)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var u User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func main() {
	http.HandleFunc("/user", handlePostUser)

	fmt.Println("Starting server on port 80")
	if err := http.ListenAndServe(":80", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
