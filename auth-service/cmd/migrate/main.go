package main

import (
	"log"
	"os"

	"github.com/rhaloubi/payment-gateway/auth-service/inits"
	"github.com/rhaloubi/payment-gateway/auth-service/inits/logger"
	"github.com/rhaloubi/payment-gateway/auth-service/internal/migrations"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate [up|down]")
	}
	if os.Getenv("APP_MODE") == "" {
		inits.InitDotEnv()
	}
	logger.Init()
	inits.InitDB()

	switch os.Args[1] {
	case "up":
		log.Println("⬆ running migrations")
		if err := migrations.RunAuthMigrations(); err != nil {
			log.Fatal(err)
		}

	case "down":
		log.Println("⬇ rolling back migrations")
		if err := migrations.RollbackAuthMigrations(); err != nil {
			log.Fatal(err)
		}

	default:
		log.Fatalf("unknown command: %s", os.Args[1])
	}
}
