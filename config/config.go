package config

import (
	"flag"

	"github.com/caarlos0/env"
)

const (
	_defaultAddr            = ":8080"
	_defaultBaseAddr        = "http://localhost:8080"
	_defaultFileStoragePath = "./urls.backup"
	_defaultDatabaseDsn     = ""
	_defaultJWTSecret       = "secret"
)

// Config конфигурация приложения.
type Config struct {
	Addr            string `env:"SERVER_ADDRESS"`
	BaseAddr        string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDsn     string `env:"DATABASE_DSN"`
	JWTSecret       string `env:"JWT_SECRET"`
	AppEnv          string `env:"APP_ENV"`
	HTTPSEnabled    bool   `env:"ENABLE_HTTPS"`
}

// NewConfig создает новый конфиг, заполняет его значениями из переменных окружения и флагов.
func NewConfig() (*Config, error) {
	cfg := &Config{
		Addr:            _defaultAddr,
		BaseAddr:        _defaultBaseAddr,
		FileStoragePath: _defaultFileStoragePath,
		DatabaseDsn:     _defaultDatabaseDsn,
		JWTSecret:       _defaultJWTSecret,
	}

	flag.StringVar(&cfg.Addr, "a", _defaultAddr, "Server address as host:port")
	flag.StringVar(&cfg.BaseAddr, "b", _defaultBaseAddr, "Base address for redirect as host:port")
	flag.StringVar(&cfg.FileStoragePath, "f", _defaultFileStoragePath, "File storage path")
	flag.StringVar(&cfg.DatabaseDsn, "d", _defaultDatabaseDsn, "DB connect address")
	flag.BoolVar(&cfg.HTTPSEnabled, "c", false, "Enable https")
	flag.Parse()

	return cfg, env.Parse(cfg)
}
