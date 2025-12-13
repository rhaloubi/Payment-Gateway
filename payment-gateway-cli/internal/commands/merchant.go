package commands

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/rhaloubi/payment-gateway-cli/internal/client"
	"github.com/rhaloubi/payment-gateway-cli/internal/config"
	"github.com/rhaloubi/payment-gateway-cli/internal/ui"
	"github.com/spf13/cobra"
)

func NewMerchantCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "merchant",
		Short: "🏪 Merchant management",
	}

	cmd.AddCommand(newMerchantCreateCommand())
	//cmd.AddCommand(newMerchantListCommand())
	//cmd.AddCommand(newMerchantGetCommand())

	return cmd
}

func newMerchantCreateCommand() *cobra.Command {
	var BusinessName, email, LegalName, BusinessType string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new merchant",
		RunE: func(cmd *cobra.Command, args []string) error {
			//check for login
			if config.GetUserEmail() == "" && config.GetAccessToken() == "" {
				ui.Warning("⚠️  Not logged in")
				ui.Info("Run: payment-cli auth login")
				return nil
			}

			email = config.GetUserEmail()

			if BusinessName == "" {
				prompt := promptui.Prompt{Label: "Business Name"}
				result, err := prompt.Run()
				if err != nil {
					return err
				}
				BusinessName = result
			}
			if LegalName == "" {
				prompt := promptui.Prompt{Label: "Legal Name"}
				result, err := prompt.Run()
				if err != nil {
					return err
				}
				LegalName = result
			}
			if BusinessType == "" {
				ui.Info("all the business types: individual sole_proprietor partnership corporation non_profit ")
				ui.Info("choose one of them")
				prompt := promptui.Prompt{Label: "Business Type"}
				result, err := prompt.Run()
				if err != nil {
					return err
				}
				BusinessType = result
			}

			spinner := ui.NewSpinner("Creating merchant...")
			spinner.Start()

			merchantClient := client.NewMerchantClient()
			merchant, err := merchantClient.Create(BusinessName, LegalName, email, BusinessType)

			spinner.Stop()

			if err != nil {
				ui.Error(fmt.Sprintf("❌ Failed: %v", err))
				return err
			}

			ui.Success("✅ Merchant created!")
			ui.Info(fmt.Sprintf("🆔 ID: %s", merchant.ID))
			ui.Info(fmt.Sprintf("📧 Email: %s", merchant.Email))
			ui.Info(fmt.Sprintf("🏪 Business Name: %s", merchant.BusinessName))
			ui.Info(fmt.Sprintf("👤 Legal Name: %s", merchant.LegalName))
			ui.Info("\n💡 Next: payment-cli apikey create --merchant-id " + merchant.ID)

			return nil
		},
	}

	cmd.Flags().StringVar(&BusinessName, "business-name", "", "merchant business name")
	cmd.Flags().StringVar(&LegalName, "legal-name", "", "merchant legal name")
	cmd.Flags().StringVar(&email, "email", "", "merchant email")

	return cmd
}

/*
func newMerchantListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all merchants",
		RunE: func(cmd *cobra.Command, args []string) error {
			email := config.GetUserEmail()
			if email == "" && config.GetAccessToken() == "" {
				ui.Warning("⚠️  Not logged in")
				ui.Info("Run: payment-cli auth login")
				return nil
			}

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
*/
