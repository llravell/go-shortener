package config

import "flag"

type Config struct {
	Addr     string
	BaseAddr string
}

type argsParser struct {
	c *Config
}

func NewArgsParser(c *Config) argsParser {
	p := argsParser{c}

	flag.StringVar(&p.c.Addr, "a", ":8080", "Server address as host:port")
	flag.StringVar(&p.c.BaseAddr, "b", "http://localhost:8080", "Base address for redirect as host:port")

	return p
}

func (p argsParser) Parse() {
	flag.Parse()
}
