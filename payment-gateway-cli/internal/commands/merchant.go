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
	cmd.AddCommand(newMerchantGetCommand())
	//cmd.AddCommand(newMerchantListCommand())
	cmd.AddCommand(newMerchantInviteCommand())
	cmd.AddCommand(accessMerchantAccounts())
	cmd.AddCommand(GetTeamCommands())
	cmd.AddCommand(GetInviteCommands())
	cmd.AddCommand(GetSettingCommands())

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

			if err := config.SetMerchantID(merchant.ID); err != nil {
				return err
			}

			ui.Success("✅ Merchant created!")
			ui.Info(fmt.Sprintf("🆔 ID: %s", merchant.ID))
			ui.Info(fmt.Sprintf("📧 Email: %s", merchant.Email))
			ui.Info(fmt.Sprintf("🏪 Business Name: %s", merchant.BusinessName))
			ui.Info(fmt.Sprintf("🏢 Business Type: %s", merchant.BusinessType))
			ui.Info(fmt.Sprintf("🔑 Status: %s", merchant.Status))
			ui.Info(fmt.Sprintf("👤 Owner ID: %s", merchant.OwnerID))

			ui.Info("\n💡 Next: payment-cli apikey create ")

			return nil
		},
	}

	cmd.Flags().StringVar(&BusinessName, "business-name", "", "merchant business name")
	cmd.Flags().StringVar(&LegalName, "legal-name", "", "merchant legal name")
	cmd.Flags().StringVar(&email, "email", "", "merchant email")

	return cmd
}

func newMerchantGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Get merchant details",
		RunE: func(cmd *cobra.Command, args []string) error {

			if config.GetAccessToken() == "" {
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

			spinner := ui.NewSpinner("Fetching merchant...")
			spinner.Start()

			merchantClient := client.NewMerchantClient()
			merchant, err := merchantClient.GetMerchant(merchantID)

			spinner.Stop()

			if err != nil {
				ui.Error(fmt.Sprintf("❌ Failed: %v", err))
				return err
			}

			ui.Info(fmt.Sprintf("🏪 ID: %s", merchant.ID))
			ui.Info(fmt.Sprintf("📧 Email: %s", merchant.Email))
			ui.Info(fmt.Sprintf("🏪 Business Name: %s", merchant.BusinessName))
			ui.Info(fmt.Sprintf("👤 Legal Name: %s", merchant.LegalName))
			ui.Info(fmt.Sprintf("🏢 Business Type: %s", merchant.BusinessType))
			ui.Info(fmt.Sprintf("🔑 Status: %s", merchant.Status))
			ui.Info(fmt.Sprintf("🌍 Country Code: %s", merchant.CountryCode))
			ui.Info(fmt.Sprintf("💵 Currency Code: %s", merchant.CurrencyCode))
			ui.Info(fmt.Sprintf("👤 Owner ID: %s", merchant.OwnerID))
			ui.Info(fmt.Sprintf("🔑 Merchant Code: %s", merchant.MerchantCode))

			return nil
		},
	}
}

func newMerchantInviteCommand() *cobra.Command {
	var email, rolename, roleID string
	return &cobra.Command{
		Use:   "invite",
		Short: "Invite a user to the merchant",
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.GetAccessToken() == "" {
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
			ui.Info("please run ' payment-cli roles view ' to get role name and id")

			if email == "" {
				prompt := promptui.Prompt{Label: "Email"}
				result, err := prompt.Run()
				if err != nil {
					return err
				}
				email = result
			}
			if rolename == "" {
				prompt := promptui.Prompt{Label: "Role Name"}
				result, err := prompt.Run()
				if err != nil {
					return err
				}
				rolename = result
			}
			if roleID == "" {
				prompt := promptui.Prompt{Label: "Role ID"}
				result, err := prompt.Run()
				if err != nil {
					return err
				}
				roleID = result
			}

			spinner := ui.NewSpinner("Fetching invitations...")
			spinner.Start()

			merchantClient := client.NewMerchantClient()
			invitation, err := merchantClient.InviteUser(merchantID, email, rolename, roleID)

			spinner.Stop()

			if err != nil {
				ui.Error(fmt.Sprintf("❌ Failed: %v", err))
				return err
			}
			ui.Info(fmt.Sprintf("📧 Email: %s", invitation.Email))
			ui.Info(fmt.Sprintf("🏪 Role Name: %s", invitation.RoleName))
			ui.Info(fmt.Sprintf("🔑 Status: %s", invitation.Status))
			ui.Info(fmt.Sprintf("🔑 Invitation Token: %s", invitation.InvitationToken))
			ui.Info(fmt.Sprintf("🕒 Expires At: %s", invitation.ExpiresAt))
			ui.Info(fmt.Sprintf("📅 Created At: %s", invitation.CreatedAt))

			return nil
		},
	}
}

func accessMerchantAccounts() *cobra.Command {
	var MerchantID string
	cmd := &cobra.Command{
		Use:   "access-accounts",
		Short: "access merchant account",
		RunE: func(cmd *cobra.Command, args []string) error {
			//check for login
			if config.GetUserEmail() == "" && config.GetAccessToken() == "" {
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
			MerchantID = merchants[0].ID
			if err := config.SetMerchantID(MerchantID); err != nil {
				return err
			}
			ui.Success("✅ Merchant account access granted!")

			return nil
		},
	}
	cmd.Flags().StringVarP(&MerchantID, "merchant-id", "m", "", "Merchant ID")

	return cmd
}

func GetTeamCommands() *cobra.Command {
	var MerchantID string
	cmd := &cobra.Command{
		Use:   "team",
		Short: "List team members",
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.GetAccessToken() == "" {
				ui.Warning("⚠️  Not logged in")
				ui.Info("Run: payment-cli auth login")
				return nil
			}
			MerchantID = config.GetMerchantID()
			if MerchantID == "" {
				ui.Warning("⚠️  Merchant ID not set")
				ui.Info("Set it with: payment-cli merchant create")
				return nil
			}
			spinner := ui.NewSpinner("Fetching team members...")
			spinner.Start()

			merchantClient := client.NewMerchantClient()
			teamMembers, err := merchantClient.ListTeamMembers(MerchantID)

			spinner.Stop()

			if err != nil {
				ui.Error(fmt.Sprintf("❌ Failed: %v", err))
				return err
			}

			if len(teamMembers) == 0 {
				ui.Info("📭 No team members found")
				return nil
			}

			for _, member := range teamMembers {
				ui.Info(fmt.Sprintf("👤 ID: %s", member.UserID))
				ui.Info(fmt.Sprintf("🏪 Role Name: %s", member.RoleName))
				ui.Info(fmt.Sprintf("🔑 Status: %s", member.Status))
				ui.Info(fmt.Sprintf("🕒 Joined At: %s", member.JoinedAt.Time))
				ui.Info("------------------------------")
			}

			return nil
		},
	}

	return cmd
}

func GetInviteCommands() *cobra.Command {
	var MerchantID string
	cmd := &cobra.Command{
		Use:   "invitations",
		Short: "List invitations",
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.GetAccessToken() == "" {
				ui.Warning("⚠️  Not logged in")
				ui.Info("Run: payment-cli auth login")
				return nil
			}
			MerchantID = config.GetMerchantID()
			if MerchantID == "" {
				ui.Warning("⚠️  Merchant ID not set")
				ui.Info("Set it with: payment-cli merchant create")
				return nil
			}
			spinner := ui.NewSpinner("Fetching invitations...")
			spinner.Start()

			merchantClient := client.NewMerchantClient()
			invitations, err := merchantClient.ListInvitations(MerchantID)

			spinner.Stop()

			if err != nil {
				ui.Error(fmt.Sprintf("❌ Failed: %v", err))
				return err
			}

			if len(invitations) == 0 {
				ui.Info("� No invitations found")
				return nil
			}

			for _, inv := range invitations {
				ui.Info(fmt.Sprintf("📧 Email: %s", inv.Email))
				ui.Info(fmt.Sprintf("🏪 Role Name: %s", inv.RoleName))
				ui.Info(fmt.Sprintf("🔑 Status: %s", inv.Status))
				ui.Info(fmt.Sprintf("🕒 Expires At: %s", inv.ExpiresAt))
				ui.Info(fmt.Sprintf("🔑 Token: %s", inv.InvitationToken))
				ui.Info("------------------------------")
			}
			return nil
		},
	}
	return cmd
}

func GetSettingCommands() *cobra.Command {
	var MerchantID string
	cmd := &cobra.Command{
		Use:   "setting",
		Short: "Manage merchant settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.GetAccessToken() == "" {
				ui.Warning("⚠️  Not logged in")
				ui.Info("Run: payment-cli auth login")
				return nil
			}
			MerchantID = config.GetMerchantID()
			if MerchantID == "" {
				ui.Warning("⚠️  Merchant ID not set")
				ui.Info("Set it with: payment-cli merchant create")
				return nil
			}
			spinner := ui.NewSpinner("Fetching settings...")
			spinner.Start()

			merchantClient := client.NewMerchantClient()
			settings, err := merchantClient.GetSettings(MerchantID)

			spinner.Stop()

			if err != nil {
				ui.Error(fmt.Sprintf("❌ Failed: %v", err))
				return err
			}

			ui.Info(fmt.Sprintf("💵 Default Currency: %s", settings.DefaultCurrency))
			ui.Info(fmt.Sprintf("📝 Statement Descriptor: %s", settings.StatementDescriptor.String))
			ui.Info(fmt.Sprintf("📧 Notification Email: %s", settings.NotificationEmail.String))
			ui.Info(fmt.Sprintf("📨 Send Email Receipts: %v", settings.SendEmailReceipts))
			ui.Info(fmt.Sprintf("💰 Auto Settle: %v", settings.AutoSettle))
			ui.Info(fmt.Sprintf("📅 Settle Schedule: %s", settings.SettleSchedule))

			return nil
		},
	}

	return cmd
}
