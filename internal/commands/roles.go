package commands

import (
	"fmt"

	"github.com/rhaloubi/payment-gateway-cli/internal/client"
	"github.com/rhaloubi/payment-gateway-cli/internal/config"
	"github.com/rhaloubi/payment-gateway-cli/internal/ui"
	"github.com/spf13/cobra"
)

func NewRolesCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "roles",
		Short: "🔐 roles commands",
		Long:  "Manage user roles and permissions",
	}

	cmd.AddCommand(viewrolesCommand())
	//cmd.AddCommand(newLoginCommand())

	return cmd
}

// ViewRoles command
func viewrolesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "👤 Show user roles",
		RunE: func(cmd *cobra.Command, args []string) error {
			email := config.GetUserEmail()
			if email == "" && config.GetAccessToken() == "" {
				ui.Warning("⚠️  Not logged in")
				ui.Info("Run: payment-cli auth login")
				return nil
			}

			authClient := client.NewAuthClient()
			roles, err := authClient.GetAllRoles()
			if err != nil {
				return fmt.Errorf("failed to get profile: %w", err)
			}
			for _, role := range roles {
				ui.Info(fmt.Sprintf("🔑 Role ID: %s", role.ID))
				ui.Info(fmt.Sprintf("👤 Role: %s", role.Name))
				ui.Info(fmt.Sprintf("📝 Description: %s", role.Description))
			}
			return nil
		},
	}
}
