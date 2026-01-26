package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	CurrentEnv   string                 `yaml:"current_env"`
	Environments map[string]Environment `yaml:"environments"`
	Credentials  Credentials            `yaml:"credentials"`
	Preferences  Preferences            `yaml:"preferences"`
}

type Environment struct {
	APIURL     string `yaml:"api_url"`
	AuthURL    string `yaml:"auth_url"`
	PaymentURL string `yaml:"payment_url"`
}

type Credentials struct {
	AccessToken  string `yaml:"access_token"`
	RefreshToken string `yaml:"refresh_token"`
	UserEmail    string `yaml:"user_email"`
	MerchantID   string `yaml:"merchant_id"`
	ApiKey       string `yaml:"api_key"`
}

type Preferences struct {
	OutputFormat string `yaml:"output_format"`
	ColorEnabled bool   `yaml:"color_enabled"`
	DebugMode    bool   `yaml:"debug_mode"`
}

var globalConfig *Config

// Init creates config directory and default config
func Init() error {
	configDir := getConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := GetConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config with production URLs
		defaultConfig := &Config{
			CurrentEnv: "production",
			Environments: map[string]Environment{
				"production": {
					APIURL:     "https://paymentgateway.redahaloubi.com",
					AuthURL:    "https://paymentgateway.redahaloubi.com",
					PaymentURL: "https://paymentgateway.redahaloubi.com",
				},
				"development": {
					APIURL:     "http://localhost:8080",
					AuthURL:    "http://localhost:8080",
					PaymentURL: "http://localhost:8080",
				},
			},
			Preferences: Preferences{
				OutputFormat: "table",
				ColorEnabled: true,
				DebugMode:    false,
			},
		}

		data, err := yaml.Marshal(defaultConfig)
		if err != nil {
			return err
		}

		if err := os.WriteFile(configPath, data, 0600); err != nil {
			return err
		}
	}

	return Load(configPath)
}

// Load loads config from file
func Load(path string) error {
	if path == "" {
		path = GetConfigPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}

	globalConfig = &cfg
	return nil
}

// Save saves config to file
func Save() error {
	if globalConfig == nil {
		return fmt.Errorf("config not loaded")
	}

	data, err := yaml.Marshal(globalConfig)
	if err != nil {
		return err
	}

	return os.WriteFile(GetConfigPath(), data, 0600)
}

// SaveCredentials saves authentication tokens
func SaveCredentials(accessToken, refreshToken, email string) error {
	if globalConfig == nil {
		if err := Load(""); err != nil {
			return err
		}
	}

	globalConfig.Credentials.AccessToken = accessToken
	globalConfig.Credentials.RefreshToken = refreshToken
	globalConfig.Credentials.UserEmail = email

	return Save()
}

// GetMerchantID returns the current merchant ID
func GetMerchantID() string {
	if globalConfig == nil {
		return ""
	}
	return globalConfig.Credentials.MerchantID
}

// GetApiKey returns the current API key
func GetApiKey() string {
	if globalConfig == nil {
		return ""
	}
	return globalConfig.Credentials.ApiKey
}

// SetMerchantID sets the merchant ID
func SetMerchantID(id string) error {
	if globalConfig == nil {
		if err := Load(""); err != nil {
			return err
		}
	}
	globalConfig.Credentials.MerchantID = id
	return Save()
}

// SetApiKey sets the API key
func SetApiKey(apiKey string) error {
	if globalConfig == nil {
		if err := Load(""); err != nil {
			return err
		}
	}
	globalConfig.Credentials.ApiKey = apiKey
	return Save()
}

// SetCurrentEnv switches the current environment
func SetCurrentEnv(env string) error {
	if globalConfig == nil {
		if err := Load(""); err != nil {
			return err
		}
	}

	if _, exists := globalConfig.Environments[env]; !exists {
		return fmt.Errorf("environment '%s' not found", env)
	}

	globalConfig.CurrentEnv = env
	return Save()
}

// SetConfigValue sets a specific config value
func SetConfigValue(key, value string) error {
	if globalConfig == nil {
		if err := Load(""); err != nil {
			return err
		}
	}

	switch key {
	case "output_format":
		globalConfig.Preferences.OutputFormat = value
	case "color_enabled":
		globalConfig.Preferences.ColorEnabled = value == "true"
	case "debug_mode":
		globalConfig.Preferences.DebugMode = value == "true"
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	return Save()
}

// ResetConfig resets config to default values
func ResetConfig() error {
	configPath := GetConfigPath()
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return Init()
}

// ClearCredentials removes saved credentials
func ClearCredentials() error {
	if globalConfig == nil {
		return nil
	}

	globalConfig.Credentials = Credentials{}
	return Save()
}

// Getters
func GetConfigPath() string {
	return filepath.Join(getConfigDir(), "config.yaml")
}

func getConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".payment-cli")
}

func GetCurrentEnv() string {
	if globalConfig == nil {
		return "production"
	}
	return globalConfig.CurrentEnv
}

func GetAPIURL() string {
	if globalConfig == nil {
		return "https://paymentgateway.redahaloubi.com"
	}
	env := globalConfig.Environments[globalConfig.CurrentEnv]
	return env.APIURL
}

func GetAuthURL() string {
	if globalConfig == nil {
		return "https://paymentgateway.redahaloubi.com"
	}
	env := globalConfig.Environments[globalConfig.CurrentEnv]
	return env.AuthURL
}

func GetPaymentURL() string {
	if globalConfig == nil {
		return "https://paymentgateway.redahaloubi.com"
	}
	env := globalConfig.Environments[globalConfig.CurrentEnv]
	return env.PaymentURL
}

func GetAccessToken() string {
	if globalConfig == nil {
		return ""
	}
	return globalConfig.Credentials.AccessToken
}

func GetUserEmail() string {
	if globalConfig == nil {
		return ""
	}
	return globalConfig.Credentials.UserEmail
}

func GetOutputFormat() string {
	if globalConfig == nil {
		return "table"
	}
	return globalConfig.Preferences.OutputFormat
}

func GetColorEnabled() bool {
	if globalConfig == nil {
		return true
	}
	return globalConfig.Preferences.ColorEnabled
}

func GetDebugMode() bool {
	if globalConfig == nil {
		return false
	}
	return globalConfig.Preferences.DebugMode
}

func SetDebug(enabled bool) {
	if globalConfig != nil {
		globalConfig.Preferences.DebugMode = enabled
	}
}

func SetOutputFormat(format string) {
	if globalConfig != nil {
		globalConfig.Preferences.OutputFormat = format
	}
}
