package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rhaloubi/payment-gateway/auth-service/inits"
	"github.com/rhaloubi/payment-gateway/auth-service/internal/api"
)

func init() {
	inits.InitDotEnv()
	inits.InitDB()
	inits.InitRedis()
	api.Routes()
}

func main() {
	// Create a channel to listen for interrupt (Ctrl+C) or system termination
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Run the API server in a goroutine so we can listen for shutdown
	go func() {
		if err := api.R.Run(); err != nil {
			fmt.Printf("❌ Server error: %v\n", err)
		}
	}()

	fmt.Println("✅ Server running... Press Ctrl+C to stop.")

	// Wait until a stop signal is received
	<-stop
	fmt.Println("\n🛑 Shutting down gracefully...")

	// ✅ Close Redis connection
	if err := inits.RDB.Close(); err != nil {
		fmt.Printf("❌ Error closing Redis: %v\n", err)
	} else {
		fmt.Println("🧹 Redis connection closed.")
	}

	fmt.Println("✅ Shutdown complete.")
}
