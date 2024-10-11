package config

import (
	"flag"

	"github.com/caarlos0/env"
)

type Parser interface {
	Parse() error
}

type argsParser struct {
	c *Config
}

var _ Parser = (*argsParser)(nil)

func NewArgsParser(c *Config) argsParser {
	p := argsParser{c}

	flag.StringVar(&p.c.Addr, "a", defaultAddr, "Server address as host:port")
	flag.StringVar(&p.c.BaseAddr, "b", defaultBaseAddr, "Base address for redirect as host:port")

	return p
}

func (p argsParser) Parse() error {
	flag.Parse()
	return nil
}

type envParser struct {
	c *Config
}

var _ Parser = (*envParser)(nil)

func NewEnvParser(c *Config) envParser {
	return envParser{c}
}

func (p envParser) Parse() error {
	return env.Parse(p.c)
}

type unionParser struct {
	parsers []Parser
}

var _ Parser = (*unionParser)(nil)

func (up *unionParser) Parse() error {
	for _, p := range up.parsers {
		err := p.Parse()

		if err != nil {
			return err
		}
	}

	return nil
}

func MakeParseStrategy(parsers ...Parser) *unionParser {
	return &unionParser{parsers}
}
