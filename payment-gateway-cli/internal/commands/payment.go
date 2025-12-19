package commands

import (
	"fmt"

	"github.com/rhaloubi/payment-gateway-cli/internal/client"
	"github.com/rhaloubi/payment-gateway-cli/internal/config"
	"github.com/rhaloubi/payment-gateway-cli/internal/ui"
	"github.com/rhaloubi/payment-gateway-cli/validation"
	"github.com/spf13/cobra"
)

func NewPaymentCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payment",
		Short: "💳 Payment operations",
		Long:  "Manage payments, including authorization and settlement",
	}
	cmd.AddCommand(NewAuthorizePaymentCommands())
	return cmd
}

func NewAuthorizePaymentCommands() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "authorize",
		Short: "Authorize a payment",
		Long:  "Authorize a payment using the payment gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := validation.NewCardValidator()
			ApiKey := config.GetApiKey()
			if ApiKey == "" {
				ui.Warning("⚠️  API key not set")
				ui.Info("Set it with: payment-cli apikey create")
				return nil
			}

			ui.Info("💳 Payment Authorization")
			ui.Info("-----------------------")

			amount, err := c.PromptAmount()
			if err != nil {
				return err
			}

			currency, err := c.PromptCurrency()
			if err != nil {
				return err
			}
			cardNumber, err := c.PromptCardNumber()
			if err != nil {
				return err
			}
			cardholder, err := c.PromptCardholderName()
			if err != nil {
				return err
			}
			expMonth, err := c.PromptExpMonth()
			if err != nil {
				return err
			}
			expYear, err := c.PromptExpYear()
			if err != nil {
				return err
			}
			cvv, err := c.PromptCVV()
			if err != nil {
				return err
			}
			email, err := c.PromptEmail()
			if err != nil {
				return err
			}

			req := &validation.AuthorizeRequest{
				Amount:   amount,
				Currency: currency,
				Card: validation.Card{
					Number:         cardNumber,
					CardholderName: cardholder,
					ExpMonth:       expMonth,
					ExpYear:        expYear,
					CVV:            cvv,
				},
				Customer: validation.Customer{
					Email: email,
				},
			}
			// NEXT STEP: send req to simulator
			paymentClient := client.NewPaymentClient()
			authResp, err := paymentClient.AuthorizePayment(req, ApiKey)
			if err != nil {
				return err
			}
			Amount := authResp.Amount / 100
			ui.Success("🧾 Payment details collected successfully")

			ui.Info(fmt.Sprintf("Authorization ID: %s", authResp.ID))
			ui.Info(fmt.Sprintf("Status: %s", authResp.Status))
			ui.Info(fmt.Sprintf("Amount: %d %s", Amount, authResp.Currency))
			ui.Info(fmt.Sprintf("Card Brand: %s", authResp.CardBrand))
			ui.Info(fmt.Sprintf("Card Last 4: •••• %s", authResp.CardLast4))
			ui.Info(fmt.Sprintf("Auth Code: %s", authResp.AuthCode))
			ui.Info(fmt.Sprintf("Fraud Decision: %s", authResp.FraudDecision))
			ui.Info(fmt.Sprintf("Response Message: %s", authResp.ResponseMessage))
			ui.Info(fmt.Sprintf("Transaction ID: %s", authResp.TransactionID))
			return nil
		},
	}

	return cmd
}
