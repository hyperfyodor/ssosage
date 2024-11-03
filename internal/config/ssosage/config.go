package ssosage

import (
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"
)

type Config struct {
	StoragePath string `json:"storage_path" env-required:"true"`
	GrpcPort    int    `json:"grpc_port" env-default:"3333"`
	Env         string `json:"env" env-default:"local"`
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
