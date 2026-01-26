# 💳 Payment CLI - Complete Guide

## 🚀 Quick Start (5 minutes)

```bash
# 1. Create project
mkdir payment-gateway-cli && cd payment-gateway-cli
go mod init github.com/rhaloubi/payment-gateway-cli

# 2. Install dependencies
go get github.com/spf13/cobra@latest \
       github.com/fatih/color@latest \
       github.com/olekukonko/tablewriter@latest \
       github.com/briandowns/spinner@latest \
       github.com/manifoldco/promptui@latest \
       gopkg.in/yaml.v3@latest

# 3. Copy the updated files from the outputs directory

# 4. Build
go build -o payment-cli cmd/main.go

# 5. Test
./payment-cli init
./payment-cli config use production
./payment-cli register
./payment-cli login
```

---

## 📂 Project Structure

```
payment-gateway-cli/
├── cmd/
│   └── main.go                    # Entry point
├── internal/
│   ├── commands/
│   │   ├── init.go                # Init command
│   │   ├── auth.go                # Register/Login/Logout
│   │   ├── merchant.go            # Merchant CRUD
│   │   ├── payment.go             # Payment operations
│   │   ├── payment_intent.go      # ✨ NEW: Payment intents with checkout
│   │   ├── apikey.go              # API key management
│   │   ├── config.go              # ✨ ENHANCED: Config commands
│   │   ├── health.go              # Health check
│   │   ├── interactive.go         # Interactive mode
│   │   └── roles.go               # Role management
│   ├── client/
│   │   ├── REST.go                # Base REST client
│   │   ├── auth_client.go         # HTTP client for auth
│   │   ├── merchant_client.go     # HTTP client for merchants
│   │   ├── payment_client.go      # HTTP client for payments
│   │   └── payment_intent_client.go # ✨ NEW: Payment intents client
│   ├── config/
│   │   └── config.go              # ✨ ENHANCED: Config with multiple URLs
│   ├── ui/
│   │   └── output.go              # Pretty output helpers
│   └── validation/
│       └── cardvali.go            # Card validation
├── go.mod
├── Makefile
└── README.md
```

---

## 🆕 What's New in This Update

### ✨ Enhanced Config Management
- **Multiple Environments**: Production (default) and Development
- **Separate Service URLs**: Auth, API, and Payment services
- **Config Commands**:
  - `config show` - View all settings (enhanced display)
  - `config set <key> <value>` - Modify preferences
  - `config use <environment>` - Switch between prod/dev
  - `config reset` - Reset to defaults

### 💳 Payment Intents (Hosted Checkout)
- **Browser-Based Checkout**: Opens checkout page automatically
- **Smart Polling**: Waits for payment completion
- **Real-Time Status**: Live updates on payment status
- **Command**: `payment intent create`

---

## 💻 CLI Commands

### Init & Configuration
```bash
payment-cli init                           # Initialize config
payment-cli config show                    # Show all config
payment-cli config set debug_mode true     # Enable debug
payment-cli config use production          # Switch to production
payment-cli config use development         # Switch to development
payment-cli config reset                   # Reset to defaults
```

### Authentication
```bash
payment-cli auth register                  # Register account
payment-cli auth login                     # Login
payment-cli auth logout                    # Logout
payment-cli whoami                         # Show current user
payment-cli auth profile                   # View profile
payment-cli auth change-password           # Change password
```

### Merchant Management
```bash
payment-cli merchant create                # Create merchant
payment-cli merchant get                   # Get merchant details
payment-cli merchant access-accounts       # Access merchant accounts
payment-cli merchant team                  # List team members
payment-cli merchant invitations           # List invitations
payment-cli merchant invite                # Invite user
payment-cli merchant setting               # View settings
```

### API Keys
```bash
payment-cli apikey create                  # Create API key
payment-cli apikey store <plain_key>       # Store API key locally
```

### Roles
```bash
payment-cli roles view                     # View all roles
```

### Payments
```bash
# Direct Authorization
payment-cli payment authorize              # Authorize payment

# ✨ NEW: Payment Intents (Hosted Checkout)
payment-cli payment intent create          # Create intent & open checkout
payment-cli payment intent create --amount 5000 --currency USD

# Transaction Management
payment-cli payment transactions           # List transactions
payment-cli payment capture                # Capture (dashboard)
payment-cli payment void                   # Void (dashboard)
payment-cli payment refund                 # Refund (dashboard)
```

### Other
```bash
payment-cli health                         # Health check
payment-cli --help                         # Show help
payment-cli --debug                        # Enable debug mode
```

---

## 🎨 Payment Intent Flow

### How It Works

1. **Create Intent**:
   ```bash
   payment-cli payment intent create
   ```

2. **CLI Prompts You For**:
   - Amount (in cents)
   - Currency (USD/EUR/MAD)
   - Description (optional)
   - Customer email (optional)
   - Capture method (automatic/manual)

3. **CLI Opens Browser**:
   - Automatically opens checkout page
   - Customer completes payment
   - Page redirects back with status

4. **CLI Polls Status**:
   - Checks payment status every 3 seconds
   - Shows real-time updates
   - Displays final result

### Example Session

```bash
$ payment-cli payment intent create

💳 Create Payment Intent
═══════════════════════════════════════

Amount (in cents): 5000
Currency: USD
Description: Premium Plan Subscription
Customer Email: john@example.com
Capture Method: automatic

✅ Payment intent created!

📋 Payment Intent Details:
  ID:          pi_abc123def456
  Amount:      5000 USD ($50.00 USD)
  Status:      created
  Description: Premium Plan Subscription
  Expires:     2026-01-26 15:30:00

🌐 Checkout URL:
  https://checkout-page-amber.vercel.app/checkout/pi_abc123def456?client_secret=...

🚀 Opening checkout page in your browser...
💡 Complete the payment in your browser

⏳ Waiting for payment completion...
   (Press Ctrl+C to cancel polling)

✅ Payment completed successfully!

📋 Payment Details:
  Intent ID:  pi_abc123def456
  Status:     authorized
  Payment ID: pay_xyz789uvw012

🎉 Transaction complete!
```

---

## ⚙️ Configuration

### Default Config Structure

```yaml
current_env: production

environments:
  production:
    api_url: https://paymentgateway.redahaloubi.com
    auth_url: https://paymentgateway.redahaloubi.com
    payment_url: https://paymentgateway.redahaloubi.com
  
  development:
    api_url: http://localhost:8080
    auth_url: http://localhost:8080
    payment_url: http://localhost:8080

credentials:
  access_token: ""
  refresh_token: ""
  user_email: ""
  merchant_id: ""
  api_key: ""

preferences:
  output_format: table
  color_enabled: true
  debug_mode: false
```

### Environment Switching

```bash
# Switch to development (localhost)
payment-cli config use development

# Switch to production (live API)
payment-cli config use production

# View current environment
payment-cli config show
```

### Preference Management

```bash
# Change output format
payment-cli config set output_format json
payment-cli config set output_format yaml
payment-cli config set output_format table

# Toggle colors
payment-cli config set color_enabled false

# Enable debug mode
payment-cli config set debug_mode true
```

---

## 🧪 Testing Flow

### Complete Test Workflow

```bash
# 1. Initialize
payment-cli init

# 2. Verify configuration
payment-cli config show

# 3. Register (if needed)
payment-cli auth register

# 4. Login
payment-cli auth login

# 5. Check status
payment-cli whoami

# 6. Create merchant
payment-cli merchant create

# 7. Create API key
payment-cli apikey create

# 8. Store API key
payment-cli apikey store pk_live_your_key_here

# 9. Test payment intent
payment-cli payment intent create
# Follow browser flow...

# 10. Check transactions
payment-cli payment transactions
```

### Testing Different Environments

```bash
# Test on localhost
payment-cli config use development
payment-cli payment intent create

# Test on production
payment-cli config use production
payment-cli payment intent create
```

---

## 🔍 Troubleshooting

### Config Issues

```bash
# View current config
payment-cli config show

# Reset if corrupted
payment-cli config reset

# Manually edit config
nano ~/.payment-cli/config.yaml
```

### Browser Won't Open

If the browser doesn't open automatically:
1. Copy the checkout URL from terminal
2. Paste it in your browser manually
3. The CLI will still poll for completion

### Payment Intent Timeout

If polling times out (15 minutes):
- Payment may still be processing
- Check your merchant dashboard
- Create a new intent if needed

### Connection Issues

```bash
# Check current environment
payment-cli config show

# Switch environment
payment-cli config use production

# Test connection
payment-cli health

# Enable debug mode
payment-cli --debug payment intent create
```

---

## 📦 Dependencies

```
github.com/spf13/cobra          # CLI framework
github.com/fatih/color          # Colors
github.com/olekukonko/tablewriter  # Tables
github.com/briandowns/spinner   # Loading spinners
github.com/manifoldco/promptui  # Interactive prompts
gopkg.in/yaml.v3                # YAML config
```

---

## 🎯 Key Features

### ✅ Production-Ready
- Default production URLs
- Secure credential storage
- Environment switching

### ✅ User-Friendly
- Interactive prompts
- Browser integration
- Real-time feedback
- Pretty formatting

### ✅ Flexible
- Multiple environments
- Configurable preferences
- Debug mode
- Multiple output formats

### ✅ Complete
- Full payment lifecycle
- Merchant management
- Team collaboration
- Role-based access

---

## 💡 Pro Tips

1. **Use Production by Default**: The CLI now defaults to production URLs
2. **Switch for Local Testing**: Use `config use development` for localhost
3. **Store API Keys Safely**: Use `apikey store` to save keys securely
4. **Monitor in Real-Time**: Payment intents show live status updates
5. **Debug When Needed**: Add `--debug` flag to any command for verbose output
6. **Check Config Often**: `config show` displays all current settings
7. **Reset if Stuck**: `config reset` fixes most configuration issues

---

## 🚀 Next Steps

### Phase 1: Core ✅ COMPLETE
- [x] Project setup
- [x] Auth commands
- [x] Merchant commands
- [x] Pretty UI
- [x] Config management
- [x] Payment intents with checkout

### Phase 2: Enhancements
- [ ] Payment intent status command
- [ ] Webhook testing
- [ ] Batch operations
- [ ] Export functionality

### Phase 3: Advanced
- [ ] Interactive dashboard
- [ ] Real-time notifications
- [ ] Analytics commands
- [ ] Integration testing

---

## 📚 API Documentation

Full API documentation: https://docs-paymentgateway.redahaloubi.com

---

## 🤝 Support

- Documentation: https://docs-paymentgateway.redahaloubi.com
- Issues: GitHub Issues
- Questions: Support Team

---

**Happy Building! 🚀**

Built with ❤️ for developers