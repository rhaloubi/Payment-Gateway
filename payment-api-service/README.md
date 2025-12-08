# 💳 Payment API Service

**Public Payment Processing Gateway**

The Payment API Service is the main entry point for merchants to process payments. It orchestrates calls to internal services (Tokenization, Fraud Detection, Transaction) and provides a unified REST API for payment operations.

---

## 📋 Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Payment Flow](#payment-flow)
- [API Endpoints](#api-endpoints)
- [Test Cards](#test-cards)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage Examples](#usage-examples)
- [Webhooks](#webhooks)
- [Error Handling](#error-handling)

---

## ✨ Features

### Core Payment Operations
- ✅ **Authorize** - Hold funds without charging
- ✅ **Sale** - Authorize + immediate capture (one-step payment)
- ✅ **Capture** - Charge previously authorized funds
- ✅ **Void** - Cancel authorization before capture
- ✅ **Refund** - Return funds to customer

### Advanced Features
- ✅ **Idempotency** - Prevents duplicate charges (24-hour cache)
- ✅ **Rate Limiting** - 20 payments/second, 10,000/hour per merchant
- ✅ **PCI Compliance** - Card data never logged or stored in this service
- ✅ **Fraud Detection** - Real-time risk scoring (mock implementation)
- ✅ **Webhooks** - Async notifications with retry logic
- ✅ **Audit Logging** - Complete payment activity tracking
- ✅ **Multi-Currency** - USD and EUR and MAD supported

### Integration
- ✅ **REST API** - Simple HTTP/JSON interface
- ✅ **API Key Authentication** - Secure merchant authentication
- ✅ **Service Orchestration** - Coordinates tokenization, fraud, and transaction services

---

## 🏗️ Architecture

```
┌────────────────────────────────────────────────────────────┐
│              PAYMENT API SERVICE (Gateway)                 │
│                    Port 8004 (HTTP)                        │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  ┌──────────────────────────────────────────────────┐    │
│  │           REST API Endpoints                      │    │
│  │  POST /v1/payments/authorize                      │    │
│  │  POST /v1/payments/sale                           │    │
│  │  POST /v1/payments/{id}/capture                   │    │
│  │  POST /v1/payments/{id}/void                      │    │
│  │  POST /v1/payments/{id}/refund                    │    │
│  │  GET  /v1/payments/{id}                           │    │
│  └──────────────────────┬───────────────────────────┘    │
│                         │                                 │
│  ┌──────────────────────▼───────────────────────────┐    │
│  │        Payment Orchestration Service             │    │
│  │  - Request validation                             │    │
│  │  - Service coordination                           │    │
│  │  - Response aggregation                           │    │
│  └──────────────────────┬───────────────────────────┘    │
│                         │                                 │
│          ┌──────────────┼──────────────┐                 │
│          │              │              │                 │
│     ┌────▼────┐    ┌───▼────┐    ┌───▼────┐            │
│     │  Token  │    │ Fraud  │    │  Txn   │            │
│     │ Service │    │Service │    │Service │            │
│     │ (gRPC)  │    │ (Mock) │    │ (Mock) │            │
│     └─────────┘    └────────┘    └────────┘            │
│                                                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │        PostgreSQL + Redis Storage                 │  │
│  │  - Payments                                       │  │
│  │  - Payment Events                                 │  │
│  │  - Webhook Deliveries                             │  │
│  └──────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────┘
```

---

## 🔄 Payment Flow

### Complete Authorization Flow

```
1. Merchant sends payment request → POST /v1/payments/authorize
   ↓
2. Validate API Key → Auth Service (HTTP)
   ↓
3. Check Rate Limit → Redis
   ↓
4. Check Idempotency → Redis (return cached if duplicate)
   ↓
5. Validate Request → Amount, currency, card format
   ↓
6. Tokenize Card → Tokenization Service (gRPC)
   • Validate card (Luhn, expiry, CVV)
   • Encrypt card data
   • Return token (tok_live_xxx)
   ↓
7. Fraud Check → Fraud Service (Mock)
   • Calculate risk score (0-100)
   • Make decision (approve/review/decline)
   • Return fraud analysis
   ↓
8. If fraud score > 70 → Decline payment
   ↓
9. Authorize Transaction → Transaction Service (Mock)
   • Process authorization
   • Simulate issuer response
   • Return auth code or decline reason
   ↓
10. Save Payment Record → PostgreSQL
    • Store payment details
    • Log payment event
    ↓
11. Cache Response → Redis (for idempotency, 24h)
    ↓
12. Send Webhook → Async (if configured)
    • payment.authorized event
    • Retry on failure (exponential backoff)
    ↓
13. Return Response to Merchant
```

---

## 🔌 API Endpoints

### Authentication
All endpoints require API key authentication:
```
X-API-Key: pk_live_your_api_key_here
```

### Base URL
```
Production: https://api.yourgateway.com
Development: http://localhost:8004
```

---

### POST /api/v1/payments/authorize

Authorize a payment (hold funds without charging).

**Request:**
```json
{
  "amount": 9999,
  "currency": "USD",
  "card": {
    "number": "4242424242424242",
    "cardholder_name": "John Doe",
    "exp_month": 12,
    "exp_year": 2027,
    "cvv": "123"
  },
  "customer": {
    "email": "customer@example.com",
    "name": "John Doe"
  },
  "description": "Order #12345"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "pay_abc123...",
    "status": "authorized",
    "amount": 9999,
    "currency": "USD",
    "card_brand": "visa",
    "card_last4": "4242",
    "auth_code": "123456",
    "fraud_score": 15,
    "fraud_decision": "approve",
    "response_code": "00",
    "response_message": "Approved",
    "transaction_id": "txn_abc123...",
    "created_at": "2025-11-18T10:00:00Z"
  }
}
```

---

### POST /api/v1/payments/sale

Process a sale (authorize + capture in one step).

**Request:** Same as authorize

**Response:** Same as authorize, but `status` will be `"captured"`

---

### POST /api/v1/payments/:id/capture

Capture previously authorized funds.

**Request:**
```json
{
  "amount": 9999
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "pay_abc123...",
    "status": "captured",
    "amount": 9999,
    ...
  }
}
```

---

### POST /api/v1/payments/:id/void

Cancel an authorization before it's captured.

**Request:**
```json
{
  "reason": "Customer requested cancellation"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "pay_abc123...",
    "status": "voided",
    ...
  }
}
```

---

### POST /api/v1/payments/:id/refund

Refund a captured payment.

**Request:**
```json
{
  "amount": 9999,
  "reason": "Product returned"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "pay_abc123...",
    "status": "refunded",
    ...
  }
}
```

---

### GET /api/v1/payments/:id

Retrieve payment details.

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "pay_abc123...",
    "status": "authorized",
    "amount": 9999,
    "currency": "USD",
    "card_brand": "visa",
    "card_last4": "4242",
    "created_at": "2025-11-18T10:00:00Z"
  }
}
```

---

## 🧪 Test Cards

Use these test card numbers for different scenarios:

| Card Number          | Result                | Response Code | Use Case              |
|---------------------|----------------------|---------------|------------------------|
| 4242 4242 4242 4242 | ✅ Approved          | 00            | Successful payment     |
| 5555 5555 5555 4444 | ✅ Approved          | 00            | Mastercard success     |
| 4000 0000 0000 0002 | ❌ Declined          | 05            | Generic decline        |
| 4000 0000 0000 9995 | ❌ Declined          | 51            | Insufficient funds     |
| 4000 0000 0000 0069 | ❌ Declined          | 54            | Expired card           |
| 4000 0000 0000 0127 | ❌ Declined          | N7            | CVV verification failed|
| 4000 0000 0000 0119 | ❌ Declined          | 96            | Processing error       |

**All test cards:**
- Expiry: Any future date
- CVV: Any 3 digits (except 4000...0127 which simulates CVV failure)

---

## 📦 Installation

### Prerequisites
- Go 1.23+
- PostgreSQL 15+
- Redis 7+
- Auth Service (running on port 8001)
- Tokenization Service (running on port 8003)

### Setup

```bash
# 1. Clone repository
cd payment-gateway/payment-api-service

# 2. Install dependencies
go mod download

# 3. Configure environment
cp .env.example .env
# Edit .env with your settings

# 4. Create database
psql -U postgres -c "CREATE DATABASE payment_api_db;"

# 5. Run migrations
go run cmd/migrate/migrate.go

# 6. Start service
go run cmd/main.go

# Or use Air for hot reload
air
```

---

## ⚙️ Configuration

### Environment Variables

```bash
# Server
PORT=8004
GIN_MODE=release

# Database
DATABASE_DSN=postgresql://user:pass@localhost:5432/payment_api_db?sslmode=disable

# Redis
REDIS_DSN=redis://localhost:6379/0

# Dependent Services
AUTH_SERVICE_URL=http://localhost:8001
TOKENIZATION_SERVICE_GRPC=localhost:50051

# Logging
LOG_LEVEL=info  # debug | info | warn | error
```

---

## 📖 Usage Examples

### Example 1: Simple Authorization

```bash
curl -X POST http://localhost:8004/api/v1/payments/authorize \
  -H "X-API-Key: pk_live_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 9999,
    "currency": "USD",
    "card": {
      "number": "4242424242424242",
      "cardholder_name": "John Doe",
      "exp_month": 12,
      "exp_year": 2027,
      "cvv": "123"
    }
  }'
```

### Example 2: Sale with Idempotency

```bash
curl -X POST http://localhost:8004/api/v1/payments/sale \
  -H "X-API-Key: pk_live_your_api_key" \
  -H "Idempotency-Key: order-12345" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 4999,
    "currency": "USD",
    "card": {
      "number": "5555555555554444",
      "cardholder_name": "Jane Smith",
      "exp_month": 6,
      "exp_year": 2028,
      "cvv": "456"
    },
    "customer": {
      "email": "jane@example.com"
    },
    "description": "Premium subscription"
  }'
```

### Example 3: Capture Authorization

```bash
# First, authorize
PAYMENT_ID=$(curl -X POST http://localhost:8004/api/v1/payments/authorize \
  -H "X-API-Key: pk_live_key" \
  -d '{"amount": 5000, "currency": "USD", ...}' \
  | jq -r '.data.id')

# Then capture
curl -X POST http://localhost:8004/api/v1/payments/$PAYMENT_ID/capture \
  -H "X-API-Key: pk_live_key" \
  -H "Content-Type: application/json" \
  -d '{"amount": 5000}'
```

### Example 4: Refund Payment

```bash
curl -X POST http://localhost:8004/api/v1/payments/pay_abc123/refund \
  -H "X-API-Key: pk_live_key" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 2500,
    "reason": "Partial refund - product damaged"
  }'
```

---

## 🔔 Webhooks

### Webhook Events

The Payment API sends webhooks for payment state changes:

| Event                 | Triggered When              |
|----------------------|----------------------------|
| `payment.authorized` | Payment authorized         |
| `payment.captured`   | Payment captured           |
| `payment.voided`     | Authorization voided       |
| `payment.refunded`   | Payment refunded           |
| `payment.failed`     | Payment failed             |

### Webhook Payload

```json
{
  "event": "payment.authorized",
  "timestamp": "2025-11-18T10:00:00Z",
  "id": "evt_abc123...",
  "data": {
    "payment_id": "pay_abc123...",
    "merchant_id": "mer_abc123...",
    "status": "authorized",
    "amount": 9999,
    "currency": "USD",
    "card_brand": "visa",
    "card_last4": "4242",
    "fraud_score": 15,
    "created_at": "2025-11-18T10:00:00Z"
  }
}
```

### Webhook Security

Webhooks include an HMAC-SHA256 signature in the `X-Webhook-Signature` header:

```python
# Verify webhook signature (Python example)
import hmac
import hashlib

def verify_webhook(payload, signature, secret):
    expected = hmac.new(
        secret.encode(),
        payload.encode(),
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(signature, expected)
```

### Webhook Retry Logic

- Failed webhooks are retried with exponential backoff:
  - 1st retry: 5 minutes
  - 2nd retry: 15 minutes
  - 3rd retry: 1 hour
  - 4th retry: 6 hours
- Maximum 5 attempts
- Webhooks expire after 24 hours

---

## ⚠️ Error Handling

### Error Response Format

```json
{
  "success": false,
  "error": "descriptive error message"
}
```

### Common Error Codes

| HTTP Status | Error                          | Cause                           |
|------------|--------------------------------|---------------------------------|
| 400        | Invalid request                | Malformed JSON or missing fields|
| 401        | Invalid API key                | API key not found or inactive   |
| 402        | Insufficient funds             | Card declined (code 51)         |
| 409        | Idempotency key conflict       | Key reused with different data  |
| 422        | Payment cannot be captured     | Payment not in authorized state |
| 429        | Rate limit exceeded            | Too many requests               |
| 500        | Internal server error          | Unexpected server error         |

### Card Decline Reasons

| Response Code | Meaning                  | Action                          |
|--------------|--------------------------|---------------------------------|
| 05           | Do not honor             | Ask customer to contact bank    |
| 51           | Insufficient funds       | Ask for different payment method|
| 54           | Expired card             | Update card expiry date         |
| N7           | CVV verification failed  | Re-enter CVV                    |
| 96           | System error             | Retry request                   |

---

## 📊 Monitoring

### Health Checks

```bash
# Health check
curl http://localhost:8004/health

# Readiness check (checks DB + Redis)
curl http://localhost:8004/ready
```

### Metrics to Track

1. **Payment Success Rate** - Target: > 95%
2. **Average Response Time** - Target: < 1s
3. **Fraud Detection Rate** - Track high-risk blocks
4. **Webhook Delivery Rate** - Target: > 99%
5. **Rate Limit Hits** - Monitor for potential attacks

---

## 🐛 Troubleshooting

### Issue: "Invalid API key"

**Solution:** Verify API key is correct and active:
```bash
# Check API keys in auth service
curl http://localhost:8001/api/v1/api-keys/merchant/{merchant_id} \
  -H "Authorization: Bearer <jwt>"
```

### Issue: "Rate limit exceeded"

**Solution:** Implement exponential backoff or contact support to increase limits.

### Issue: "Idempotency key conflict"

**Solution:** Use unique idempotency keys per request. Keys expire after 24 hours.

### Issue: "Payment cannot be captured"

**Solution:** Verify payment status is "authorized" before capturing:
```bash
curl http://localhost:8004/api/v1/payments/{id} \
  -H "X-API-Key: pk_live_key"
```

---

## 🚀 Development

### Run with Hot Reload

```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Start with hot reload
air
```

### Run Migrations

```bash
go run cmd/migrate/migrate.go
```

---

## 📄 License

Copyright © 2025 Payment Gateway. All rights reserved.

---
## Support

For issues and questions:

- GitHub : https://github.com/rhaloubi/Payment-Gateway-Microservices
- Email: redahaloubi8@gmail.com
---

**Service Version:** v1.0.0  
**API Version:** v1  
**Last Updated:** November 18, 2025