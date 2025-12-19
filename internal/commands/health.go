package commands

import (
	"github.com/rhaloubi/payment-gateway-cli/internal/ui"
	"github.com/spf13/cobra"
)

func NewHealthCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "❤️  Health check",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.Success("✅ CLI is working!")
			return nil
		},
	}
}
