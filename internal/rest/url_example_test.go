package rest_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/llravell/go-shortener/internal/rest"
)

func ExampleURLRoutes_common_case() {
	resp, err := http.Post("http://localhost:8080/", "text/plain", strings.NewReader("https://foo.ru"))
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	buf, _ := io.ReadAll(resp.Body)
	shortURL := string(buf)

	redirResp, err := http.Get("http://localhost:8080/" + shortURL)
	if err != nil {
		log.Fatal(err)
	}

	redirResp.Body.Close()
}

func ExampleURLRoutes_json_save() {
	resp, err := http.Post(
		"http://localhost:8080/api/shorten/",
		"application/json",
		strings.NewReader(`
			"url": "https://foo.ru"
		`),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	response := make(map[string]string)
	json.NewDecoder(resp.Body).Decode(&response)

	redirResp, err := http.Get(response["result"])
	if err != nil {
		log.Fatal(err)
	}

	redirResp.Body.Close()
}

func ExampleURLRoutes_batch() {
	batchItems := []rest.URLBatchRequestItem{
		{
			CorrelationID: "1",
			OriginalURL:   "https://foo.ru",
		},
		{
			CorrelationID: "2",
			OriginalURL:   "https://bar.ru",
		},
	}

	buf := &bytes.Buffer{}
	json.NewEncoder(buf).Encode(batchItems)

	resp, err := http.Post(
		"http://localhost:8080/api/shorten/batch",
		"application/json",
		buf,
	)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	var responseItems []*rest.URLBatchResponseItem

	json.NewDecoder(resp.Body).Decode(&responseItems)

	for _, item := range responseItems {
		redirResp, err := http.Get(item.ShortURL)

		if err != nil {
			log.Fatal(err)
		}

		redirResp.Body.Close()
	}
}
