# 🔐 Tokenization Service

**PCI-DSS Compliant Card Tokenization Microservice**

The Tokenization Service provides secure card data tokenization, reducing PCI compliance scope for merchants. It encrypts sensitive card data and issues non-sensitive tokens that can be safely stored and transmitted.

---

## 📋 Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [API Endpoints](#api-endpoints)
- [gRPC Interface](#grpc-interface)
- [Security](#security)
- [Authentication](#authentication)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage Examples](#usage-examples)
- [Testing](#testing)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)

---

## ✨ Features

### Core Capabilities

- ✅ **Secure Card Tokenization** - AES-256-GCM encryption with per-merchant keys
- ✅ **Duplicate Detection** - Returns existing token for duplicate cards (via fingerprinting)
- ✅ **Token Lifecycle Management** - Active, expired, revoked, and used states
- ✅ **Single-Use Tokens** - One-time payment tokens with automatic invalidation
- ✅ **Key Rotation** - Automated key rotation (90 days or 1M encryptions)
- ✅ **BIN Database** - Card type detection and issuer information
- ✅ **Idempotency** - Prevents duplicate tokenization on network retries (24-hour cache)
- ✅ **Rate Limiting** - Per-merchant limits (50/sec, 5000/hour)
- ✅ **Audit Logging** - Comprehensive PCI-compliant activity logs

### PCI Compliance

- ✅ **Field-Level Encryption** - Each field encrypted separately
- ✅ **Never Logs Sensitive Data** - Full PAN/CVV never logged
- ✅ **Access Controls** - Merchant-scoped data access
- ✅ **Secure Key Management** - HashiCorp Vault integration (production) 'NOT USED FOR NOW'
- ✅ **Audit Trail** - All tokenization/detokenization requests logged

### Communication Protocols

- ✅ **REST API** - Public merchant-facing endpoints
- ✅ **gRPC** - Internal service-to-service communication
- ✅ **HTTP/2** - gRPC with multiplexing and streaming

---

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                    TOKENIZATION SERVICE                       │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌─────────────┐      ┌──────────────┐                       │
│  │  REST API   │      │  gRPC Server │                       │
│  │  (Port 8003)│      │  (Port 50052)│                       │
│  └──────┬──────┘      └───────┬──────┘                       │
│         │                     │                               │
│         └─────────┬───────────┘                               │
│                   │                                           │
│         ┌─────────▼─────────┐                                 │
│         │   Handler Layer    │                                │
│         │  - TokenizeCard    │                                │
│         │  - Detokenize      │                                │
│         │  - ValidateToken   │                                │
│         └─────────┬──────────┘                                │
│                   │                                           │
│         ┌─────────▼─────────┐                                 │
│         │  Service Layer     │                                │
│         │  - Tokenization    │                                │
│         │  - KeyManagement   │                                │
│         │  - Validation      │                                │
│         │  - Encryption      │                                │
│         └─────────┬──────────┘                                │
│                   │                                           │
│         ┌─────────▼─────────┐                                 │
│         │ Repository Layer   │                                │
│         │  - CardVault       │                                │
│         │  - EncryptionKeys  │                                │
│         │  - UsageLogs       │                                │
│         └─────────┬──────────┘                                │
│                   │                                           │
│         ┌─────────▼─────────┐      ┌──────────┐              │
│         │    PostgreSQL      │      │  Redis   │              │
│         │  (Encrypted Data)  │      │ (Cache)  │              │
│         └────────────────────┘      └──────────┘              │
│                                                               │
└──────────────────────────────────────────────────────────────┘
```

### Data Flow

```
1. Merchant Request → API Gateway
2. Authentication (JWT/API Key)
3. Rate Limit Check (Redis)
4. Idempotency Check (Redis)
5. Card Validation (Luhn, expiry, CVV)
6. Fingerprint Generation (SHA-256)
7. Duplicate Check (PostgreSQL)
8. Get/Create Encryption Key (per merchant)
9. Encrypt Card Data (AES-256-GCM)
10. Generate Token (tok_live_xxx)
11. Store in Vault PostgreSQL
12. Cache Response (Redis - 24h for idempotency)
13. Return Token to Merchant
```

---

## WE REMOUVED THE REST API ENDPOINT SO THIS IS A INTERNAL SERVER GO TO THE PAYMENT SERVICE TO ACCESS THE ENDPOINTS

## 🔌 API Endpoints

### Public Endpoints (REST)

#### **POST /api/v1/tokenize**

Tokenize card data and return a secure token.

**Authentication:** JWT or API Key  
**Idempotency:** Supported (via `Idempotency-Key` header)  
**Rate Limit:** 50 requests/second per merchant

**Request:**

```json
{
  "card_number": "4242424242424242",
  "cardholder_name": "John Doe",
  "exp_month": 12,
  "exp_year": 2027,
  "cvv": "123",
  "is_single_use": false
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "token": "tok_live_4xJ3kL9mN2pQ5rT8vW1yZ4bC7dE0fG",
    "card": {
      "brand": "visa",
      "type": "credit",
      "last4": "4242",
      "exp_month": 12,
      "exp_year": 2027,
      "fingerprint": "abc123..."
    },
    "is_new_token": true
  }
}
```

---

#### **GET /api/v1/tokens/:token/validate**

Validate if a token is active and usable.

**Authentication:** JWT or API Key

**Response:**

```json
{
  "success": true,
  "data": {
    "valid": true,
    "card": {
      "brand": "visa",
      "type": "credit",
      "last4": "4242",
      "exp_month": 12,
      "exp_year": 2027
    },
    "status": "active",
    "usage_count": 0,
    "is_single_use": false
  }
}
```

---

#### **GET /api/v1/tokens/:token**

Retrieve token metadata (without decrypting card data).

**Authentication:** JWT or API Key

**Response:**

```json
{
  "success": true,
  "data": {
    "token": "tok_live_...",
    "card": {
      "brand": "visa",
      "last4": "4242",
      ...
    },
    "status": "active",
    "usage_count": 0,
    "created_at": "2025-11-18T10:00:00Z"
  }
}
```

---

#### **DELETE /api/v1/tokens/:token**

Revoke a token (mark as unusable).

**Authentication:** JWT or API Key

**Request:**

```json
{
  "reason": "Customer requested deletion"
}
```

**Response:**

```json
{
  "success": true,
  "message": "token revoked successfully"
}
```

---

#### **GET /api/v1/keys/statistics**

Get encryption key statistics for merchant.

**Authentication:** JWT (Owner/Admin only)  
**Permission Required:** `keys:read`

**Response:**

```json
{
  "success": true,
  "data": {
    "total_keys": 2,
    "active_keys": 1,
    "rotated_keys": 1,
    "oldest_key_age": "45d",
    "last_rotation": "2025-10-15T10:00:00Z"
  }
}
```

---

#### **POST /api/v1/keys/rotate**

Manually rotate merchant encryption key.

**Authentication:** JWT (Owner/Admin only)  
**Permission Required:** `keys:rotate`

**Response:**

```json
{
  "success": true,
  "data": {
    "new_key_id": "key_merchant_uuid_v2",
    "message": "encryption key rotated successfully"
  }
}
```

---

### Internal Endpoints (REST)

#### **POST /internal/v1/detokenize**

Retrieve original card data from token (internal services only).

**Authentication:** Internal Service Headers  
**Required Headers:**

- `X-Internal-Service: transaction-service`
- `X-Internal-Secret: <secret>`

**Request:**

```json
{
  "token": "tok_live_...",
  "merchant_id": "uuid",
  "transaction_id": "uuid",
  "usage_type": "payment",
  "amount": 9999,
  "currency": "USD"
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "card_number": "4242424242424242",
    "cardholder_name": "John Doe",
    "exp_month": 12,
    "exp_year": 2027,
    "card_brand": "visa",
    "last4": "4242"
  }
}
```
---

## 🌐 gRPC Interface

### Service Definition

```protobuf
service TokenizationService {
  rpc TokenizeCard(TokenizeCardRequest) returns (TokenizeCardResponse);
  rpc Detokenize(DetokenizeRequest) returns (DetokenizeResponse);
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc RevokeToken(RevokeTokenRequest) returns (RevokeTokenResponse);
}
```

### Usage Example (Transaction Service)

```go
package main

import (
    "context"
    "log"

    pb "github.com/rhaloubi/payment-gateway/tokenization-service/proto"
    "google.golang.org/grpc"
)

func main() {
    // Connect to tokenization service
    conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := pb.NewTokenizationServiceClient(conn)

    // Detokenize a card
    response, err := client.Detokenize(context.Background(), &pb.DetokenizeRequest{
        Token:         "tok_live_abc123...",
        MerchantId:    "merchant-uuid",
        TransactionId: "transaction-uuid",
        UsageType:     "payment",
        Amount:        9999, // $99.99
        Currency:      "USD",
    })

    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Card: %s", response.CardNumber)
}
```

---

## 🔒 Security

### Encryption

**Algorithm:** AES-256-GCM (Galois/Counter Mode)  
**Key Size:** 256 bits  
**Key Derivation:** Per-merchant unique keys  
**Key Storage:** HashiCorp Vault (production) or in-memory (development)

**Encryption Flow:**

```
1. Generate per-merchant DEK (Data Encryption Key)
2. Store DEK in Vault, encrypted by KEK (Key Encryption Key)
3. Cache DEK in memory for performance
4. Encrypt each field separately (card_number, name, expiry)
5. Store encrypted data + nonce + authentication tag
```

### Key Rotation

Keys are automatically rotated when:

- **Age:** 90 days old
- **Usage:** 1 million encryptions

**Rotation Process:**

1. Create new key version
2. Mark old key as rotated (but keep for decryption)
3. All new encryptions use new key
4. Old data remains encrypted with old key (lazy re-encryption)

### Card Fingerprinting

**Algorithm:** SHA-256  
**Input:** `card_number + exp_month + exp_year`  
**Purpose:** Detect duplicate cards without storing PAN

```go
fingerprint := SHA256(card_number + exp_month + exp_year)
// Example: "abc123def456..."
```

---

## 🔑 Authentication

### 1. JWT Authentication

**Use Case:** Dashboard, merchant portal  
**Header:** `Authorization: Bearer <jwt_token>`

**JWT Payload:**

```json
{
  "user_id": "uuid",
  "merchant_id": "uuid",
  "email": "user@merchant.com",
  "roles": ["owner", "admin"],
  "exp": 1700000000
}
```

### 2. API Key Authentication

**Use Case:** Server-to-server integration  
**Header:** `X-API-Key: pk_live_xxxxx`

**API Key Format:**

- **Production:** `pk_live_<64_chars>`
- **Test:** `pk_test_<64_chars>`

**Creation:**

```bash
curl -X POST http://localhost:8001/api/v1/api-keys \
  -H "Authorization: Bearer <jwt>" \
  -H "Content-Type: application/json" \
  -d '{"name": "Production API Key"}'
```

### 3. Internal Service Authentication

**Use Case:** Service-to-service (Transaction → Tokenization)  
**Headers:**

- `X-Internal-Service: transaction-service`
- `X-Internal-Secret: <shared_secret>`

**Allowed Services:**

- `transaction-service`
- `payment-api-service`
- `fraud-detection-service`

---

## 📦 Installation

### Prerequisites

- Go 1.23+
- PostgreSQL 15+
- Redis 7+
- HashiCorp Vault (production only)

### Setup Steps

```bash
# 1. Clone repository
git clone https://github.com/rhaloubi/payment-gateway.git
cd payment-gateway/tokenization-service

# 2. Install dependencies
go mod download

# 3. Copy environment file
cp .env.example .env

# 4. Configure database
psql -U postgres -c "CREATE DATABASE tokenization_db;"

# 5. Run migrations
go run cmd/migrate/migrate.go

# 6. Generate gRPC code (if modified proto files)
protoc \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/tokenization.proto

# 7. Start service
go run cmd/main.go

# Or use Air for hot reload
air
```

---

## ⚙️ Configuration

### Environment Variables

```bash
# Server
PORT=8003                    # HTTP server port
GRPC_PORT=50052             # gRPC server port
GIN_MODE=release            # gin mode: debug | release

# Database
DATABASE_DSN=postgresql://user:pass@localhost:5432/tokenization_db?sslmode=disable

# Redis
REDIS_DSN=redis://localhost:6379/3



# Auth Service
AUTH_SERVICE_URL=http://localhost:8001

# Internal Service Authentication
INTERNAL_SERVICE_SECRET=<change_in_production>


# Logging
LOG_LEVEL=info              # debug | info | warn | error
```

---

## 📖 Usage Examples

### Example 1: Tokenize Card (REST)

```bash
curl -X POST http://localhost:8003/api/v1/tokenize \
  -H "X-API-Key: pk_live_your_api_key" \
  -H "Idempotency-Key: unique-key-12345" \
  -H "Content-Type: application/json" \
  -d '{
    "card_number": "4242424242424242",
    "cardholder_name": "John Doe",
    "exp_month": 12,
    "exp_year": 2027,
    "cvv": "123"
  }'
```

### Example 2: Validate Token

```bash
curl http://localhost:8003/api/v1/tokens/tok_live_abc123.../validate \
  -H "X-API-Key: pk_live_your_api_key"
```

### Example 3: Detokenize (Internal gRPC)

```go
import (
    pb "github.com/rhaloubi/payment-gateway/tokenization-service/proto"
    "google.golang.org/grpc"
)

conn, _ := grpc.Dial("localhost:50052", grpc.WithInsecure())
client := pb.NewTokenizationServiceClient(conn)

response, err := client.Detokenize(ctx, &pb.DetokenizeRequest{
    Token: "tok_live_abc123...",
    MerchantId: "merchant-uuid",
    TransactionId: "txn-uuid",
    UsageType: "payment",
})
```

### Example 4: Check Idempotency

```bash
# First request
curl -X POST http://localhost:8003/api/v1/tokenize \
  -H "X-API-Key: pk_live_key" \
  -H "Idempotency-Key: test-123" \
  -d '{"card_number": "4242424242424242", ...}'

# Duplicate request (returns cached response)
curl -X POST http://localhost:8003/api/v1/tokenize \
  -H "X-API-Key: pk_live_key" \
  -H "Idempotency-Key: test-123" \
  -d '{"card_number": "4242424242424242", ...}'
```

---

## 🧪 Testing

### Health Checks

```bash
# Health check
curl http://localhost:8003/health

# Readiness check (checks DB + Redis)
curl http://localhost:8003/ready
```

### Test Cards

| Card Number         | Result      | Use Case           |
| ------------------- | ----------- | ------------------ |
| 4242 4242 4242 4242 | ✅ Success  | Valid card         |
| 4000 0000 0000 0002 | ❌ Declined | Generic failure    |
| 4000 0000 0000 9995 | ❌ Declined | Insufficient funds |
| 4000 0000 0000 0069 | ❌ Declined | Expired card       |
| 4000 0000 0000 0127 | ❌ Declined | Incorrect CVV      |

---

## 📊 Monitoring

### Metrics to Track

1. **Tokenization Success Rate**

   - Target: > 99.5%
   - Alert: < 95%

2. **Response Time (p95)**

   - Tokenization: < 500ms
   - Detokenization: < 200ms

3. **Cache Hit Rate (Idempotency)**

   - Target: > 5%
   - Indicates proper idempotency usage

4. **Key Rotation Compliance**

   - Keys rotated before 90 days
   - Keys rotated before 1M encryptions

5. **Failed Authentication Attempts**
   - Monitor for potential attacks
   - Alert: > 100 failures/hour

### Logs to Monitor

```bash
# Successful tokenization
INFO  Tokenization request  merchant_id=uuid last4=4242 card_brand=visa

# Failed validation
WARN  JWT validation failed  ip=192.168.1.1

# Key rotation needed
WARN  Key rotation needed  merchant_id=uuid reason="Key is 95 days old"

# Rate limit exceeded
WARN  Rate limit exceeded  merchant_id=uuid ip=192.168.1.1
```

---

## 🐛 Troubleshooting

### Issue: "Invalid API key"

**Cause:** API key not found or inactive

**Solution:**

```bash
# Verify API key in auth service
curl http://localhost:8001/api/v1/api-keys/merchant/{merchant_id} \
  -H "Authorization: Bearer <jwt>"
```

### Issue: "Token not found"

**Cause:** Token doesn't exist or belongs to different merchant

**Solution:**

- Verify token format: `tok_live_...` or `tok_test_...`
- Check merchant_id matches token owner
- Verify token hasn't been revoked

### Issue: "Rate limit exceeded"

**Cause:** Too many requests from merchant

**Solution:**

- Implement exponential backoff
- Cache tokens on client side
- Contact support to increase limits

### Issue: "Key rotation needed"

**Cause:** Encryption key is > 90 days old

**Solution:**

```bash
# Manually rotate key
curl -X POST http://localhost:8003/api/v1/keys/rotate \
  -H "Authorization: Bearer <jwt>"
```

### Issue: "Idempotency key already used"

**Cause:** Same key used with different request body

**Solution:**

- Use unique idempotency keys per request
- Don't reuse keys for different cards
- Keys expire after 24 hours

---

## 🔧 Development

### Generate gRPC Code

```bash
# Install protoc compiler
brew install protobuf  # macOS
apt install protobuf-compiler  # Ubuntu

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate code
protoc \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/tokenization.proto```

### Run with Hot Reload

```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Start with hot reload
air
```

### Database Migrations

```bash
# Run migrations
go run cmd/migrate/migrate.go

# Rollback (if needed)
# Edit migrate.go and run rollback function
```

---

## 📚 References

- [PCI DSS Requirements](https://www.pcisecuritystandards.org/)
- [NIST Cryptographic Standards](https://csrc.nist.gov/)
- [gRPC Documentation](https://grpc.io/docs/)
- [PostgreSQL TDE](https://www.postgresql.org/docs/current/encryption-options.html)

---

## 📄 License

Copyright © 2025 Payment Gateway. All rights reserved.

---
## Support

For issues and questions:

- GitHub : https://github.com/rhaloubi/Payment-Gateway-Microservices
- Email: redahaloubi8@gmail.com
---

**Last Updated:** November 18, 2025  
**Service Version:** v1.0.0  
**API Version:** v1
