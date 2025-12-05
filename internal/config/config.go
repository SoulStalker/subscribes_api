package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	DB     DBConfig     `yaml:"database"`
	Log    LogConfig    `yaml:"log"`
}

type ServerConfig struct {
	Port int    `yaml:"port" env-default:"8080"`
	Mode string `yaml:"mode" env-default:"release"`
}

type DBConfig struct {
	Host               string `yaml:"host"`
	Port               int    `yaml:"port"`
	User               string `yaml:"user"`
	Password           string `yaml:"password"`
	DbName             string `yaml:"dbname"`
	SSLMode            string `yaml:"sslmode"`
	MaxConnections     int    `yaml:"25"`
	MaxIdleConnections int    `yaml:"5"`
}

type LogConfig struct {
	Level    string `yaml:"level" env-default:"info"`
	Encoding string `yaml:"encoding"`
}

func MustLoad(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exits: %s", configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %v", err)
	}

	return &cfg
}
