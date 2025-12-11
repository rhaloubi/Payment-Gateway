package commands

import (
	"fmt"

	"github.com/rhaloubi/payment-gateway-cli/internal/config"
	"github.com/rhaloubi/payment-gateway-cli/internal/ui"
	"github.com/spf13/cobra"
)

func NewInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "🔧 Initialize CLI configuration",
		Long:  "Creates configuration directory and default config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := ui.NewSpinner("Initializing CLI...")
			spinner.Start()

			if err := config.Init(); err != nil {
				spinner.Stop()
				return fmt.Errorf("failed to initialize: %w", err)
			}

			spinner.Stop()
			ui.Success("✅ CLI initialized successfully!")
			ui.Info(fmt.Sprintf("📁 Config file: %s", config.GetConfigPath()))
			ui.Info("\n🚀 Next steps:")
			ui.Info("  1. payment-cli register")
			ui.Info("  2. payment-cli login")
			ui.Info("  3. payment-cli merchant create")
			
			return nil
		},
	}
}
