package model

import (
	"time"

	"github.com/google/uuid"
)

type IdempotencyKey struct {
	ID          uuid.UUID `json:"id"`
	Key         string    `json:"key"`
	MerchantID  uuid.UUID `json:"merchant_id"`
	RequestHash string    `json:"request_hash"`

	ResponseBody   []byte `json:"response_body"`
	ResponseStatus int    `json:"response_status"`

	// Metadata
	Endpoint  string `json:"endpoint"` // /v1/tokenize
	Method    string `json:"method"`   // POST
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"` // 24 hours from creation
}

func (ik *IdempotencyKey) IsExpired() bool {
	return time.Now().After(ik.ExpiresAt)
}

func (ik *IdempotencyKey) IsValid() bool {
	return !ik.IsExpired()
}
