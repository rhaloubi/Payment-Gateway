package commands

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/rhaloubi/payment-gateway-cli/internal/client"
	"github.com/rhaloubi/payment-gateway-cli/internal/ui"
	"github.com/spf13/cobra"
)

func NewMerchantCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "merchant",
		Short: "🏪 Merchant management",
	}

	cmd.AddCommand(newMerchantCreateCommand())
	cmd.AddCommand(newMerchantListCommand())
	cmd.AddCommand(newMerchantGetCommand())

	return cmd
}

func newMerchantCreateCommand() *cobra.Command {
	var name, email string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new merchant",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				prompt := promptui.Prompt{Label: "Merchant Name"}
				result, err := prompt.Run()
				if err != nil {
					return err
				}
				name = result
			}

			if email == "" {
				prompt := promptui.Prompt{Label: "Email"}
				result, err := prompt.Run()
				if err != nil {
					return err
				}
				email = result
			}

			spinner := ui.NewSpinner("Creating merchant...")
			spinner.Start()

			merchantClient := client.NewMerchantClient()
			merchant, err := merchantClient.Create(name, email)

			spinner.Stop()

			if err != nil {
				ui.Error(fmt.Sprintf("❌ Failed: %v", err))
				return err
			}

			ui.Success("✅ Merchant created!")
			ui.Info(fmt.Sprintf("🆔 ID: %s", merchant.ID))
			ui.Info(fmt.Sprintf("📧 Email: %s", merchant.Email))
			ui.Info("\n💡 Next: payment-cli apikey create --merchant-id " + merchant.ID)

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "merchant name")
	cmd.Flags().StringVar(&email, "email", "", "merchant email")

	return cmd
}

func newMerchantListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all merchants",
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := ui.NewSpinner("Fetching merchants...")
			spinner.Start()

			merchantClient := client.NewMerchantClient()
			merchants, err := merchantClient.List()

			spinner.Stop()

			if err != nil {
				ui.Error(fmt.Sprintf("❌ Failed: %v", err))
				return err
			}

			if len(merchants) == 0 {
				ui.Info("📭 No merchants found")
				ui.Info("Create one with: payment-cli merchant create")
				return nil
			}

			table := ui.NewTable([]string{"ID", "Name", "Email", "Status"})
			for _, m := range merchants {
				table.AddRow([]string{m.ID, m.Name, m.Email, m.Status})
			}
			table.Render()

			return nil
		},
	}
}

func newMerchantGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <merchant-id>",
		Short: "Get merchant details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			merchantID := args[0]

			spinner := ui.NewSpinner("Fetching merchant...")
			spinner.Start()

			merchantClient := client.NewMerchantClient()
			merchant, err := merchantClient.Get(merchantID)

			spinner.Stop()

			if err != nil {
				ui.Error(fmt.Sprintf("❌ Failed: %v", err))
				return err
			}

			ui.Info(fmt.Sprintf("🏪 Merchant: %s", merchant.Name))
			ui.Info(fmt.Sprintf("🆔 ID: %s", merchant.ID))
			ui.Info(fmt.Sprintf("📧 Email: %s", merchant.Email))
			ui.Info(fmt.Sprintf("📊 Status: %s", merchant.Status))

			return nil
		},
	}
}
