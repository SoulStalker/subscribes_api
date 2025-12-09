package config

import (
	"fmt"
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
	Port string `yaml:"port" env-default:"8080"`
	Mode string `yaml:"mode" env-default:"release"`
}

type DBConfig struct {
	Host               string `yaml:"host"`
	Port               int    `yaml:"port"`
	User               string `yaml:"user"`
	Password           string `yaml:"password"`
	DbName             string `yaml:"dbname"`
	SSLMode            string `yaml:"sslmode"`
	MaxConnections     int    `yaml:"max_conns" env-default:"10"`
	MaxIdleConnections int    `yaml:"max_idle_conns"`
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.DbName,
		c.SSLMode,
	)
}

type LogConfig struct {
	Level    string `yaml:"level" env-default:"info"`
	Encoding string `yaml:"encoding"`
}

// MustLoad - загружает конфигурацию из yaml файла
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
