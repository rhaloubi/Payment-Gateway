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
	APIURL      string `yaml:"api_url"`
	AuthURL     string `yaml:"auth_url"`
	PaymentURL  string `yaml:"payment_url"`
	MerchantURL string `yaml:"merchant_url"`
}

type Credentials struct {
	AccessToken  string `yaml:"access_token"`
	RefreshToken string `yaml:"refresh_token"`
	UserEmail    string `yaml:"user_email"`
	MerchantID   string `yaml:"merchant_id"`
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
		// Create default config
		defaultConfig := &Config{
			CurrentEnv: "development",
			Environments: map[string]Environment{
				"development": {
					APIURL:      "http://localhost:8000",
					AuthURL:     "http://localhost:8001",
					PaymentURL:  "http://localhost:8004",
					MerchantURL: "http://localhost:8002",
				},
				"production": {
					APIURL:      "https://api.example.com",
					AuthURL:     "https://api.example.com",
					PaymentURL:  "https://api.example.com",
					MerchantURL: "https://api.example.com",
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

func GetMerchantID() string {
	if globalConfig == nil {
		return ""
	}
	return globalConfig.Credentials.MerchantID
}

func SetMerchantID(id string) error {
	if globalConfig != nil {
		if err := Load(""); err != nil {
			return err
		}
	}
	globalConfig.Credentials.MerchantID = id
	return Save()
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
		return "development"
	}
	return globalConfig.CurrentEnv
}

func GetAPIURL() string {
	if globalConfig == nil {
		return "http://localhost:8001"
	}
	env := globalConfig.Environments[globalConfig.CurrentEnv]
	return env.AuthURL
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
