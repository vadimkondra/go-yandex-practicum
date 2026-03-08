package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {

	url := initUrl()

	fmt.Println("Sending request to:", url)

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, nil)

	if err != nil {
		panic(err)
	}

	response, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("Status:", response.Status)
	fmt.Println("Body:", string(body))

	defer response.Body.Close()
}

func initUrl() string {
	serverHost := "http://localhost:8080"

	return serverHost + "/update/counter/testCounter/100"
}
