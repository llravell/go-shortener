package rest_test

import (
	"fmt"
	"log"
	"net/http"
)

func ExampleHealthRoutes() {
	response, err := http.Get("http://localhost:8080/ping")
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	fmt.Println(response.StatusCode)
}
