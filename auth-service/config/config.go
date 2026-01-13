package config

import (
	"log"
	"os"
	"strings"
)

func GetEnv(key string) string {
	fileKey := key + "_FILE"
	if filePath := os.Getenv(fileKey); filePath != "" {
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read %s from file %s: %v", key, filePath, err)
		}
		return strings.TrimSpace(string(content))
	}

	return os.Getenv(key)
}

func GetEnvWithDefault(key, defaultValue string) string {
	value := GetEnv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
