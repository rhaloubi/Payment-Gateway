package inits

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/rhaloubi/payment-gateway/payment-api-service/config"
)

var RDB *redis.Client
var Ctx = context.Background()

func InitRedis() {
	dsn := config.GetEnv("REDIS_DSN")

	opt, err := redis.ParseURL(dsn)
	if err != nil {
		panic(err)
	}

	RDB = redis.NewClient(opt)

	_, err = RDB.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("✅ Connected to Redis successfully")
}
