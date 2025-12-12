package commands

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/rhaloubi/payment-gateway-cli/internal/client"
	"github.com/rhaloubi/payment-gateway-cli/internal/config"
	"github.com/rhaloubi/payment-gateway-cli/internal/ui"
	"github.com/spf13/cobra"
)

func NewAuthCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "🔐 Authentication commands",
		Long:  "Register, login, and manage authentication",
	}

	cmd.AddCommand(newRegisterCommand())
	cmd.AddCommand(newLoginCommand())
	cmd.AddCommand(newLogoutCommand())
	cmd.AddCommand(newWhoamiCommand())

	return cmd
}

// Register command
func newRegisterCommand() *cobra.Command {
	var email, name, password string

	cmd := &cobra.Command{
		Use:   "register",
		Short: "📝 Register a new user",
		Example: `  payment-cli register
  payment-cli register --email admin@company.com --name "John Doe"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Prompt for missing fields
			if email == "" {
				prompt := promptui.Prompt{
					Label: "Email",
					Validate: func(input string) error {
						if len(input) < 5 || !strings.Contains(input, "@") {
							return fmt.Errorf("invalid email")
						}
						return nil
					},
				}
				result, err := prompt.Run()
				if err != nil {
					return err
				}
				email = result
			}

			if name == "" {
				prompt := promptui.Prompt{
					Label: "Full Name",
				}
				result, err := prompt.Run()
				if err != nil {
					return err
				}
				name = result
			}

			if password == "" {
				prompt := promptui.Prompt{
					Label: "Password",
					Mask:  '*',
					Validate: func(input string) error {
						if len(input) < 8 {
							return fmt.Errorf("password must be at least 8 characters")
						}
						return nil
					},
				}
				result, err := prompt.Run()
				if err != nil {
					return err
				}
				password = result

				// Confirm password
				confirmPrompt := promptui.Prompt{
					Label: "Confirm Password",
					Mask:  '*',
				}
				confirm, err := confirmPrompt.Run()
				if err != nil {
					return err
				}
				if password != confirm {
					return fmt.Errorf("passwords do not match")
				}
			}

			// Register
			spinner := ui.NewSpinner("Creating account...")
			spinner.Start()

			authClient := client.NewAuthClient()
			user, err := authClient.Register(email, name, password)

			spinner.Stop()

			if err != nil {
				ui.Error(fmt.Sprintf("❌ Registration failed: %v", err))
				return err
			}

			ui.Success("✅ Account created successfully!")
			ui.Info(fmt.Sprintf("📧 Email: %s", user.Email))
			ui.Info(fmt.Sprintf("👤 Name: %s", user.Name))
			ui.Info(fmt.Sprintf("🆔 User ID: %s", user.ID))
			ui.Info("\n🚀 Next step: payment-cli login")

			return nil
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "email address")
	cmd.Flags().StringVar(&name, "name", "", "full name")
	cmd.Flags().StringVar(&password, "password", "", "password")

	return cmd
}

// Login command
func newLoginCommand() *cobra.Command {
	var email, password string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "🔐 Login to your user account",
		Example: `  payment-cli login
  payment-cli login --email admin@company.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Prompt for credentials
			if email == "" {
				prompt := promptui.Prompt{
					Label: "Email",
				}
				result, err := prompt.Run()
				if err != nil {
					return err
				}
				email = result
			}

			if password == "" {
				prompt := promptui.Prompt{
					Label: "Password",
					Mask:  '*',
				}
				result, err := prompt.Run()
				if err != nil {
					return err
				}
				password = result
			}

			// Login
			spinner := ui.NewSpinner("Logging in...")
			spinner.Start()

			authClient := client.NewAuthClient()
			tokens, user, err := authClient.Login(email, password)

			spinner.Stop()

			if err != nil {
				ui.Error(fmt.Sprintf("❌ Login failed: %v", err))
				ui.Info("\n💡 Tip: Make sure you've registered with 'payment-cli register'")
				return err
			}

			// Save credentials
			if err := config.SaveCredentials(tokens.AccessToken, tokens.RefreshToken, user.Email); err != nil {
				ui.Warning(fmt.Sprintf("⚠️  Could not save credentials: %v", err))
			}

			ui.Success("✅ Login successful!")
			ui.Info(fmt.Sprintf("👤 Logged in as: %s", user.Email))
			ui.Info("\n🚀 You're ready to go! Try:")
			ui.Info("  payment-cli merchant create")

			return nil
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "email address")
	cmd.Flags().StringVar(&password, "password", "", "password")

	return cmd
}

// Logout command
func newLogoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "👋 Logout from your account",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.ClearCredentials(); err != nil {
				return fmt.Errorf("failed to logout: %w", err)
			}

			ui.Success("✅ Logged out successfully!")
			return nil
		},
	}
}

// Whoami command
func newWhoamiCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "👤 Show current user",
		RunE: func(cmd *cobra.Command, args []string) error {
			email := config.GetUserEmail()
			if email == "" {
				ui.Warning("⚠️  Not logged in")
				ui.Info("Run: payment-cli login")
				return nil
			}

			ui.Info(fmt.Sprintf("👤 Logged in as: %s", email))
			ui.Info(fmt.Sprintf("🌍 Environment: %s", config.GetCurrentEnv()))
			ui.Info(fmt.Sprintf("🔗 API URL: %s", config.GetAPIURL()))

			return nil
		},
	}
}
