# Payment Gateway Microservices Architecture

## Educational Fintech Platform - Multi-Tenant Payment Processing System

> **⚠️ LEARNING PROJECT DISCLAIMER**: This architecture is designed for educational purposes using fake/test data only. Not intended for production use with real payment card data without full PCI DSS compliance and proper licensing.

---

## Table of Contents

1. [System Overview](#system-overview)
2. [Microservices Architecture](#microservices-architecture)
3. [Security Architecture](#security-architecture)
4. [Database Design](#database-design)
5. [API Design](#api-design)
6. [Infrastructure & DevOps](#infrastructure--devops)
7. [Compliance Considerations](#compliance-considerations)
8. [Development Roadmap](#development-roadmap)

---

## System Overview

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Load Balancer / API Gateway             │
│                         (Kong / Traefik / Nginx)                │
└─────────────────────────────────────────────────────────────────┘
                                   │
                ┌──────────────────┼──────────────────┐
                │                  │                  │
        ┌───────▼────────┐ ┌──────▼─────┐  ┌────────▼────────┐
        │  Auth Service  │ │  Merchant  │  │  Payment API    │
        │   (Public)     │ │  Service   │  │   (Public)      │
        └────────────────┘ └────────────┘  └─────────────────┘
                │                  │                  │
                └──────────────────┼──────────────────┘
                                   │
                         ┌─────────▼─────────┐
                         │  Message Broker   │
                         │     (Kafka)       │
                         └─────────┬─────────┘
                                   │
        ┌──────────────────────────┼──────────────────────────┐
        │                          │                          │
┌───────▼──────┐         ┌─────────▼────────┐      ┌─────────▼────────┐
│ Transaction  │         │   Tokenization   │      │  Fraud Detection │
│   Service    │         │     Service      │      │     Service      │
└──────────────┘         └──────────────────┘      └──────────────────┘
        │                          │                          │
        └──────────────────────────┼──────────────────────────┘
                                   │
        ┌──────────────────────────┼──────────────────────────┐
        │                          │                          │
┌───────▼──────┐         ┌─────────▼────────┐      ┌─────────▼────────┐
│  Settlement  │         │   3DS Service    │      │   Notification   │
│   Service    │         │                  │      │     Service      │
└──────────────┘         └──────────────────┘      └──────────────────┘
        │                          │                          │
┌───────▼──────┐         ┌─────────▼────────┐      ┌─────────▼────────┐
│   Batching   │         │    Invoice       │      │    Webhook       │
│   Service    │         │    Service       │      │    Service       │
└──────────────┘         └──────────────────┘      └──────────────────┘
        │                          │                          │
┌───────▼──────┐         ┌─────────▼────────┐      ┌─────────▼────────┐
│  Recurring   │         │   Reporting      │      │   Audit/Logging  │
│  Payments    │         │   Service        │      │     Service      │
└──────────────┘         └──────────────────┘      └──────────────────┘
```

### Tech Stack

- **Backend**: Go (Golang)
- **Frontend**: Next.js (React)
- **Database**: PostgreSQL (with encryption at rest)
- **Message Queue**: Apache Kafka
- **Cache**: Redis
- **Container Orchestration**: Kubernetes (K8s)
- **Service Mesh**: Istio (optional but recommended)
- **Secret Management**: HashiCorp Vault

---

## Microservices Architecture

### 1. **Auth Service** (Gateway Entry Point)

**Purpose**: Authentication, authorization, and API key management

**Responsibilities**:

- Merchant authentication (OAuth 2.0 / JWT)
- API key generation and validation
- Role-based access control (RBAC)
- Session management
- Token refresh mechanisms

**Tech Details**:

- JWT with RS256 signing
- API keys with HMAC-SHA256
- Rate limiting per merchant
- IP whitelisting support

---

### 2. **Merchant Service** (Multi-Tenancy Management)

**Purpose**: Merchant onboarding, configuration, and management

**Responsibilities**:

- Merchant registration and KYC (Know Your Customer) - simplified
- Merchant profile management
- Business settings and configurations
- Fee structure management
- Currency and payment method settings

### 3. **Payment API Service** (Public Entry Point)

**Purpose**: Main payment processing interface for merchants

**Responsibilities**:

- Accept payment requests from merchants
- Validate payment data
- Route to appropriate internal services
- Return payment status synchronously
- Handle hosted payment page requests

**Key Features**:

- RESTful API design
- Input validation and sanitization
- Idempotency key support
- Request/response logging
- Error handling with PCI-safe messages (no card data in errors)

---

### 4. **Tokenization Service** (PCI DSS Scope Reduction)

**Purpose**: Convert sensitive card data to tokens

**Responsibilities**:

- Generate tokens for card data
- Securely store card data in encrypted vault
- Retrieve card data using tokens (only for processing)
- Token lifecycle management
- Card verification value (CVV) handling (never stored)

**Security Requirements**:

- AES-256 encryption for card data at rest
- Separate encryption keys per merchant (DEK - Data Encryption Key)
- Master key in Vault (KEK - Key Encryption Key)
- Field-level encryption
- HSM integration (Hardware Security Module) - simulated for learning

**Token Format**:

```
tok_live_4xJ3kL9mN2pQ5rT8vW1yZ4bC7dE0fG
tok_test_1aB2cD3eF4gH5iJ6kL7mN8oP9qR0sT
```

---

### 5. **Transaction Service** (Core Processing Engine)

**Purpose**: Process all payment transactions

**Responsibilities**:

- Execute authorization requests
- Process sales (auth + capture)
- Handle captures of authorized amounts
- Process voids/reversals
- Handle refunds (full and partial)
- Transaction state management
- Interact with payment networks (simulated)

**Transaction States**:

```
PENDING → AUTHORIZED → CAPTURED → SETTLED
       ↓              ↓           ↓
    FAILED        VOIDED      REFUNDED
```

**Kafka Topics** (Producer):

```
transaction.created
transaction.authorized
transaction.captured
transaction.voided
transaction.refunded
transaction.failed
```

---

### 6. **Fraud Detection Service** (Risk Management)

**Purpose**: Real-time fraud analysis and prevention

**Responsibilities**:

- Velocity checks (transactions per time period)
- Amount thresholds
- Geolocation checks
- Device fingerprinting
- Behavioral analysis
- Risk scoring (0-100)
- Block/allow lists

**Fraud Rules Engine**:

- Rule-based scoring
- Machine learning models (simple ML for learning - logistic regression)
- Real-time decision making (<100ms)

**Kafka Topics** (Consumer/Producer):

```
Consumer: transaction.created
Producer: fraud.detected, fraud.score.calculated
```

**Decision Flow**:

```
Transaction → Fraud Check → Score (0-100)
  ↓
Score < 30: APPROVE (low risk)
Score 30-70: REVIEW (medium risk)
Score > 70: DECLINE (high risk)
```

---

### 7. **3DS Service** (3D Secure Authentication)

**Purpose**: Implement 3D Secure 2.0 protocol for SCA (Strong Customer Authentication)

**Responsibilities**:

- Initiate 3DS authentication flow
- Communicate with 3DS server/ACS (simulated)
- Handle challenge/frictionless flows
- Store authentication results
- Liability shift management

**3DS Flow**:

```
1. Merchant Request → 2. Directory Lookup → 3. ACS Challenge
   ↓                      ↓                      ↓
4. Customer Auth → 5. Validation → 6. Result to Merchant
```

**Database Tables**:

- `three_ds_authentications`
- `three_ds_challenges`

**Endpoints**:

```
POST   /v1/3ds/initiate
GET    /v1/3ds/{auth_id}/status
POST   /v1/3ds/{auth_id}/complete
```

---

### 8. **Settlement/Batching Service**

**Purpose**: Group transactions for settlement processing

**Responsibilities**:

- Create transaction batches
- Manual batch creation
- Automatic batch scheduling (daily, weekly)
- Calculate batch totals
- Generate settlement reports
- Mark transactions as settled

**Batch Types**:

- **Manual**: Merchant-initiated
- **Automatic**: Scheduled (cron-based)
- **Partial**: Selected transactions

**Database Tables**:

- `batches`
- `batch_transactions`
- `settlement_reports`

**Kafka Topics**:

```
Consumer: transaction.captured
Producer: batch.created, batch.settled
```

**Endpoints**:

```
POST   /v1/batches/create
GET    /v1/batches/{batch_id}
POST   /v1/batches/{batch_id}/settle
GET    /v1/batches?status=pending
```

---

### 9. **Recurring Payments Service**

**Purpose**: Handle subscription and installment payments

**Responsibilities**:

- Create payment plans (subscriptions/installments)
- Schedule recurring charges
- Process scheduled payments automatically
- Handle failed payment retries
- Subscription lifecycle management (pause, cancel, resume)
- Dunning management (retry logic)

**Payment Plan Types**:

- **Subscription**: Ongoing, until cancelled
- **Installment**: Fixed number of payments

**Database Tables**:

- `payment_plans`
- `plan_schedules`
- `recurring_transactions`
- `failed_payment_retries`

**Kafka Topics**:

```
Producer: plan.created, plan.charged, plan.failed, plan.cancelled
```

**Endpoints**:

```
POST   /v1/recurring/plans
GET    /v1/recurring/plans/{plan_id}
PATCH  /v1/recurring/plans/{plan_id}/pause
PATCH  /v1/recurring/plans/{plan_id}/cancel
POST   /v1/recurring/plans/{plan_id}/charge-now
```

---

### 10. **Invoice Service**

**Purpose**: Generate and manage payment invoices

**Responsibilities**:

- Create invoices with line items
- Generate payment links
- Track invoice status (draft, sent, paid, overdue)
- Calculate totals, taxes, discounts
- Support manual payment recording

**Database Tables**:

- `invoices`
- `invoice_items`
- `invoice_payments`

**Endpoints**:

```
POST   /v1/invoices
GET    /v1/invoices/{invoice_id}
PATCH  /v1/invoices/{invoice_id}
POST   /v1/invoices/{invoice_id}/send
POST   /v1/invoices/{invoice_id}/record-payment
GET    /v1/invoices/{invoice_id}/payment-link
```

---

### 11. **Notification Service**

**Purpose**: Send notifications via email and SMS

**Responsibilities**:

- Email notifications (invoices, receipts, alerts)
- SMS notifications (OTP, payment confirmations)
- Template management
- Delivery tracking
- Queue management

**Technologies**:

- Email: SMTP / SendGrid / Amazon SES (configurable)
- SMS: Twilio / MessageBird (configurable)
- Template engine: Go templates / Handlebars

**Database Tables**:

- `notification_templates`
- `notification_logs`

**Kafka Topics** (Consumer):

```
invoice.created, transaction.completed, payment.failed
```

---

### 12. **Webhook Service**

**Purpose**: Send event notifications to merchant systems

**Responsibilities**:

- Deliver webhooks to merchant endpoints
- Webhook signing (HMAC-SHA256)
- Retry logic with exponential backoff
- Webhook delivery logs
- Merchant webhook configuration

**Security**:

- HMAC signature in headers: `X-Signature`
- Timestamp validation (prevent replay attacks)
- HTTPS-only endpoints

**Database Tables**:

- `webhook_endpoints`
- `webhook_deliveries`
- `webhook_events`

**Retry Strategy**:

```
Attempt 1: Immediate
Attempt 2: After 1 minute
Attempt 3: After 5 minutes
Attempt 4: After 30 minutes
Attempt 5: After 2 hours
Max Attempts: 5
```

**Kafka Topics** (Consumer):

```
All domain events (transaction.*, invoice.*, etc.)
```

---

### 13. **Audit/Logging Service**

**Purpose**: Comprehensive audit trail and security logging

**Responsibilities**:

- Log all API requests/responses
- Log all state changes
- Security event logging
- PCI DSS audit requirements
- Sensitive data masking in logs
- Log retention and archival

**Log Types**:

- **Access Logs**: Who accessed what, when
- **Transaction Logs**: All transaction events
- **Security Logs**: Failed auth attempts, anomalies
- **Admin Logs**: Configuration changes

**Database Tables**:

- `audit_logs`
- `security_events`
- `api_access_logs`

**Log Format** (Structured JSON):

```json
{
  "timestamp": "2025-10-31T14:30:00Z",
  "level": "INFO",
  "service": "transaction-service",
  "merchant_id": "mch_123",
  "event": "transaction.authorized",
  "transaction_id": "txn_456",
  "amount": 99.99,
  "currency": "USD",
  "masked_card": "****1234",
  "ip_address": "192.168.1.1",
  "user_agent": "...",
  "trace_id": "abc-def-ghi"
}
```

---

### 14. **Reporting Service**

**Purpose**: Generate business intelligence and reports

**Responsibilities**:

- Transaction reports (daily, weekly, monthly)
- Revenue analytics
- Chargeback reports
- Settlement summaries
- Custom report generation
- Data export (CSV, PDF)

**Database**:

- Read replicas for reporting queries
- OLAP database (optional - ClickHouse / TimescaleDB)

**Endpoints**:

```
GET    /v1/reports/transactions
GET    /v1/reports/revenue
GET    /v1/reports/settlements
POST   /v1/reports/custom
GET    /v1/reports/{report_id}/download
```

---

### 15. **Virtual Terminal (Frontend)**

**Purpose**: Web-based interface for merchants

**Tech Stack**: Next.js, React, TypeScript, TailwindCSS

**Features**:

- Dashboard with key metrics
- Manual transaction entry
- Transaction search and filtering
- Refund/void management
- Invoice creation and management
- Recurring payment setup
- Customer management
- Reporting and analytics
- Settings and configuration
- Webhook management

**Pages**:

```
/dashboard
/transactions
/transactions/new
/customers
/invoices
/invoices/new
/recurring
/batches
/reports
/settings
/api-keys
```

---

## Security Architecture

### 1. **Encryption Strategy**

#### Data at Rest

- **Database Encryption**: PostgreSQL TDE (Transparent Data Encryption)
- **Card Data**: AES-256-GCM encryption
- **Encryption Keys**: Stored in HashiCorp Vault
- **Key Rotation**: Every 90 days

#### Data in Transit

- **TLS 1.3**: All external communications
- **mTLS**: Service-to-service communication
- **Certificate Management**: Let's Encrypt / cert-manager

#### Field-Level Encryption

```go
// Example: Encrypt card number
type CardData struct {
    CardNumber     string // Encrypted
    CardholderName string // Encrypted
    ExpiryMonth    int    // Encrypted
    ExpiryYear     int    // Encrypted
    CVV            string // Never stored
    Token          string // Plain text token reference
}
```

---

### 2. **Key Management Architecture**

```
┌─────────────────────────────────────────┐
│         HashiCorp Vault (KEK)           │
│  Master Encryption Key (KEK)            │
└────────────┬────────────────────────────┘
             │
    ┌────────┴────────┐
    │                 │
┌───▼────┐      ┌────▼────┐
│ DEK 1  │      │ DEK 2   │  (Data Encryption Keys)
│Merchant│      │Merchant │  Per-merchant encryption
│  123   │      │  456    │
└───┬────┘      └────┬────┘
    │                │
┌───▼─────────────────▼────┐
│   Encrypted Card Data    │
│   (PostgreSQL Database)  │
└──────────────────────────┘
```

**Key Hierarchy**:

- **KEK** (Key Encryption Key): Stored in Vault, encrypts DEKs
- **DEK** (Data Encryption Key): Per-merchant, encrypts actual data
- **Token**: Public reference, no cryptographic value

---

### 3. **Authentication & Authorization**

#### Merchant Authentication (Virtual Terminal)

- **Method**: JWT tokens with RS256
- **Token Expiry**: 15 minutes (access token)
- **Refresh Token**: 7 days
- **Storage**: HttpOnly cookies (refresh token)

#### API Authentication (Server-to-Server)

- **Method**: API Keys with HMAC-SHA256
- **Format**: `pk_live_{random_32_chars}` or `sk_live_{random_32_chars}`
- **Rate Limiting**: Per API key
- **IP Whitelisting**: Optional per merchant

#### Authorization (RBAC)

```
Roles:
- MERCHANT_ADMIN: Full access
- MERCHANT_USER: Read + Create transactions
- MERCHANT_VIEWER: Read-only access
- DEVELOPER: API key management
```

---

### 4. **Network Security**

#### Service Mesh (Istio)

- Mutual TLS between services
- Traffic encryption
- Service-to-service authentication
- Traffic policies and routing

#### API Gateway

- Rate limiting (per merchant, per IP)
- Request validation
- DDoS protection
- IP whitelisting/blacklisting
- WAF (Web Application Firewall)

#### Network Segmentation

```
┌─────────────────────────────────────────┐
│         DMZ (Public Services)           │
│  - API Gateway                          │
│  - Load Balancer                        │
└─────────────┬───────────────────────────┘
              │
┌─────────────▼───────────────────────────┐
│     Application Zone (Private)          │
│  - All microservices                    │
└─────────────┬───────────────────────────┘
              │
┌─────────────▼───────────────────────────┐
│         Data Zone (Most Secure)         │
│  - PostgreSQL                           │
│  - Vault                                │
│  - Redis (encrypted)                    │
└─────────────────────────────────────────┘
```

---

### 5. **Input Validation & Sanitization**

#### Card Number Validation

```go
// Luhn algorithm
func ValidateCardNumber(number string) bool {
    // Remove spaces and dashes
    // Check length (13-19 digits)
    // Perform Luhn check
}

// Card type detection
func DetectCardType(number string) string {
    // Visa: starts with 4
    // Mastercard: starts with 51-55 or 2221-2720
    // Amex: starts with 34 or 37
}
```

#### SQL Injection Prevention

- Parameterized queries (prepared statements)
- ORM with query builders (gorm / sqlx)
- No dynamic SQL string concatenation

#### XSS Prevention

- Content Security Policy (CSP) headers
- Input sanitization
- Output encoding
- React's built-in XSS protection

---

### 6. **PCI DSS Key Requirements (Educational Implementation)**

Even for learning, implement these concepts:

1. **Build and Maintain Secure Network**

   - Firewall configuration
   - No default passwords

2. **Protect Cardholder Data**

   - Encryption at rest and in transit
   - Mask PAN when displayed (show only last 4 digits)
   - Never log full PAN, CVV, or PIN

3. **Maintain Vulnerability Management**

   - Regular security updates
   - Secure coding practices
   - Code reviews

4. **Implement Strong Access Control**

   - Unique IDs per user
   - Need-to-know access
   - Physical/logical access controls

5. **Monitor and Test Networks**

   - Audit logs
   - Security testing
   - Intrusion detection

6. **Information Security Policy**
   - Document security policies
   - Security awareness

**Data Retention**:

```
- Full PAN: Only during transaction processing
- Masked PAN (last 4): Can be stored
- CVV: NEVER store after authorization
- Transaction logs: 7 years (compliance)
- Audit logs: 3 years minimum
```

---

### Database Security Best Practices

1. **Connection Security**

   - SSL/TLS connections only
   - Certificate-based authentication
   - Connection pooling with max connections

2. **Access Control**

   - Separate DB users per service (least privilege)
   - No root/admin access from applications
   - Use connection strings from secrets manager

3. **Encryption**

   - TDE (Transparent Data Encryption) enabled
   - Field-level encryption for sensitive data
   - Encrypted backups

4. **Backup Strategy**

   - Daily automated backups
   - Point-in-time recovery enabled
   - Backup retention: 30 days
   - Encrypted backup storage

5. **Monitoring**
   - Query performance monitoring
   - Connection pool monitoring
   - Slow query logging
   - Anomaly detection

---

## API Design

### REST API Conventions

#### Base URL

```
Production:  https://api.yourgateway.com
Staging:     https://api.staging.yourgateway.com
Development: http://localhost:8080
```

#### Versioning

```
/v1/payments
/v1/invoices
```

#### Authentication Headers

```http
Authorization: Bearer <JWT_TOKEN>
X-API-Key: pk_live_xxxxx
```

#### Standard Response Format

**Success Response**:

```json
{
  "success": true,
  "data": {
    "id": "txn_abc123",
    "amount": 99.99,
    "currency": "USD",
    "status": "authorized"
  },
  "meta": {
    "request_id": "req_xyz789",
    "timestamp": "2025-10-31T14:30:00Z"
  }
}
```

**Error Response**:

```json
{
  "success": false,
  "error": {
    "code": "INVALID_CARD",
    "message": "The card number is invalid",
    "type": "card_error",
    "param": "card_number"
  },
  "meta": {
    "request_id": "req_xyz789",
    "timestamp": "2025-10-31T14:30:00Z"
  }
}
```

#### Error Codes

```
AUTHENTICATION_FAILED
INVALID_API_KEY
INSUFFICIENT_PERMISSIONS
INVALID_REQUEST
INVALID_CARD
CARD_DECLINED
INSUFFICIENT_FUNDS
FRAUD_DETECTED
DUPLICATE_TRANSACTION
RESOURCE_NOT_FOUND
RATE_LIMIT_EXCEEDED
INTERNAL_ERROR
```

---

### Key API Endpoints

#### 1. Payment Processing

**Authorize Payment**

```http
POST /v1/payments/authorize
Content-Type: application/json
X-API-Key: pk_live_xxxxx
Idempotency-Key: unique_key_123

{
  "amount": 99.99,
  "currency": "USD",
  "payment_method": {
    "type": "card",
    "card": {
      "number": "4242424242424242",
      "exp_month": 12,
      "exp_year": 2027,
      "cvc": "123",
      "cardholder_name": "John Doe"
    }
  },
  "customer": {
    "email": "customer@example.com",
    "name": "John Doe",
    "phone": "+1234567890"
  },
  "billing_address": {
    "line1": "123 Main St",
    "city": "New York",
    "state": "NY",
    "postal_code": "10001",
    "country": "US"
  },
  "description": "Order #12345",
  "metadata": {
    "order_id": "12345",
    "customer_id": "cust_abc"
  }
}
```

**Response**:

```json
{
  "success": true,
  "data": {
    "id": "txn_abc123",
    "type": "authorize",
    "status": "authorized",
    "amount": 99.99,
    "currency": "USD",
    "authorized_amount": 99.99,
    "card": {
      "brand": "visa",
      "last4": "4242",
      "exp_month": 12,
      "exp_year": 2027
    },
    "token": "tok_live_xyz789",
    "fraud_score": 15,
    "fraud_status": "approved",
    "three_ds": {
      "authenticated": true,
      "liability_shift": true
    },
    "created_at": "2025-10-31T14:30:00Z"
  }
}
```

**Capture Authorized Payment**

```http
POST /v1/payments/{transaction_id}/capture
Content-Type: application/json
X-API-Key: pk_live_xxxxx

{
  "amount": 99.99 // Optional: partial capture
}
```

**Void Transaction**

```http
POST /v1/payments/{transaction_id}/void
Content-Type: application/json
X-API-Key: pk_live_xxxxx

{
  "reason": "Customer requested cancellation"
}
```

**Refund Transaction**

```http
POST /v1/payments/{transaction_id}/refund
Content-Type: application/json
X-API-Key: pk_live_xxxxx

{
  "amount": 49.99, // Optional: partial refund
  "reason": "Product returned"
}
```

---

#### 2. Invoice Management

**Create Invoice**

```http
POST /v1/invoices
Content-Type: application/json
X-API-Key: pk_live_xxxxx

{
  "customer": {
    "email": "customer@example.com",
    "name": "John Doe"
  },
  "items": [
    {
      "description": "Web Development Services",
      "quantity": 10,
      "unit_price": 150.00
    },
    {
      "description": "Hosting (Monthly)",
      "quantity": 1,
      "unit_price": 29.99
    }
  ],
  "tax_amount": 153.00,
  "due_date": "2025-11-30",
  "notes": "Payment due within 30 days",
  "send_email": true
}
```

---

#### 3. Recurring Payments

**Create Subscription**

```http
POST /v1/recurring/plans
Content-Type: application/json
X-API-Key: pk_live_xxxxx

{
  "type": "subscription",
  "amount": 29.99,
  "currency": "USD",
  "frequency": "monthly",
  "interval": 1,
  "start_date": "2025-11-01",
  "payment_method": {
    "token": "tok_live_xyz789"
  },
  "customer": {
    "email": "customer@example.com",
    "name": "John Doe"
  },
  "description": "Premium Subscription"
}
```

---

### Rate Limiting

```
Per API Key:
- 100 requests per second
- 10,000 requests per hour
- Burst allowance: 200 requests

Headers:
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1698758400
```

---

### Idempotency

- Use `Idempotency-Key` header for POST requests
- Key format: UUID v4 or random 32+ character string
- Keys expire after 24 hours
- Same key = same response (cached)

---

#### 5. Audit Requirements

```
What to Log:
✓ All access to cardholder data
✓ All actions by users with root/admin privileges
✓ All invalid access attempts
✓ All authentication attempts (success/failure)
✓ All audit log initialization
✓ All changes to authentication mechanisms
✓ All changes to audit logs
✓ All changes to system-level objects

Log Retention:
✓ Minimum 3 months immediately available
✓ Minimum 1 year for archived logs
```

---

## Development Roadmap

### Phase 1: Foundation (Weeks 1-4)

**Goal**: Core infrastructure and basic transaction processing

**Week 1: Project Setup**

- [ ] Initialize Git repository
- [ ] Set up project structure for microservices
- [ ] Configure Docker Compose for local development
- [ ] Set up PostgreSQL with initial schema
- [ ] Set up Kafka and basic topics
- [ ] Initialize HashiCorp Vault (dev mode)

**Week 2: Authentication & Merchant Services**

- [ ] Implement Auth Service (JWT, API keys)
- [ ] Implement Merchant Service (CRUD operations)
- [ ] Set up API Gateway (Kong/Traefik)
- [ ] Implement basic RBAC
- [ ] Add rate limiting

**Week 3: Core Payment Processing**

- [ ] Implement Tokenization Service
- [ ] Implement encryption/decryption logic
- [ ] Integrate with Vault for key management
- [ ] Implement Transaction Service (authorize, capture)
- [ ] Add Kafka event publishing

**Week 4: Payment API & Testing**

- [ ] Implement Payment API Service
- [ ] Add input validation
- [ ] Implement idempotency
- [ ] Write unit tests
- [ ] Write integration tests
- [ ] Test with fake card data

---

### Phase 2: Extended Features (Weeks 5-8)

**Week 5: Fraud Detection**

- [ ] Implement Fraud Detection Service
- [ ] Create rule engine
- [ ] Add velocity checks
- [ ] Implement risk scoring
- [ ] Add blocked lists functionality

**Week 6: 3D Secure & Advanced Transactions**

- [ ] Implement 3DS Service (simulated)
- [ ] Add void/reversal functionality
- [ ] Add refund functionality
- [ ] Implement partial captures/refunds
- [ ] Test complete transaction lifecycle

**Week 7: Settlement & Batching**

- [ ] Implement Settlement Service
- [ ] Add manual batching
- [ ] Add automatic batching (cron)
- [ ] Generate settlement reports
- [ ] Test batch processing

**Week 8: Recurring Payments**

- [ ] Implement Recurring Service
- [ ] Add subscription management
- [ ] Add installment plans
- [ ] Implement retry logic for failed payments
- [ ] Add dunning management

---

### Phase 3: Customer-Facing Features (Weeks 9-12)

**Week 9: Invoice System**

- [ ] Implement Invoice Service
- [ ] Add invoice creation and management
- [ ] Generate payment links
- [ ] Track invoice status
- [ ] Implement manual payment recording

**Week 10: Notifications**

- [ ] Implement Notification Service
- [ ] Set up email templates
- [ ] Set up SMS functionality (Twilio simulator)
- [ ] Add notification scheduling
- [ ] Test delivery mechanisms

**Week 11: Webhooks**

- [ ] Implement Webhook Service
- [ ] Add webhook HMAC signing
- [ ] Implement retry logic
- [ ] Add delivery tracking
- [ ] Test webhook delivery

**Week 12: Virtual Terminal (Frontend) - Part 1**

- [ ] Set up Next.js project
- [ ] Implement authentication UI
- [ ] Create dashboard
- [ ] Build transaction list page
- [ ] Add transaction detail page

---

### Phase 4: Frontend & Polish (Weeks 13-16)

**Week 13: Virtual Terminal - Part 2**

- [ ] Build manual transaction entry form
- [ ] Create invoice management UI
- [ ] Build recurring payment UI
- [ ] Add customer management
- [ ] Implement search and filtering

**Week 14: Virtual Terminal - Part 3**

- [ ] Create reporting dashboards
- [ ] Build settings pages
- [ ] Add API key management UI
- [ ] Create webhook configuration UI
- [ ] Add batch management UI

**Week 15: Audit & Reporting**

- [ ] Implement Audit Service
- [ ] Implement Reporting Service
- [ ] Add report generation
- [ ] Create audit log viewer
- [ ] Add data export functionality

**Week 16: Monitoring & Observability**

- [ ] Add Prometheus metrics
- [ ] Set up Grafana dashboards
- [ ] Implement distributed tracing
- [ ] Add structured logging
- [ ] Set up log aggregation (ELK/Loki)

---

### Phase 5: Security & Testing (Weeks 17-20)

**Week 17: Security Hardening**

- [ ] Implement mTLS between services
- [ ] Add request signing
- [ ] Implement IP whitelisting
- [ ] Add security headers
- [ ] Perform penetration testing

**Week 18: Comprehensive Testing**

- [ ] Write end-to-end tests
- [ ] Perform load testing
- [ ] Test failure scenarios
- [ ] Test data consistency
- [ ] Test Kafka message processing

**Week 19: Documentation**

- [ ] Write API documentation (OpenAPI/Swagger)
- [ ] Create developer guides
- [ ] Document architecture
- [ ] Create deployment guides
- [ ] Write security guidelines

**Week 20: Kubernetes & Production Prep**

- [ ] Create Kubernetes manifests
- [ ] Set up Helm charts
- [ ] Configure CI/CD pipelines
- [ ] Set up monitoring alerts
- [ ] Perform final testing

---
