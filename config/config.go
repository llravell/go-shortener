package config

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/caarlos0/env"
)

const (
	_defaultAddr            = ":8080"
	_defaultBaseAddr        = "http://localhost:8080"
	_defaultFileStoragePath = "./urls.backup"
	_defaultJWTSecret       = "secret"
)

// Config конфигурация приложения.
type Config struct {
	Addr            string     `env:"SERVER_ADDRESS"    json:"server_address"`
	BaseAddr        string     `env:"BASE_URL"          json:"base_url"`
	FileStoragePath string     `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DatabaseDsn     string     `env:"DATABASE_DSN"      json:"database_dsn"`
	TrustedSubnet   string     `env:"TRUSTED_SUBNET"    json:"trusted_subnet"`
	HTTPSEnabled    bool       `env:"ENABLE_HTTPS"      json:"enable_https"`
	JWTSecret       string     `env:"JWT_SECRET"        json:"-"`
	AppEnv          string     `env:"APP_ENV"           json:"-"`
	Meta            configMeta `json:"-"`
}

type configMeta struct {
	SRC string `env:"CONFIG" json:"-"`
}

func newDefaultConfig() *Config {
	return &Config{
		Addr:            _defaultAddr,
		BaseAddr:        _defaultBaseAddr,
		FileStoragePath: _defaultFileStoragePath,
		JWTSecret:       _defaultJWTSecret,
	}
}

// NewConfig создает новый конфиг, заполняет его значениями из переменных окружения и флагов.
func NewConfig() (*Config, error) {
	cfg := newDefaultConfig()
	cfg.parseFromFlags()

	if err := cfg.parseFromEnv(); err != nil {
		return nil, err
	}

	if len(cfg.Meta.SRC) == 0 {
		return cfg, nil
	}

	cfgFromFile := newDefaultConfig()
	if err := cfgFromFile.parseFromFile(); err != nil {
		return nil, err
	}

	cfgFromFile.merge(cfg)

	return cfgFromFile, nil
}

func (cfg *Config) parseFromFlags() {
	flag.StringVar(&cfg.Meta.SRC, "config", "", "Path to config file")

	flag.StringVar(&cfg.Addr, "a", _defaultAddr, "Server address as host:port")
	flag.StringVar(&cfg.BaseAddr, "b", _defaultBaseAddr, "Base address for redirect as host:port")
	flag.StringVar(&cfg.FileStoragePath, "f", _defaultFileStoragePath, "File storage path")
	flag.StringVar(&cfg.DatabaseDsn, "d", "", "DB connect address")
	flag.StringVar(&cfg.TrustedSubnet, "t", "", "Trusted subnet for /api/internal/ routes")
	flag.BoolVar(&cfg.HTTPSEnabled, "s", false, "Enable https")
	flag.Parse()
}

func (cfg *Config) parseFromEnv() error {
	return env.Parse(cfg)
}

func (cfg *Config) parseFromFile() error {
	buf, err := os.ReadFile(cfg.Meta.SRC)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, cfg)
}

func (cfg *Config) merge(target *Config) {
	if len(target.Addr) != 0 {
		cfg.Addr = target.Addr
	}

	if len(target.BaseAddr) != 0 {
		cfg.BaseAddr = target.BaseAddr
	}

	if len(target.FileStoragePath) != 0 {
		cfg.FileStoragePath = target.FileStoragePath
	}

	if len(target.DatabaseDsn) != 0 {
		cfg.DatabaseDsn = target.DatabaseDsn
	}

	if len(target.TrustedSubnet) != 0 {
		cfg.TrustedSubnet = target.TrustedSubnet
	}

	if target.HTTPSEnabled {
		cfg.HTTPSEnabled = target.HTTPSEnabled
	}

	if len(target.JWTSecret) != 0 {
		cfg.JWTSecret = target.JWTSecret
	}

	if len(target.AppEnv) != 0 {
		cfg.AppEnv = target.AppEnv
	}

	if len(target.Meta.SRC) != 0 {
		cfg.Meta.SRC = target.Meta.SRC
	}
}
