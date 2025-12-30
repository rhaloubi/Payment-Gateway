# 💳 Payment Processing System

**A modern, microservices-based payment gateway built for scale, security, and learning**

> *Enterprise-grade payment processing infrastructure demonstrating real-world microservices architecture, PCI compliance, and distributed systems design.*

---

## 📋 Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [System Architecture](#system-architecture)
- [Services Overview](#services-overview)
- [Technology Stack](#technology-stack)
- [Quick Start](#quick-start)
- [Payment Flow](#payment-flow)
- [API Gateway](#api-gateway)
- [Security & Compliance](#security--compliance)
- [Learning Resources](#learning-resources)
- [Project Status](#project-status)
- [Support](#support)

---

## 🎯 Overview

The Payment Processing System is a complete, production-ready payment gateway implementation showcasing modern microservices architecture. Built as both a functional payment platform and an educational resource, it demonstrates:

- **Microservices Design Patterns** - Service discovery, inter-service communication, data consistency
- **PCI-DSS Compliance** - Card tokenization, encryption at rest, secure key management
- **Financial Operations** - Multi-currency processing, settlements, chargebacks, reconciliation
- **Production Practices** - Rate limiting, idempotency, audit logging, monitoring

### Use Cases

- 🎓 **Learning Platform** - Understand how payment systems work under the hood
- 🏗️ **Architecture Reference** - Real-world microservices patterns and practices
- 💼 **Internal Development** - Foundation for building payment features
- 🔍 **System Design Study** - Distributed systems, event-driven architecture

---

## ✨ Key Features

### Payment Operations
- ✅ **Authorization** - Hold funds without charging (7-day expiry)
- ✅ **Capture** - Charge previously authorized funds (full/partial)
- ✅ **Void** - Cancel authorization before capture
- ✅ **Refund** - Return funds to customer (full/partial)
- ✅ **Payment Intents** - Hosted checkout with browser-friendly flow

### Financial Management
- ✅ **Multi-Currency Support** - USD, EUR, MAD with automatic conversion
- ✅ **Settlement Processing** - Daily batch processing (T+2 settlement)
- ✅ **Processing Fees** - Automatic calculation (2.9% + $0.30)
- ✅ **Chargeback Management** - Complete dispute handling workflow

### Security & Compliance
- ✅ **Card Tokenization** - PCI-compliant token-based system
- ✅ **AES-256-GCM Encryption** - Per-merchant encryption keys
- ✅ **Role-Based Access Control** - Granular permissions system
- ✅ **API Key Management** - Secure merchant authentication
- ✅ **Audit Logging** - Complete transaction history

### Developer Experience
- ✅ **REST & gRPC APIs** - Multiple integration options
- ✅ **Idempotency Support** - Safe retry mechanisms
- ✅ **Rate Limiting** - Per-merchant API limits
- ✅ **Webhook Notifications** - Real-time event delivery
- ✅ **Test Card Support** - Comprehensive testing scenarios

---

## 🏗️ System Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         CLIENT APPLICATIONS                              │
│          (Merchant Dashboard, Mobile Apps, Third-party Integrations)     │
└────────────────────────────────┬────────────────────────────────────────┘
                                 │
                                 │ HTTPS
                                 │
┌────────────────────────────────▼────────────────────────────────────────┐
│                          API GATEWAY (Port 8080)                         │
│                  CORS • Rate Limiting • Request Logging                  │
└────────────────────────────────┬────────────────────────────────────────┘
                                 │
                ┌────────────────┼────────────────┐
                │                │                │
        ┌───────▼──────┐  ┌─────▼──────┐  ┌─────▼──────┐
        │              │  │            │  │            │
        │ Auth Service │  │  Merchant  │  │  Payment   │
        │  (Port 8001) │  │  Service   │  │   API      │
        │              │  │ (Port 8002)│  │ (Port 8004)│
        │              │  │            │  │            │
        └──────┬───────┘  └──────┬─────┘  └──────┬─────┘
               │                 │                │
               │                 │                │ gRPC
               │                 │        ┌───────┴──────────┐
               │                 │        │                  │
        ┌──────▼─────────────────▼────┐   │   ┌──────────────▼────────┐
        │                              │   │   │                       │
        │     PostgreSQL Databases     │   │   │  Tokenization Service │
        │  (Auth, Merchant, Payment)   │   │   │    (Port 8003/50052)  │
        │                              │   │   │                       │
        └──────────────────────────────┘   │   └──────────┬────────────┘
                                           │              │
        ┌──────────────────────────────┐   │              │ gRPC
        │                              │   │              │
        │        Redis Cache           │   │   ┌──────────▼────────────┐
        │  Sessions • Idempotency      │   │   │                       │
        │  Rate Limiting • Cache       │   │   │  Transaction Service  │
        │                              │   │   │    (gRPC: 50053)      │
        └──────────────────────────────┘   │   │                       │
                                           │   └───────────────────────┘
                                           │
                                           └──→ Fraud Detection (Mock)
```

### Service Communication

- **REST APIs** - Public merchant-facing endpoints
- **gRPC** - High-performance internal service communication
- **HTTP/2** - Gateway to service communication
- **Event-Driven** - Webhooks for async notifications

---

## 🔧 Services Overview

### 1. [API Gateway](./api-gateway) 🌐
**Port:** 8080  
**Protocol:** HTTP/REST

Central entry point for all client requests.

**Responsibilities:**
- CORS handling and security headers
- Request/response logging
- Rate limiting per merchant
- Request routing to services
- SSL/TLS termination

**Key Features:**
- Per-route rate limits
- Request ID generation
- Structured logging
- Health check aggregation

---

### 2. [Auth Service](./auth-service) 🔐
**Port:** 8001  
**Protocol:** HTTP/REST

Authentication, authorization, and access control.

**Responsibilities:**
- User registration and login
- JWT token management
- Role-based access control (RBAC)
- API key generation and validation
- Session management

**Key Features:**
- bcrypt password hashing
- 5 failed login attempts = 30min lockout
- Redis-cached permissions
- Multi-tenant role assignments

**[📖 Full Documentation →](./auth-service/README.md)**

---

### 3. [Merchant Service](./merchant-service) 🏪
**Port:** 8002  
**Protocol:** HTTP/REST

Merchant account and team management.

**Responsibilities:**
- Merchant onboarding
- Business profile management
- Team member invitations
- Payment settings configuration
- Webhook configuration

**Key Features:**
- Multi-merchant support per user
- Role-based team permissions
- Invitation system with tokens
- Payment method configuration

**[📖 Full Documentation →](./merchant-service/README.md)**

---

### 4. [Payment API Service](./payment-api-service) 💳
**Port:** 8004  
**Protocol:** HTTP/REST

Public payment processing gateway.

**Responsibilities:**
- Payment orchestration (authorize, capture, void, refund)
- Payment Intents (hosted checkout)
- Idempotency handling
- Webhook delivery
- Transaction status tracking

**Key Features:**
- Single and multi-step payments
- Payment attempt tracking
- Automatic expiration (1 hour)
- Retry logic for webhooks
- Test card support

**[📖 Full Documentation →](./payment-api-service/README.md)**

---

### 5. [Tokenization Service](./tokenization-service) 🔐
**Port:** 8003 (REST), 50052 (gRPC)  
**Protocol:** HTTP/REST + gRPC

PCI-compliant card tokenization.

**Responsibilities:**
- Card data encryption (AES-256-GCM)
- Token generation and validation
- Card fingerprinting
- Key rotation management
- BIN database lookups

**Key Features:**
- Per-merchant encryption keys
- Duplicate card detection
- Single-use tokens
- 90-day automatic key rotation
- Luhn validation

**[📖 Full Documentation →](./tokenization-service/README.md)**

---

### 6. [Transaction Service](./transaction-service) 🔄
**Port:** 50053  
**Protocol:** gRPC (Internal Only)

Core payment transaction engine.

**Responsibilities:**
- Transaction lifecycle management
- Card simulator (test mode)
- Multi-currency conversion
- Settlement batch processing
- Chargeback handling

**Key Features:**
- Transaction state machine
- Daily settlement (T+2)
- Auto-void expired authorizations
- Currency conversion (USD, EUR, MAD)
- Processing fee calculation

**[📖 Full Documentation →](./transaction-service/README.md)**

---

### 7. Fraud Detection Service 🛡️
**Status:** Mock Implementation

Basic risk scoring for transactions.

**Current Implementation:**
- Random risk scores (0-100)
- Simple rule-based decisions
- Decline if score > 70

**Future Plans:**
- Machine learning models
- Velocity checks
- Device fingerprinting
- IP geolocation
- Historical pattern analysis

---

## 💻 Technology Stack

### Backend Services
- **Language:** Go 1.23+
- **HTTP Framework:** Gin
- **gRPC:** Protocol Buffers + gRPC
- **ORM:** GORM

### Databases & Cache
- **PostgreSQL 15+** - Transactional data
- **Redis 7+** - Caching, sessions, rate limiting

### Security
- **JWT** - golang-jwt/jwt
- **Encryption** - AES-256-GCM
- **Hashing** - bcrypt, SHA-256

### Development Tools
- **Air** - Hot reload for Go
- **Docker** - Containerization (Dockerfiles available)
- **Protocol Buffers** - gRPC service definitions

### Infrastructure (In Progress)
- **Kubernetes** - Container orchestration
- **AWS** - Cloud hosting

---

## 🚀 Quick Start

### Prerequisites

```bash
# Required
- Go 1.23+
- PostgreSQL 15+
- Redis 7+
- Docker (optional)

# Recommended
- Air (hot reload)
- grpcurl (gRPC testing)
```

### 1. Clone Repository

```bash
git clone https://github.com/rhaloubi/Payment-Gateway-Microservices.git
cd Payment-Gateway-Microservices
```

### 2. Start Infrastructure

```bash
# Option 1: Docker Compose (coming soon)
docker-compose up -d postgres redis

# Option 2: Local installations
# Start PostgreSQL on port 5432
# Start Redis on port 6379
```

### 3. Create Databases

```bash
psql -U postgres -c "CREATE DATABASE auth_db;"
psql -U postgres -c "CREATE DATABASE merchant_db;"
psql -U postgres -c "CREATE DATABASE tokenization_db;"
psql -U postgres -c "CREATE DATABASE payment_api_db;"
psql -U postgres -c "CREATE DATABASE transaction_db;"
```

### 4. Configure Services

Each service needs a `.env` file. See individual service READMEs for details.

```bash
# Example: Auth Service
cd auth-service
cp .env.example .env
# Edit .env with your database credentials
```

### 5. Run Migrations

```bash
# Run for each service
cd auth-service && go run cmd/migrate up
cd ../merchant-service && go run cmd/migrate up
cd ../tokenization-service && go run cmd/migrate up
cd ../payment-api-service && go run cmd/migrate up
cd ../transaction-service && go run cmd/migrate up
```

### 6. Start Services

```bash
# Terminal 1: Auth Service
cd auth-service
air # or: go run cmd/main.go

# Terminal 2: Merchant Service
cd merchant-service
air

# Terminal 3: Tokenization Service
cd tokenization-service
air

# Terminal 4: Transaction Service
cd transaction-service
air

# Terminal 5: Payment API Service
cd payment-api-service
air

# Terminal 6: API Gateway
cd api-gateway
air
```

### 7. Verify Health

```bash
# Check all services are running
curl http://localhost:8080/health  # API Gateway
curl http://localhost:8001/health  # Auth Service
curl http://localhost:8002/health  # Merchant Service
curl http://localhost:8003/health  # Tokenization Service
curl http://localhost:8004/health  # Payment API Service
```

---

## 💳 Payment Flow

### Complete Payment Journey

```
1. Merchant Registration
   ↓
   POST /api/v1/auth/register
   POST /api/v1/auth/login
   ↓
2. Create Merchant Profile
   ↓
   POST /api/v1/merchants
   ↓
3. Generate API Key
   ↓
   POST /api/v1/api-keys
   ↓
4. Process Payment
   ↓
   POST /api/v1/payments/authorize
   │
   ├─→ Validate API Key (Auth Service)
   ├─→ Tokenize Card (Tokenization Service)
   ├─→ Check Fraud (Fraud Service - Mock)
   ├─→ Authorize Transaction (Transaction Service)
   └─→ Return Payment ID
   ↓
5. Capture Payment
   ↓
   POST /api/v1/payments/{id}/capture
   │
   └─→ Charge Card (Transaction Service)
   ↓
6. Settlement (Daily at Midnight)
   ↓
   Transaction Service creates settlement batch
   ↓
7. Transfer Funds (T+2)
   ↓
   Merchant receives payout
```

### Example: Authorization Flow

```bash
# 1. Register and Login
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@merchant.com",
    "password": "SecurePass123!"
  }'

# 2. Create Merchant
curl -X POST http://localhost:8080/api/v1/merchants \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Acme Corp",
    "email": "billing@acme.com"
  }'

# 3. Generate API Key
curl -X POST http://localhost:8080/api/v1/api-keys \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "merchant_id": "<merchant_uuid>",
    "name": "Production Key"
  }'

# 4. Process Payment
curl -X POST http://localhost:8080/api/v1/payments/authorize \
  -H "X-API-Key: pk_live_..." \
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

---

## 🌐 API Gateway

The API Gateway serves as the single entry point for all client requests.

### Features

**Security:**
- CORS policy enforcement
- Rate limiting (per merchant)
- Request ID generation
- Security headers (HSTS, CSP, etc.)

**Observability:**
- Request/response logging
- Latency tracking
- Error monitoring
- Health check aggregation

**Routing:**
```
/api/v1/auth/*          → Auth Service (8001)
/api/v1/merchants/*     → Merchant Service (8002)
/api/v1/payments/*      → Payment API Service (8004)
/api/v1/tokenize        → Tokenization Service (8003)
```

**Rate Limits:**
- Authentication: 20 req/sec
- Payments: 50 req/sec
- Default: 100 req/sec

---

## 🔒 Security & Compliance

### PCI-DSS Compliance

**Card Data Protection:**
- ✅ Card numbers never logged or stored in plaintext
- ✅ Tokenization reduces PCI scope
- ✅ AES-256-GCM encryption at rest
- ✅ TLS 1.3 for data in transit

**Access Control:**
- ✅ Role-based access control (RBAC)
- ✅ Principle of least privilege
- ✅ API key rotation
- ✅ Session management

**Audit & Compliance:**
- ✅ Complete audit trail
- ✅ Transaction event logging
- ✅ Failed access attempt tracking
- ✅ Key rotation policies

### Security Best Practices

**Authentication:**
- JWT tokens (24-hour expiry)
- bcrypt password hashing (cost 10)
- 5 failed attempts = 30-minute lockout
- Secure session management

**Encryption:**
- Per-merchant encryption keys
- Automatic key rotation (90 days)
- Secure key storage (Vault planned)
- Field-level encryption

**API Security:**
- Idempotency keys (24-hour cache)
- Rate limiting per merchant
- HMAC webhook signatures
- CORS and CSP headers

---

## 📚 Learning Resources

### Understanding the System

This project is designed as a learning resource for:

**1. Microservices Architecture**
- Service decomposition strategies
- Inter-service communication (REST vs gRPC)
- Data consistency patterns
- Service discovery and routing

**2. Payment Processing**
- Transaction lifecycle management
- Authorization vs capture flows
- Settlement and reconciliation
- Chargeback handling

**3. Distributed Systems**
- Idempotency and retries
- Event-driven architecture
- Cache strategies (Redis)
- State machine design

**4. Security & Compliance**
- PCI-DSS requirements
- Tokenization and encryption
- Key management
- Audit logging

### Recommended Reading Order

1. **Start Here:** [Auth Service](./auth-service/README.md) - Understand authentication
2. **Then:** [Merchant Service](./merchant-service/README.md) - Learn multi-tenancy
3. **Next:** [Tokenization Service](./tokenization-service/README.md) - See PCI compliance
4. **After:** [Payment API Service](./payment-api-service/README.md) - Orchestration patterns
5. **Finally:** [Transaction Service](./transaction-service/README.md) - Core processing logic

---

## 📊 Project Status

### ✅ Completed Features

- [x] Complete authentication and authorization system
- [x] Merchant onboarding and management
- [x] Card tokenization (PCI-compliant)
- [x] Payment processing (authorize, capture, void, refund)
- [x] Payment Intents (hosted checkout)
- [x] Multi-currency support
- [x] Settlement processing
- [x] API Gateway with rate limiting
- [x] Webhook delivery system
- [x] Comprehensive API documentation

### 🚧 In Progress

- [ ] Kubernetes deployment manifests
- [ ] AWS infrastructure setup
- [ ] Docker Compose orchestration
- [ ] CLI tool for merchants ([README](./README.md))

### 🔮 Future Enhancements

- [ ] Machine learning fraud detection
- [ ] Advanced reporting and analytics
- [ ] Recurring payments and subscriptions
- [ ] 3D Secure integration
- [ ] Apple Pay / Google Pay support
- [ ] Multi-region deployment
- [ ] GraphQL API option

---

## 🛠️ Development

### Service Dependencies

```
Auth Service (8001)
  └─→ PostgreSQL
  └─→ Redis

Merchant Service (8002)
  ├─→ PostgreSQL
  ├─→ Redis
  └─→ Auth Service (HTTP + gRPC)

Tokenization Service (8003/50052)
  ├─→ PostgreSQL
  └─→ Redis

Payment API Service (8004)
  ├─→ PostgreSQL
  ├─→ Redis
  ├─→ Auth Service (HTTP)
  ├─→ Tokenization Service (gRPC)
  └─→ Transaction Service (gRPC)

Transaction Service (50053)
  ├─→ PostgreSQL
  ├─→ Redis
  └─→ Tokenization Service (gRPC)

API Gateway (8080)
  └─→ All HTTP Services
```

### Adding a New Service

1. Create service directory: `mkdir new-service`
2. Initialize Go module: `go mod init github.com/rhaloubi/payment-gateway/new-service`
3. Add dependencies and implement handlers
4. Update API Gateway routing
5. Add health checks
6. Document API in README
7. Add to this main README

---

## 🧪 Testing

### Test Cards

| Card Number          | Result      | Use Case              |
|---------------------|-------------|-----------------------|
| 4242 4242 4242 4242 | ✅ Approved | Successful payment    |
| 5555 5555 5555 4444 | ✅ Approved | Mastercard success    |
| 4000 0000 0000 0002 | ❌ Declined | Generic decline       |
| 4000 0000 0000 9995 | ❌ Declined | Insufficient funds    |
| 4000 0000 0000 0069 | ❌ Declined | Expired card          |
| 4000 0000 0000 0127 | ❌ Declined | CVV verification fail |

**For all test cards:**
- Use any future expiry date (e.g., 12/2027)
- Use any 3-digit CVV (except 0127 for CVV fail test)

---

## 📞 Support

### Getting Help

- **GitHub Issues:** [Report bugs or request features](https://github.com/rhaloubi/Payment-Gateway-Microservices/issues)
- **Email:** redahaloubi8@gmail.com
- **Documentation:** Check individual service READMEs for detailed guides

### Contributing

This project is primarily for learning and internal development. If you'd like to contribute:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

---

## 📄 License

Copyright © 2025 Payment Gateway. All rights reserved.

---

## 🙏 Acknowledgments

Built with:
- Go (Golang)
- PostgreSQL
- Redis
- Gin Framework
- gRPC
- Docker

---

**Project Version:** v1.0.0  
**Last Updated:** December 2025  
**Maintained by:** [Reda Haloubi](https://github.com/rhaloubi)

---

## Quick Links

- [Auth Service →](./auth-service/README.md)
- [Merchant Service →](./merchant-service/README.md)
- [Tokenization Service →](./tokenization-service/README.md)
- [Payment API Service →](./payment-api-service/README.md)
- [Transaction Service →](./transaction-service/README.md)
- [CLI Tool →](./README.md)

**Happy Building! 🚀**