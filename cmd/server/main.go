package main

import (
	"github.com/minisource/notifier/api"
	"github.com/minisource/notifier/config"
)

func main() {
	cfg := config.GetConfig()

	api.InitServer(cfg)
}
