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

func buildConfig() (*config.Config, error) {
	c := config.WithDefaultValues()
	p := config.MakeParseStrategy(
		config.NewArgsParser(c),
		config.NewEnvParser(c),
	)

	err := p.Parse()
	return c, err
}

func runServer(c *config.Config) error {
	us := storages.NewURLStorage()
	rsg := models.NewRandomStringGenerator()

	r := api.BuildRouter(us, rsg, c.BaseAddr)

	fmt.Printf("Running server on %s\n", c.Addr)
	return http.ListenAndServe(c.Addr, r)
}

func main() {
	c, err := buildConfig()
	if err != nil {
		panic(err)
	}

	log.Fatal(runServer(c))
}
