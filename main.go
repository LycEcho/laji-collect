package main

import (
	"lajiCollect/app/bootstrap"
	"lajiCollect/config"
)

func main() {
	b := bootstrap.New(config.ServerConfig.Port, config.ServerConfig.LogLevel)
	b.Serve()
}


