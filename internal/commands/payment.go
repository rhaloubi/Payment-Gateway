package commands

import (
	"fmt"

	"github.com/rhaloubi/payment-gateway-cli/internal/client"
	"github.com/rhaloubi/payment-gateway-cli/internal/config"
	"github.com/rhaloubi/payment-gateway-cli/internal/ui"
	"github.com/rhaloubi/payment-gateway-cli/internal/validation"
	"github.com/spf13/cobra"
)

type TransactionFilters struct {
	Limit  *int
	Offset *int
	Status *string
}

func NewPaymentCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payment",
		Short: "💳 Payment operations",
		Long:  "Manage payments, including authorization ",
	}
	cmd.AddCommand(NewAuthorizePaymentCommands())
	cmd.AddCommand(NewCapturePaymentCommands())
	cmd.AddCommand(NewVoidPaymentCommands())
	cmd.AddCommand(NewRefundPaymentCommands())
	cmd.AddCommand(NewTransactionCommands())

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

func NewCapturePaymentCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "capture",
		Short: "Capture a payment",
		Long:  "Capture a payment using the payment gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.Info("💳 Payment Capture")
			ui.Warning("⚠️  Pleace go to the Payment Gateway Dashboard to capture the payment!")
			return nil
		},
	}
	return cmd
}

func NewVoidPaymentCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "void",
		Short: "Void a payment",
		Long:  "Void a payment using the payment gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.Info("💳 Payment Void")
			ui.Warning("⚠️  Pleace go to the Payment Gateway Dashboard to void the payment!")
			return nil
		},
	}
	return cmd
}

func NewRefundPaymentCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refund",
		Short: "Refund a payment",
		Long:  "Refund a payment using the payment gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.Info("💳 Payment Refund")
			ui.Warning("⚠️  Pleace go to the Payment Gateway Dashboard to refund the payment!")
			return nil
		},
	}
	return cmd
}

func NewTransactionCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transactions",
		Short: "List transactions",
		Long:  "List transactions using the payment gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := validation.NewCardValidator()

			ApiKey := config.GetApiKey()
			if ApiKey == "" {
				ui.Warning("⚠️  API key not set")
				ui.Info("Set it with: payment-cli apikey create")
				return nil
			}
			filters := TransactionFilters{}

			limit, err := c.PromptOptionalLimit()
			if err != nil {
				return err
			}
			filters.Limit = limit
			offset, err := c.PromptOptionalOffset()
			if err != nil {
				return err
			}
			filters.Offset = offset
			status, err := c.PromptOptionalStatus()
			if err != nil {
				return err
			}
			filters.Status = status

			paymentClient := client.NewPaymentClient()
			transactions, err := paymentClient.ListTransactions(
				ApiKey,
				&validation.TransactionFilters{
					Limit:  filters.Limit,
					Offset: filters.Offset,
					Status: filters.Status,
				},
			)
			if err != nil {
				return err
			}
			for _, transaction := range transactions {
				ui.Info(fmt.Sprintf("Transaction ID: %s", transaction.ID))
				ui.Info(fmt.Sprintf("Status: %s", transaction.Status))
				ui.Info(fmt.Sprintf("Amount: %d %s", transaction.Amount/100, transaction.Currency))
				ui.Info(fmt.Sprintf("Card Brand: %s", transaction.CardBrand))
				ui.Info(fmt.Sprintf("Card Last 4: •••• %s", transaction.CardLast4))

				ui.Info("-----------------------")
			}
			return nil
		},
	}
	return cmd
}
