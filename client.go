package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
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

func loadCertPool(certFile string) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()
	certData, err := os.ReadFile(certFile)
	if err != nil {
		return nil, err
	}
	if !certPool.AppendCertsFromPEM(certData) {
		return nil, fmt.Errorf("failed to append certificate")
	}
	return certPool, nil
}

func printTLSDetails(connState tls.ConnectionState) {
	fmt.Printf("TLS Version: %s\n", tlsVersionToString(connState.Version))
	fmt.Printf("CipherSuite: %s\n", tlsCipherSuiteToString(connState.CipherSuite))
	if len(connState.PeerCertificates) > 0 {
		issuer := connState.PeerCertificates[0].Issuer
		fmt.Printf("Issuer Organization: %s\n", issuer.Organization)
	}
}

func tlsVersionToString(version uint16) string {
	switch version {
	case tls.VersionTLS13:
		return "TLS 1.3"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS10:
		return "TLS 1.0"
	default:
		return "Unknown"
	}
}

func tlsCipherSuiteToString(cipherSuite uint16) string {
	switch cipherSuite {
	case tls.TLS_AES_128_GCM_SHA256:
		return "TLS_AES_128_GCM_SHA256"
	case tls.TLS_AES_256_GCM_SHA384:
		return "TLS_AES_256_GCM_SHA384"
	case tls.TLS_CHACHA20_POLY1305_SHA256:
		return "TLS_CHACHA20_POLY1305_SHA256"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:
		return "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:
		return "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"
	case tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:
		return "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
	case tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:
		return "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
	case tls.TLS_RSA_WITH_AES_128_GCM_SHA256:
		return "TLS_RSA_WITH_AES_128_GCM_SHA256"
	case tls.TLS_RSA_WITH_AES_256_GCM_SHA384:
		return "TLS_RSA_WITH_AES_256_GCM_SHA384"
	default:
		return "Unknown"
	}
}

func waitServer(url string, duration time.Duration, client *http.Client) bool {
	deadline := time.Now().Add(duration)
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		handler.HandleError(err)
		if resp != nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return false
		}
	}
	return true
}

func main() {
	certPool, err := loadCertPool("./certificate/cert.pem")
	handler.HandleError(err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,
			},
		},
	}

	if waitServer("https://localhost:9876", 5*time.Second, client) {
		fmt.Println("Server not found")
		return
	}
	var choice int
	for {
		fmt.Println("Main Menu")
		fmt.Println("1. Get message")
		fmt.Println("2. Send file")
		fmt.Println("3. Print TLS details")
		fmt.Println("4. Quit")
		fmt.Print(">> ")
		fmt.Scanf("%d\n", &choice)
		if choice == 1 {
			getMessage(client)
		} else if choice == 2 {
			sendFile(client)
		} else if choice == 3 {
			printTLSInfo(client)
		} else if choice == 4 {
			break
		} else {
			fmt.Println("Invalid choice")
		}
	}
}

func getMessage(client *http.Client) {
	resp, err := client.Get("https://localhost:9876")
	handler.HandleError(err)
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	handler.HandleError(err)

	fmt.Println("Server:", string(data))
}

func sendFile(client *http.Client) {
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

	req, err := http.NewRequest("POST", "https://localhost:9876/sendFile", temp)
	handler.HandleError(err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := client.Do(req)
	handler.HandleError(err)
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	handler.HandleError(err)

	fmt.Println("Server:", string(data))
}

func printTLSInfo(client *http.Client) {
	resp, err := client.Get("https://localhost:9876")
	handler.HandleError(err)
	defer resp.Body.Close()

	connState := resp.TLS
	if connState != nil {
		printTLSDetails(*connState)
	} else {
		fmt.Println("No TLS connection state found")
	}
}
