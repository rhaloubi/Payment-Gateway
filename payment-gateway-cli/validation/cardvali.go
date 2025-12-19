package validation

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
)

type CardValidator struct {
}

func NewCardValidator() *CardValidator {
	return &CardValidator{}
}

type AuthorizeRequest struct {
	Amount   int64    `json:"amount"`
	Currency string   `json:"currency"`
	Card     Card     `json:"card"`
	Customer Customer `json:"customer"`
}

type Card struct {
	Number         string `json:"number"`
	CardholderName string `json:"cardholder_name"`
	ExpMonth       int    `json:"exp_month"`
	ExpYear        int    `json:"exp_year"`
	CVV            string `json:"cvv"`
}

type Customer struct {
	Email string `json:"email"`
}

func (*CardValidator) PromptAmount() (int64, error) {
	prompt := promptui.Prompt{
		Label: "Amount (in cents)",
		Validate: func(input string) error {
			val, err := strconv.ParseInt(input, 10, 64)
			if err != nil || val <= 0 {
				return errors.New("enter a valid amount > 0")
			}
			return nil
		},
	}

	result, err := prompt.Run()
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(result, 10, 64)
}
func (*CardValidator) PromptCardNumber() (string, error) {
	prompt := promptui.Prompt{
		Label: "Card Number",
		Validate: func(input string) error {
			if len(input) != 16 {
				return errors.New("card number must be 16 digits")
			}

			if _, err := strconv.Atoi(input); err != nil {
				return errors.New("card number must be numeric")
			}

			if !strings.HasPrefix(input, "4") &&
				!(input[:2] >= "51" && input[:2] <= "55") {
				return errors.New("only Visa (4) or Mastercard (51–55) supported")
			}

			return nil
		},
	}

	return prompt.Run()
}
func (*CardValidator) PromptCardholderName() (string, error) {
	prompt := promptui.Prompt{
		Label: "Cardholder Name",
		Validate: func(input string) error {
			if strings.TrimSpace(input) == "" {
				return errors.New("cardholder name cannot be empty")
			}
			return nil
		},
	}

	return prompt.Run()
}
func (*CardValidator) PromptExpMonth() (int, error) {
	months := []string{
		"01", "02", "03", "04", "05", "06",
		"07", "08", "09", "10", "11", "12",
	}

	selectPrompt := promptui.Select{
		Label: "Expiration Month",
		Items: months,
	}

	_, result, err := selectPrompt.Run()
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(result)
}
func (*CardValidator) PromptExpYear() (int, error) {
	prompt := promptui.Prompt{
		Label: "Expiration Year",
		Validate: func(input string) error {
			val, err := strconv.Atoi(input)
			if err != nil || val <= time.Now().Year() || val > time.Now().Year()+10 {
				return errors.New("enter a valid year >= current year and <= current year + 10")
			}
			return nil
		},
	}

	result, err := prompt.Run()
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(result)
}
func (*CardValidator) PromptCVV() (string, error) {
	prompt := promptui.Prompt{
		Label: "CVV",
		Mask:  '*',
		Validate: func(input string) error {
			if len(input) != 3 {
				return errors.New("CVV must be 3 digits")
			}
			if _, err := strconv.Atoi(input); err != nil {
				return errors.New("CVV must be numeric")
			}
			return nil
		},
	}

	return prompt.Run()
}
func (*CardValidator) PromptCurrency() (string, error) {
	currencies := []string{"USD", "EUR", "MAD"}

	selectPrompt := promptui.Select{
		Label: "Currency",
		Items: currencies,
	}

	_, result, err := selectPrompt.Run()
	return result, err
}
func (*CardValidator) PromptEmail() (string, error) {
	prompt := promptui.Prompt{
		Label: "Customer Email",
		Validate: func(input string) error {
			if !strings.Contains(input, "@") {
				return errors.New("invalid email address")
			}
			return nil
		},
	}

	return prompt.Run()
}

/*
func (c *CardValidator) runAuthorizePayment(cmd *cobra.Command, args []string) error {
	ui.Info("💳 Payment Authorization")
	ui.Info("-----------------------")

	amount, err := c.promptAmount()
	if err != nil {
		return err
	}

	currency, _ := c.promptCurrency()
	cardNumber, _ := c.promptCardNumber()
	cardholder, _ := c.promptCardholderName()
	expMonth, _ := c.promptExpMonth()
	expYear, _ := c.promptExpYear()
	cvv, _ := c.promptCVV()
	email, _ := c.promptEmail()

	req := AuthorizeRequest{
		Amount:   amount,
		Currency: currency,
		Card: Card{
			Number:         cardNumber,
			CardholderName: cardholder,
			ExpMonth:       expMonth,
			ExpYear:        expYear,
			CVV:            cvv,
		},
		Customer: Customer{
			Email: email,
		},
	}

	ui.Success("🧾 Payment details collected successfully")
	ui.Info(fmt.Sprintf("Card: •••• %s", cardNumber[len(cardNumber)-4:]))

	// NEXT STEP: send req to simulator
	return nil
}
*/
