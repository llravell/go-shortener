package config

const DefaultAddr = ":8080"
const DefaultBaseAddr = "http://localhost:8080"

type Config struct {
	Addr     string `env:"SERVER_ADDRESS"`
	BaseAddr string `env:"BASE_URL"`
}

func WithDefaultValues() *Config {
	return &Config{
		Addr:     DefaultAddr,
		BaseAddr: DefaultBaseAddr,
	}
}
