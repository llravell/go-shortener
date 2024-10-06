package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/llravell/go-shortener/internal/api"
	"github.com/llravell/go-shortener/internal/config"
	"github.com/llravell/go-shortener/internal/models"
	"github.com/llravell/go-shortener/internal/storages"
)

func runServer(c *config.Config) error {
	us := storages.NewUrlStorage()
	rsg := models.NewRandomStringGenerator()

	r := api.BuildRouter(us, rsg, c.BaseAddr)

	fmt.Printf("Server has started on %s\n", c.Addr)
	return http.ListenAndServe(c.Addr, r)
}

func main() {
	c := &config.Config{}
	p := config.NewArgsParser(c)
	p.Parse()

	log.Fatal(runServer(c))
}
