package commands

import (
	"github.com/rhaloubi/payment-gateway-cli/internal/ui"
	"github.com/spf13/cobra"
)

func NewInteractiveCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "interactive",
		Aliases: []string{"i"},
		Short:   "🎮 Interactive mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.Info("🎮 Interactive mode")
			ui.Warning("⚠️  Coming in v1.1!")
			return nil
		},
	}
}
