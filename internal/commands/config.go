package commands

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/rhaloubi/payment-gateway-cli/internal/config"
	"github.com/rhaloubi/payment-gateway-cli/internal/ui"
	"github.com/spf13/cobra"
)

func NewConfigCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "⚙️  Configuration management",
		Long:  "Manage CLI configuration settings, environments, and preferences",
	}

	cmd.AddCommand(newConfigShowCommand())
	cmd.AddCommand(newConfigSetCommand())
	cmd.AddCommand(newConfigUseCommand())
	cmd.AddCommand(newConfigResetCommand())

	return cmd
}

func newConfigShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  "Display all current configuration settings including environment, URLs, and preferences",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.Info("📋 Current Configuration")
			ui.Info("═══════════════════════════════════════")
			ui.Info("")

			ui.Info("🌍 Environment Settings:")
			ui.Info(fmt.Sprintf("  Current Environment: %s", config.GetCurrentEnv()))
			ui.Info(fmt.Sprintf("  API URL:            %s", config.GetAPIURL()))
			ui.Info(fmt.Sprintf("  Auth URL:           %s", config.GetAuthURL()))
			ui.Info(fmt.Sprintf("  Payment URL:        %s", config.GetPaymentURL()))
			ui.Info("")

			ui.Info("🔐 Credentials:")
			if config.GetAccessToken() != "" {
				ui.Info(fmt.Sprintf("  Access Token:       %s...%s",
					config.GetAccessToken()[:10],
					config.GetAccessToken()[len(config.GetAccessToken())-10:]))
			} else {
				ui.Warning("  Access Token:       Not set")
			}
			if config.GetUserEmail() != "" {
				ui.Info(fmt.Sprintf("  User Email:         %s", config.GetUserEmail()))
			} else {
				ui.Warning("  User Email:         Not logged in")
			}
			if config.GetMerchantID() != "" {
				ui.Info(fmt.Sprintf("  Merchant ID:        %s", config.GetMerchantID()))
			} else {
				ui.Warning("  Merchant ID:        Not set")
			}
			if config.GetApiKey() != "" {
				ui.Info(fmt.Sprintf("  API Key:            %s...%s",
					config.GetApiKey()[:10],
					config.GetApiKey()[len(config.GetApiKey())-10:]))
			} else {
				ui.Warning("  API Key:            Not set")
			}
			ui.Info("")

			ui.Info("⚙️  Preferences:")
			ui.Info(fmt.Sprintf("  Output Format:      %s", config.GetOutputFormat()))
			ui.Info(fmt.Sprintf("  Color Enabled:      %v", config.GetColorEnabled()))
			ui.Info(fmt.Sprintf("  Debug Mode:         %v", config.GetDebugMode()))
			ui.Info("")

			ui.Info(fmt.Sprintf("📁 Config File:        %s", config.GetConfigPath()))

			return nil
		},
	}
}

func newConfigSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long: `Set a specific configuration value.

Available keys:
  output_format    Output format (table, json, yaml)
  color_enabled    Enable/disable colors (true, false)
  debug_mode       Enable/disable debug mode (true, false)

Examples:
  payment-cli config set output_format json
  payment-cli config set color_enabled false
  payment-cli config set debug_mode true`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			// Validate values
			switch key {
			case "output_format":
				if value != "table" && value != "json" && value != "yaml" {
					return fmt.Errorf("invalid output_format: must be 'table', 'json', or 'yaml'")
				}
			case "color_enabled", "debug_mode":
				if value != "true" && value != "false" {
					return fmt.Errorf("invalid value for %s: must be 'true' or 'false'", key)
				}
			default:
				return fmt.Errorf("unknown config key: %s", key)
			}

			if err := config.SetConfigValue(key, value); err != nil {
				ui.Error(fmt.Sprintf("❌ Failed to set config: %v", err))
				return err
			}

			ui.Success(fmt.Sprintf("✅ Set %s = %s", key, value))
			return nil
		},
	}

	return cmd
}

func newConfigUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use <environment>",
		Short: "Switch environment",
		Long: `Switch between different environments (production, development).

Examples:
  payment-cli config use production
  payment-cli config use development`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var env string

			if len(args) == 0 {
				// Interactive selection
				prompt := promptui.Select{
					Label: "Select Environment",
					Items: []string{"production", "development"},
				}
				_, result, err := prompt.Run()
				if err != nil {
					return err
				}
				env = result
			} else {
				env = args[0]
			}

			// Validate environment
			if env != "production" && env != "development" {
				return fmt.Errorf("invalid environment: must be 'production' or 'development'")
			}

			spinner := ui.NewSpinner(fmt.Sprintf("Switching to %s...", env))
			spinner.Start()

			if err := config.SetCurrentEnv(env); err != nil {
				spinner.Stop()
				ui.Error(fmt.Sprintf("❌ Failed to switch environment: %v", err))
				return err
			}

			spinner.Stop()
			ui.Success(fmt.Sprintf("✅ Switched to %s environment", env))
			ui.Info(fmt.Sprintf("🔗 API URL: %s", config.GetAPIURL()))

			return nil
		},
	}

	return cmd
}

func newConfigResetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset configuration to defaults",
		Long:  "Reset all configuration settings to their default values. This will clear all credentials and preferences.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Confirm reset
			prompt := promptui.Prompt{
				Label:     "Are you sure you want to reset all configuration? This will clear credentials",
				IsConfirm: true,
			}

			_, err := prompt.Run()
			if err != nil {
				ui.Info("Reset cancelled")
				return nil
			}

			spinner := ui.NewSpinner("Resetting configuration...")
			spinner.Start()

			if err := config.ResetConfig(); err != nil {
				spinner.Stop()
				ui.Error(fmt.Sprintf("❌ Failed to reset config: %v", err))
				return err
			}

			spinner.Stop()
			ui.Success("✅ Configuration reset to defaults")
			ui.Info("🚀 Run 'payment-cli init' to reinitialize if needed")

			return nil
		},
	}

	return cmd
}
