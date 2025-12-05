package main

import (
	"fmt"

	"github.com/SoulStalker/subscribes_api/internal/config"
)

func main() {
	cfgPath := "./config/config.yaml"
	cfg := config.MustLoad(cfgPath)

	fmt.Println(cfg.DB.DbName)
	fmt.Println(cfg.Log.Level)
	fmt.Println(cfg.Server.Port)
}
