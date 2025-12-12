package commands

import (
	"github.com/rhaloubi/payment-gateway-cli/internal/ui"
	"github.com/spf13/cobra"
)

func NewPaymentCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payment",
		Short: "💳 Payment operations",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "authorize",
		Short: "Authorize a payment",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.Info("💳 Payment authorize command")
			ui.Warning("⚠️  Coming soon!")
			return nil
		},
	})

	return cmd
}
