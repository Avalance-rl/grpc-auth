package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yaml:"env" env-default:"prod"`
	GRPC        GRPCConfig
	StoragePath string `yaml:"storage_path" env-default:"./storage"`
}

type GRPCConfig struct {
	Port      int           `yaml:"port"`
	Timeout   time.Duration `yaml:"timeout"`
	TokenTTL  time.Duration `yaml:"token_ttl" env-default:"1h"`
	SecretKey string        `yaml:"secret_key"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config file path is empty")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file not found")
	}
	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string
	res = os.Getenv("CONFIG_PATH")
	return res
}
