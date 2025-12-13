# 💳 Payment CLI - Quick Reference Card

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

# 3. Copy code from artifacts into files:
#    - cmd/main.go
#    - internal/commands/*.go
#    - internal/client/*.go
#    - internal/config/config.go
#    - internal/ui/output.go

# 4. Build
go build -o payment-cli cmd/main.go

# 5. Test
./payment-cli init
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
│   │   ├── apikey.go              # API key management
│   │   ├── config.go              # Config commands
│   │   ├── health.go              # Health check
│   │   └── interactive.go         # Interactive mode
│   ├── client/
│   │   ├── auth_client.go         # HTTP client for auth
│   │   ├── merchant_client.go     # HTTP client for merchants
│   │   └── payment_client.go      # HTTP client for payments
│   ├── config/
│   │   └── config.go              # Config management
│   └── ui/
│       └── output.go              # Pretty output helpers
├── go.mod
├── Makefile
└── README.md
```

---

## 🛠️ Build Commands

```bash
# Build for current OS
go build -o payment-cli cmd/main.go

# Or use Makefile
make build          # Build
make install        # Install globally
make clean          # Clean artifacts
make build-all      # Build for all platforms
make test           # Run tests
```

---

## 💻 CLI Commands

### Init & Auth
```bash
payment-cli init                    # Initialize config
payment-cli register                # Register account
payment-cli login                   # Login
payment-cli logout                  # Logout
payment-cli whoami                  # Show current user
```

### Merchant Management
```bash
payment-cli merchant create         # Create merchant
payment-cli merchant list           # List all merchants
payment-cli merchant get <id>       # Get merchant details
```

### API Keys
```bash
payment-cli apikey create --merchant-id <id>
payment-cli apikey list --merchant-id <id>
```

### Payments (Coming Soon)
```bash
payment-cli payment authorize --amount 5000
payment-cli payment capture <payment-id>
payment-cli payment void <payment-id>
payment-cli payment refund <payment-id>
```

### Config
```bash
payment-cli config show             # Show current config
payment-cli config use production   # Switch environment
```

### Other
```bash
payment-cli health                  # Health check
payment-cli --help                  # Show help
```

---

## 📝 Key Files Content

### `cmd/main.go` - Entry Point

```go
package main

import (
	"github.com/rhaloubi/payment-gateway-cli/internal/commands"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "payment-cli",
		Short: "💳 Payment Gateway CLI",
	}

	rootCmd.AddCommand(commands.NewInitCommand())
	rootCmd.AddCommand(commands.NewAuthCommands())
	rootCmd.AddCommand(commands.NewMerchantCommands())
	// ... add more commands

	rootCmd.Execute()
}
```

### `internal/config/config.yaml` - Config File

```yaml
current_env: development

environments:
  development:
    api_url: http://localhost:8000
    auth_url: http://localhost:8001
    payment_url: http://localhost:8004

preferences:
  output_format: table
  color_enabled: true
  debug_mode: false
```

---

## 🎨 UI Helpers

### Colors
```go
ui.Success("✅ Success message")     // Green
ui.Error("❌ Error message")         // Red
ui.Warning("⚠️  Warning message")    // Yellow
ui.Info("ℹ️  Info message")          // Cyan
```

### Spinner
```go
spinner := ui.NewSpinner("Loading...")
spinner.Start()
// ... do work ...
spinner.Stop()
```

### Tables
```go
table := ui.NewTable([]string{"ID", "Name", "Status"})
table.AddRow([]string{"1", "Test", "active"})
table.Render()
```

---

## 🔧 Common Patterns

### Command with Flags
```go
cmd := &cobra.Command{
	Use:   "create",
	Short: "Create something",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Your logic here
		return nil
	},
}
cmd.Flags().StringVar(&name, "name", "", "name")
return cmd
```

### Interactive Prompt
```go
prompt := promptui.Prompt{
	Label: "Email",
	Validate: func(input string) error {
		if len(input) < 5 {
			return fmt.Errorf("too short")
		}
		return nil
	},
}
result, err := prompt.Run()
```

### HTTP Request
```go
func (c *Client) post(url string, payload interface{}) error {
	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.httpClient.Do(req)
	// Handle response...
}
```

---

## 🐛 Debug Mode

```bash
# Enable debug output
payment-cli --debug merchant list

# Or set in config
payment-cli config set debug true
```

---

## 🧪 Testing Flow

```bash
# 1. Initialize
payment-cli init

# 2. Register
payment-cli register
# Email: test@example.com
# Name: Test User
# Password: password123

# 3. Login
payment-cli login
# Email: test@example.com
# Password: password123

# 4. Check status
payment-cli whoami

# 5. Create merchant
payment-cli merchant create
# Name: Test Corp
# Email: corp@example.com

# 6. List merchants
payment-cli merchant list
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

## 🔍 Troubleshooting

### CLI not found
```bash
# Check if built
ls build/payment-cli

# Rebuild
go build -o build/payment-cli cmd/main.go
```

### Import errors
```bash
# Download dependencies
go mod download
go mod tidy
```

### Config not found
```bash
# Initialize config
payment-cli init

# Check config exists
cat ~/.payment-cli/config.yaml
```

### Service connection failed
```bash
# Check service is running
curl http://localhost:8001/health

# Check config
payment-cli config show

# Update URL if needed
nano ~/.payment-cli/config.yaml
```

---

## 🚀 Next Steps

### Phase 1: Core (NOW) ✅
- [x] Project setup
- [x] Auth commands
- [x] Merchant commands (mock)
- [x] Pretty UI

### Phase 2: Real Services (Day 2)
- [ ] Connect to Auth Service
- [ ] Connect to Merchant Service
- [ ] Connect to Payment API
- [ ] Error handling

### Phase 3: Advanced (Day 3-4)
- [ ] Payment commands
- [ ] API key commands
- [ ] Interactive mode
- [ ] Testing & docs

---

## 💡 Pro Tips

1. **Use Makefile**: `make build` is faster than typing full commands
2. **Test Often**: Build and test after each command
3. **Mock First**: Get UI working with mock data, then connect services
4. **Debug Mode**: Use `--debug` flag to see what's happening
5. **Read Artifacts**: All the code you need is in the artifacts above!

---

## 📚 Full Documentation

See these artifacts for complete code:
- "Payment CLI - Complete Working Code"
- "CLI Tool - Complete Setup Instructions"
- "Makefile for CLI Tool"

---

**Happy Building! 🚀**

Time to complete: ~1 hour
Difficulty: ⭐⭐ (Easy-Medium)