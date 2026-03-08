package main

import (
	"net/http"
)

func main() {

	serverHost := "http://localhost:8080"

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, serverHost, nil)

	if err != nil {
		panic(err)
	}

	response, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	defer response.Body.Close()
}
