package main

import (
	"os"
	"redis-crud/app"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"
)

func main() {
	server := &app.Server{}

	if err := server.Initialize(&redis.Options{
		Addr: "redis:6379", Password: "", DB: 0,
	}); err != nil {
		panic(err)
	}

	server.InitializeRoutes()

	port := strings.TrimSpace(os.Getenv("SERVER_PORT"))
	if _, err := strconv.Atoi(port); err != nil {
		panic("port address should be a number")
	}

	port = ":" + port
	server.Run(port)
}
