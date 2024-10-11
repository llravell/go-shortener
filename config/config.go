package config

import (
	"flag"

	"github.com/caarlos0/env"
)

const defaultAddr = ":8080"
const defaultBaseAddr = "http://localhost:8080"

type Config struct {
	Addr     string `env:"SERVER_ADDRESS"`
	BaseAddr string `env:"BASE_URL"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{
		Addr:     defaultAddr,
		BaseAddr: defaultBaseAddr,
	}

	flag.StringVar(&cfg.Addr, "a", defaultAddr, "Server address as host:port")
	flag.StringVar(&cfg.BaseAddr, "b", defaultBaseAddr, "Base address for redirect as host:port")
	flag.Parse()

	env.Parse(cfg)

	return cfg, nil
}
