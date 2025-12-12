package commands

import (
	"fmt"

	"github.com/rhaloubi/payment-gateway-cli/internal/config"
	"github.com/rhaloubi/payment-gateway-cli/internal/ui"
	"github.com/spf13/cobra"
)

func NewConfigCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "⚙️  Configuration management",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.Info(fmt.Sprintf("🌍 Environment: %s", config.GetCurrentEnv()))
			ui.Info(fmt.Sprintf("🔗 API URL: %s", config.GetAPIURL()))
			ui.Info(fmt.Sprintf("📁 Config file: %s", config.GetConfigPath()))
			return nil
		},
	})

	return cmd
}
