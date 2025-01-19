package httpm_test

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/kmio11/go-httpm"
)

func executeRequest(client *http.Client, method, url string) {
	defer fmt.Println("---------------------")

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error:", r)
		}
	}()

	fmt.Println("---------------------")
	fmt.Println("Request:", method, url)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Status Code:", resp.StatusCode)
	fmt.Println("Response Body:", string(body))
}

func Example() {
	testdataDir := "testdata/Example"

	transport, err := httpm.NewTransportFromConfigFile(filepath.Join(testdataDir, "config.toml"))
	if err != nil {
		panic(err)
	}
	client := &http.Client{
		Transport: transport,
	}

	// executeRequest is a helper function that creates a request and send it by client.Do.
	executeRequest(client, http.MethodGet, "https://example.com/get")
	executeRequest(client, http.MethodPost, "https://example.com/submit")
	executeRequest(client, http.MethodPut, "https://example.com/submit")
	executeRequest(client, http.MethodPut, "https://example.com/panic")

	// Output:
	// ---------------------
	// Request: GET https://example.com/get
	// Status Code: 200
	// Response Body: This is a mock response for GET https://example.com.
	// ---------------------
	// ---------------------
	// Request: POST https://example.com/submit
	// Status Code: 200
	// Response Body: This is a mock response for https://example.com/submit
	// ---------------------
	// ---------------------
	// Request: PUT https://example.com/submit
	// Status Code: 200
	// Response Body: This is a mock response for https://example.com/submit
	// ---------------------
	// ---------------------
	// Request: PUT https://example.com/panic
	// Error: [mock] request caused a panic. method:PUT url:https://example.com/panic
	// ---------------------
}
