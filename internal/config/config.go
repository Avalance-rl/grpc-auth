package config

import (
	"os"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config структура для инициализации конфига
type Config struct {
	Env         string `yaml:"env" env-default:"prod"`
	GRPC        GRPCConfig
	StoragePath string `yaml:"storage_path" env-default:"./storage"`
}

var (
	configInstance *Config
	once           sync.Once // once механизм синхронизации
)

type GRPCConfig struct {
	Port      int           `yaml:"port"`
	Timeout   time.Duration `yaml:"timeout"`
	TokenTTL  time.Duration `yaml:"token_ttl" env-default:"1h"`
	SecretKey string        `yaml:"secret_key"`
}

// GetConfig функция для получения синглтона конфига
func GetConfig() *Config {
	once.Do(func() {
		configInstance = &Config{}
		configInstance.mustLoad()
	})
	return configInstance
}

// mustLoad считывает конфиг файл
func (c *Config) mustLoad() {
	path := fetchConfigPath()
	if path == "" {
		panic("config file path is empty")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file not found")
	}

	if err := cleanenv.ReadConfig(path, configInstance); err != nil {
		panic("failed to read config: " + err.Error())
	}

}

func fetchConfigPath() string {
	var res string
	res = os.Getenv("CONFIG_PATH")
	return res
}
