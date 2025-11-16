package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

type EncryptionService struct{}

func NewEncryptionService() *EncryptionService {
	return &EncryptionService{}
}

type CardData struct {
	CardNumber     string
	CardholderName string
	ExpiryMonth    string
	ExpiryYear     string
}

type EncryptedCardData struct {
	EncryptedCardNumber     string
	EncryptedCardholderName string
	EncryptedExpiryMonth    string
	EncryptedExpiryYear     string
}

func (s *EncryptionService) Encrypt(plaintext string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", errors.New("encryption key must be 32 bytes (AES-256)")
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and prepend nonce
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Return base64-encoded result
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64-encoded ciphertext using AES-256-GCM
func (s *EncryptionService) Decrypt(ciphertextBase64 string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", errors.New("decryption key must be 32 bytes (AES-256)")
	}

	// Decode base64
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plaintext), nil
}

// EncryptCardData encrypts all card fields
func (s *EncryptionService) EncryptCardData(data CardData, key []byte) (*EncryptedCardData, error) {
	encrypted := &EncryptedCardData{}

	// Encrypt card number
	cardNumber, err := s.Encrypt(data.CardNumber, key)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt card number: %w", err)
	}
	encrypted.EncryptedCardNumber = cardNumber

	// Encrypt cardholder name
	if data.CardholderName != "" {
		cardholderName, err := s.Encrypt(data.CardholderName, key)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt cardholder name: %w", err)
		}
		encrypted.EncryptedCardholderName = cardholderName
	}

	// Encrypt expiry month
	expiryMonth, err := s.Encrypt(data.ExpiryMonth, key)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt expiry month: %w", err)
	}
	encrypted.EncryptedExpiryMonth = expiryMonth

	// Encrypt expiry year
	expiryYear, err := s.Encrypt(data.ExpiryYear, key)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt expiry year: %w", err)
	}
	encrypted.EncryptedExpiryYear = expiryYear

	return encrypted, nil
}

// DecryptCardData decrypts all card fields
func (s *EncryptionService) DecryptCardData(encrypted EncryptedCardData, key []byte) (*CardData, error) {
	data := &CardData{}

	// Decrypt card number
	cardNumber, err := s.Decrypt(encrypted.EncryptedCardNumber, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt card number: %w", err)
	}
	data.CardNumber = cardNumber

	// Decrypt cardholder name (if present)
	if encrypted.EncryptedCardholderName != "" {
		cardholderName, err := s.Decrypt(encrypted.EncryptedCardholderName, key)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt cardholder name: %w", err)
		}
		data.CardholderName = cardholderName
	}

	// Decrypt expiry month
	expiryMonth, err := s.Decrypt(encrypted.EncryptedExpiryMonth, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt expiry month: %w", err)
	}
	data.ExpiryMonth = expiryMonth

	// Decrypt expiry year
	expiryYear, err := s.Decrypt(encrypted.EncryptedExpiryYear, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt expiry year: %w", err)
	}
	data.ExpiryYear = expiryYear

	return data, nil
}

// Used to detect duplicate cards
func (s *EncryptionService) GenerateCardFingerprint(cardNumber, expiryMonth, expiryYear string) string {
	data := fmt.Sprintf("%s:%s:%s", cardNumber, expiryMonth, expiryYear)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// HashToken creates a SHA-256 hash of a token (for comparison)
func (s *EncryptionService) HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}

// GenerateKey generates a random 256-bit (32-byte) encryption key
func (s *EncryptionService) GenerateKey() ([]byte, error) {
	key := make([]byte, 32) // AES-256 requires 32 bytes
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return key, nil
}

// GenerateKeyID generates a unique identifier for an encryption key
func (s *EncryptionService) GenerateKeyID(merchantID string, version int) string {
	return fmt.Sprintf("key_%s_v%d", merchantID, version)
}
