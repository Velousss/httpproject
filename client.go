package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type User struct {
	First string `json:"first"`
	Last  string `json:"last"`
}

func postUser(url string, user User) error {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(&user)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("expected status %d; actual status %d", http.StatusAccepted, resp.StatusCode)
	}
	return nil
}

func postMultipartForm(url string, formFields map[string]string, files []string) error {
	reqBody := new(bytes.Buffer)
	w := multipart.NewWriter(reqBody)

	for k, v := range formFields {
		err := w.WriteField(k, v)
		if err != nil {
			return err
		}
	}

	for i, file := range files {
		filePart, err := w.CreateFormFile(fmt.Sprintf("file%d", i+1), filepath.Base(file))
		if err != nil {
			return err
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(filePart, f)
		if err != nil {
			return err
		}
	}

	err := w.Close()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status %d; actual status %d", http.StatusOK, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("Response: %s\n", body)
	return nil
}

func main() {
	serverURL := "http://localhost:80/user"
	user := User{First: "Marvel", Last: "Cokro"}

	err := postUser(serverURL, user)
	if err != nil {
		fmt.Println("Error posting user:", err)
		return
	}
	fmt.Println("User posted successfully")

	formFields := map[string]string{
		"date": time.Now().Format(time.RFC3339),
	}
	files := []string{
		"./files/tes.txt",
		"./files/anjay.txt",
	}

	err = postMultipartForm("https://httpbin.org/post", formFields, files)
	if err != nil {
		fmt.Println("Error posting multipart form:", err)
		return
	}
	fmt.Println("Multipart form posted successfully")
}
