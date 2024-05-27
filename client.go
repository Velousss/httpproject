package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"main/data"
	"main/handler"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

func waitServer(url string, duration time.Duration) bool {
	deadline := time.Now().Add(duration)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		handler.HandleError(err)
		if resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return false
		}
	}
	return true
}

func main() {

	if waitServer("http://localhost:9876", 5*time.Second) {
		fmt.Println("Server not found")
		return
	}
	var choice int
	for {
		fmt.Println("Main Menu")
		fmt.Println("1. Get message")
		fmt.Println("2. Send file")
		fmt.Println("3. Quit")
		fmt.Print(">> ")
		fmt.Scanf("%d\n", &choice)
		if choice == 1 {
			getMessage()
		} else if choice == 2 {
			sendFile()
		} else if choice == 3 {
			break
		} else {
			fmt.Println("Invalid choice")
		}
	}
}

func getMessage() {
	resp, err := http.Get("http://localhost:9876")
	handler.HandleError(err)
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	handler.HandleError(err)

	fmt.Println("Server:", string(data))
}

func sendFile() {
	var name string
	var age int

	scanner := bufio.NewReader(os.Stdin)
	fmt.Print("Input name: ")
	name, _ = scanner.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Input age: ")
	fmt.Scanf("%d\n", &age)

	person := data.Person{
		Name: name,
		Age:  age,
	}

	jsonData, err := json.Marshal(person)
	handler.HandleError(err)

	temp := new(bytes.Buffer)
	w := multipart.NewWriter(temp)

	personField, err := w.CreateFormField("Person")
	handler.HandleError(err)

	_, err = personField.Write(jsonData)
	handler.HandleError(err)

	file, err := os.Open("./file.txt")
	handler.HandleError(err)
	defer file.Close()

	fileField, err := w.CreateFormFile("file", file.Name())
	handler.HandleError(err)

	_, err = io.Copy(fileField, file)
	handler.HandleError(err)

	err = w.Close()
	handler.HandleError(err)

	req, err := http.NewRequest("POST", "http://localhost:9876/sendFile", temp)
	handler.HandleError(err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	handler.HandleError(err)
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	handler.HandleError(err)

	fmt.Println("Server:", string(data))
}
