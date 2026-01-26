package commands

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/rhaloubi/payment-gateway-cli/internal/client"
	"github.com/rhaloubi/payment-gateway-cli/internal/config"
	"github.com/rhaloubi/payment-gateway-cli/internal/ui"
	"github.com/spf13/cobra"
)

func NewPaymentIntentCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "intent",
		Aliases: []string{"intents"},
		Short:   "💳 Payment Intent operations",
		Long:    "Create and manage payment intents for hosted checkout",
	}

	cmd.AddCommand(newCreatePaymentIntentCommand())

	return cmd
}

func newCreatePaymentIntentCommand() *cobra.Command {
	var (
		amount        int64
		currency      string
		successURL    string
		cancelURL     string
		description   string
		customerEmail string
		captureMethod string
		orderID       string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a payment intent",
		Long: `Create a payment intent for hosted checkout.
This will generate a checkout URL and open it in your browser.
The CLI will wait for the payment to complete.`,
		Example: `  payment-cli payment intent create
  payment-cli payment intent create --amount 5000 --currency USD
  payment-cli payment intent create --amount 9999 --currency EUR --email customer@example.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			apiKey := config.GetApiKey()
			if apiKey == "" {
				ui.Warning("⚠️  API key not set")
				ui.Info("Set it with: payment-cli apikey create")
				return nil
			}

			ui.Info("💳 Create Payment Intent")
			ui.Info("═══════════════════════════════════════")
			ui.Info("")

			// Prompt for required fields if not provided
			if amount == 0 {
				prompt := promptui.Prompt{
					Label: "Amount (in cents)",
					Validate: func(input string) error {
						var val int64
						_, err := fmt.Sscanf(input, "%d", &val)
						if err != nil || val <= 0 {
							return fmt.Errorf("enter a valid amount > 0")
						}
						return nil
					},
				}
				result, err := prompt.Run()
				if err != nil {
					return err
				}
				fmt.Sscanf(result, "%d", &amount)
			}

			if currency == "" {
				selectPrompt := promptui.Select{
					Label: "Currency",
					Items: []string{"USD", "EUR", "MAD"},
				}
				var err error
				_, currency, err = selectPrompt.Run()
				if err != nil {
					return err
				}
			}

			if description == "" {
				prompt := promptui.Prompt{
					Label: "Description (optional, press enter to skip)",
				}
				description, _ = prompt.Run()
			}

			if customerEmail == "" {
				prompt := promptui.Prompt{
					Label: "Customer Email (optional, press enter to skip)",
					Validate: func(input string) error {
						if input != "" && !strings.Contains(input, "@") {
							return fmt.Errorf("invalid email address")
						}
						return nil
					},
				}
				customerEmail, _ = prompt.Run()
			}

			if captureMethod == "" {
				selectPrompt := promptui.Select{
					Label: "Capture Method",
					Items: []string{"automatic", "manual"},
				}
				var err error
				_, captureMethod, err = selectPrompt.Run()
				if err != nil {
					return err
				}
			}

			// Use localhost callback URL for CLI
			if successURL == "" {
				successURL = "http://localhost:8765/success?payment_intent={CHECKOUT_SESSION_ID}"
			}
			if cancelURL == "" {
				cancelURL = "http://localhost:8765/cancel"
			}

			// Create payment intent request
			req := &client.PaymentIntentRequest{
				Amount:        amount,
				Currency:      currency,
				SuccessURL:    successURL,
				CancelURL:     cancelURL,
				Description:   description,
				CustomerEmail: customerEmail,
				CaptureMethod: captureMethod,
				OrderID:       orderID,
			}

			ui.Info("")
			spinner := ui.NewSpinner("Creating payment intent...")
			spinner.Start()

			intentClient := client.NewPaymentIntentClient()
			intent, err := intentClient.CreatePaymentIntent(req, apiKey)

			spinner.Stop()

			if err != nil {
				ui.Error(fmt.Sprintf("❌ Failed to create payment intent: %v", err))
				return err
			}

			ui.Success("✅ Payment intent created!")
			ui.Info("")
			ui.Info("📋 Payment Intent Details:")
			ui.Info(fmt.Sprintf("  ID:          %s", intent.ID))
			ui.Info(fmt.Sprintf("  Amount:      %d %s (%.2f %s)", intent.Amount, intent.Currency, float64(intent.Amount)/100, intent.Currency))
			ui.Info(fmt.Sprintf("  Status:      %s", intent.Status))
			if intent.Description != "" {
				ui.Info(fmt.Sprintf("  Description: %s", intent.Description))
			}
			ui.Info(fmt.Sprintf("  Expires:     %s", intent.ExpiresAt.Format("2006-01-02 15:04:05")))
			ui.Info("")
			ui.Info("🌐 Checkout URL:")
			ui.Info(fmt.Sprintf("  %s", intent.CheckoutURL))
			ui.Info("")

			// Open browser
			ui.Info("🚀 Opening checkout page in your browser...")
			ui.Info("💡 Complete the payment in your browser")
			ui.Info("")

			if err := openBrowser(intent.CheckoutURL); err != nil {
				ui.Warning(fmt.Sprintf("⚠️  Could not open browser automatically: %v", err))
				ui.Info("Please open this URL manually:")
				ui.Info(fmt.Sprintf("  %s", intent.CheckoutURL))
			}

			// Start polling for payment status
			ui.Info("⏳ Waiting for payment completion...")
			ui.Info("   (Press Ctrl+C to cancel polling)")
			ui.Info("")

			return pollPaymentStatus(intentClient, intent.ID)
		},
	}

	cmd.Flags().Int64VarP(&amount, "amount", "a", 0, "Amount in cents")
	cmd.Flags().StringVarP(&currency, "currency", "c", "", "Currency (USD, EUR, MAD)")
	cmd.Flags().StringVar(&successURL, "success-url", "", "Success redirect URL")
	cmd.Flags().StringVar(&cancelURL, "cancel-url", "", "Cancel redirect URL")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Payment description")
	cmd.Flags().StringVarP(&customerEmail, "email", "e", "", "Customer email")
	cmd.Flags().StringVar(&captureMethod, "capture", "", "Capture method (automatic/manual)")
	cmd.Flags().StringVar(&orderID, "order-id", "", "Your internal order ID")

	return cmd
}

// openBrowser opens the specified URL in the default browser
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

// pollPaymentStatus polls the payment intent status until completion
func pollPaymentStatus(client *client.PaymentIntentClient, intentID string) error {
	spinner := ui.NewSpinner("Checking payment status...")
	spinner.Start()

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	timeout := time.After(15 * time.Minute) // 15 minute timeout

	for {
		select {
		case <-timeout:
			spinner.Stop()
			ui.Warning("⏰ Polling timeout - payment intent may still be processing")
			ui.Info(fmt.Sprintf("Check status later with: payment-cli payment intent status %s", intentID))
			return nil

		case <-ticker.C:
			status, err := client.GetPaymentIntent(intentID)
			if err != nil {
				// Don't stop on error, keep polling
				continue
			}

			switch status.Status {
			case "authorized", "completed", "succeeded":
				spinner.Stop()
				ui.Success("✅ Payment completed successfully!")
				ui.Info("")
				ui.Info("📋 Payment Details:")
				ui.Info(fmt.Sprintf("  Intent ID:  %s", status.ID))
				ui.Info(fmt.Sprintf("  Status:     %s", status.Status))
				if status.PaymentID != "" {
					ui.Info(fmt.Sprintf("  Payment ID: %s", status.PaymentID))
				}
				ui.Info("")
				ui.Success("🎉 Transaction complete!")
				return nil

			case "failed", "expired", "cancelled":
				spinner.Stop()
				ui.Error(fmt.Sprintf("❌ Payment %s", status.Status))
				ui.Info("")
				ui.Info("💡 You can create a new payment intent to try again")
				return fmt.Errorf("payment %s", status.Status)

			case "created", "processing":
				// Continue polling
				continue

			default:
				// Unknown status, continue polling
				continue
			}
		}
	}
}
