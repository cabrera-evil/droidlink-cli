package app

import "go-template/internal/config"

func Run() string {
	cfg := config.Load()
	return "Hello, " + cfg.Name + "!"
}
