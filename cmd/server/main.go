package main

import (
	"github.com/minisource/notifire/api"
	"github.com/minisource/notifire/config"
)

func main() {
	cfg := config.GetConfig()

	api.InitServer(cfg)
}
