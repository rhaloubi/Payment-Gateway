package main

import (
	"log"
	"os"

	"github.com/rhaloubi/payment-gateway/tokenization-service/config"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits/logger"
	"github.com/rhaloubi/payment-gateway/tokenization-service/internal/migrations"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate [up|down]")
	}
	if config.GetEnv("APP_MODE") == "" {
		inits.InitDotEnv()
	}
	logger.Init()
	inits.InitDB()

	switch os.Args[1] {
	case "up":
		log.Println("⬆ running migrations")
		if err := migrations.RunMigrations(); err != nil {
			log.Fatal(err)
		}

	case "down":
		log.Println("⬇ rolling back migrations")
		if err := migrations.RollbackMigrations(); err != nil {
			log.Fatal(err)
		}

	default:
		log.Fatalf("unknown command: %s", os.Args[1])
	}
}
