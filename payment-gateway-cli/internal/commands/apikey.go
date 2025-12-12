package commands

import (
	"github.com/rhaloubi/payment-gateway-cli/internal/ui"
	"github.com/spf13/cobra"
)

func NewAPIKeyCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apikey",
		Short: "🔑 API key management",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "create",
		Short: "Create API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.Info("🔑 API key create command")
			ui.Warning("⚠️  Coming soon!")
			return nil
		},
	})

	return cmd
}
