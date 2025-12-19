package commands

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/rhaloubi/payment-gateway-cli/internal/client"
	"github.com/rhaloubi/payment-gateway-cli/internal/config"
	"github.com/rhaloubi/payment-gateway-cli/internal/ui"
	"github.com/spf13/cobra"
)

func NewAPIKeyCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apikey",
		Short: "🔑 API key management",
	}

	cmd.AddCommand(CreateAPIKeyCommands())
	cmd.AddCommand(StoreAPIKeyCommands())

	return cmd
}

func CreateAPIKeyCommands() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			//check for login
			if config.GetUserEmail() == "" && config.GetAccessToken() == "" {
				ui.Warning("⚠️  Not logged in")
				ui.Info("Run: payment-cli auth login")
				return nil
			}
			merchantID := config.GetMerchantID()
			if merchantID == "" {
				ui.Warning("⚠️  Merchant ID not set")
				ui.Info("Set it with: payment-cli merchant create")
				return nil
			}
			if name == "" {
				prompt := promptui.Prompt{Label: "API Key Name"}
				result, err := prompt.Run()
				if err != nil {
					return err
				}
				name = result
			}
			spinner := ui.NewSpinner("Fetching merchant...")
			spinner.Start()

			merchantClient := client.NewMerchantClient()
			res, err := merchantClient.CreateAPIKey(merchantID, name)

			spinner.Stop()

			if err != nil {
				ui.Error(fmt.Sprintf("❌ Failed: %v", err))
				return err
			}
			ui.Success("✅ API key created!")
			ui.Info(fmt.Sprintf("🔑 API Key: %s", res.APIKey.Name))
			ui.Error("Please copy the plain key, it will not be shown again")
			ui.Info(fmt.Sprintf("🔑 Plain Key: %s", res.PlainKey))
			return nil
		},
	}
	return cmd
}

func DeactivateAPIKeyCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deactivate",
		Short: "🔑 Deactivate API key",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "deactivate",
		Short: "Deactivate API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.Info("🔑 API key deactivate command")
			ui.Warning("⚠️  Coming soon!")
			return nil
		},
	})

	return cmd
}

func DeleteAPIKeyCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apikey",
		Short: "🔑 API key management",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "delete",
		Short: "Delete API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.Info("🔑 API key delete command")
			ui.Warning("⚠️  Coming soon!")
			return nil
		},
	})

	return cmd
}

func StoreAPIKeyCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store <plain_key>",
		Short: "🔑 Let store the plain key to use !!",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				ui.Error("Please provide the plain key")
				return nil
			}
			plainKey := args[0]
			config.SetApiKey(plainKey)
			ui.Success("✅ API key stored!")
			return nil
		},
	}
	return cmd
}
