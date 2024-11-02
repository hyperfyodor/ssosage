package migrator

import (
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	MigrationsPath  string `json:"migrations_path" env-required:"true"`
	StoragePath     string `json:"storage_path" env-required:"true"`
	MigrationsTable string `json:"migrations_table" env-default:"migrations"`
}

func MustLoad(configPath string) *Config {
	if configPath == "" {
		panic("config path is empty")
	}

	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("config path is empty: " + err.Error())
	}

	return &cfg
}
