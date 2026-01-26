package main

import (
	"fmt"
	"os"

	"github.com/rhaloubi/payment-gateway-cli/internal/commands"
	"github.com/rhaloubi/payment-gateway-cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	version = "1.1.0"
	cfgFile string
	debug   bool
	output  string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "payment-cli",
		Short: "💳 Payment Gateway CLI",
		Long: `
╔═══════════════════════════════════════════════════════════════╗
║                  💳 Payment Gateway CLI v1.1.0                ║
║                                                               ║
║  Beautiful command-line tool for payment gateway management   ║
╚═══════════════════════════════════════════════════════════════╝

A developer-friendly CLI for managing merchants, testing payments,
and debugging your payment infrastructure.

New in v1.1.0:
  • Payment Intents with hosted checkout
  • Enhanced configuration management
  • Production-first default URLs
  • Real-time payment status polling

Examples:
  payment-cli init
  payment-cli auth register
  payment-cli auth login
  payment-cli merchant create
  payment-cli payment intent create
  payment-cli payment transactions
`,
		Version: version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if err := config.Load(cfgFile); err != nil {
				// Config doesn't exist yet, that's OK
			}
			if debug {
				config.SetDebug(true)
			}
			if output != "" {
				config.SetOutputFormat(output)
			}
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.payment-cli/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug mode")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "output format (table|json|yaml)")

	// Add commands
	rootCmd.AddCommand(commands.NewInitCommand())
	rootCmd.AddCommand(commands.NewAuthCommands())
	rootCmd.AddCommand(commands.NewMerchantCommands())
	rootCmd.AddCommand(commands.NewAPIKeyCommands())
	rootCmd.AddCommand(commands.NewPaymentCommands())
	rootCmd.AddCommand(commands.NewConfigCommands())
	rootCmd.AddCommand(commands.NewHealthCommand())
	rootCmd.AddCommand(commands.NewInteractiveCommand())
	rootCmd.AddCommand(commands.NewRolesCommands())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
}
