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

**Database Tables**:
- `merchants`
- `api_keys`
- `sessions`
- `permissions`
- `roles`

**Endpoints**:
```
POST   /v1/auth/register
POST   /v1/auth/login
POST   /v1/auth/refresh
POST   /v1/auth/logout
GET    /v1/auth/verify
POST   /v1/api-keys/generate
DELETE /v1/api-keys/{key_id}
```

---

### 2. **Merchant Service** (Multi-Tenancy Management)
**Purpose**: Merchant onboarding, configuration, and management

**Responsibilities**:
- Merchant registration and KYC (Know Your Customer) - simplified
- Merchant profile management
- Business settings and configurations
- Fee structure management
- Currency and payment method settings

**Database Tables**:
- `merchants`
- `merchant_settings`
- `merchant_fees`
- `merchant_kyc` (basic info for learning)

**Endpoints**:
```
POST   /v1/merchants
GET    /v1/merchants/{merchant_id}
PATCH  /v1/merchants/{merchant_id}
GET    /v1/merchants/{merchant_id}/settings
PUT    /v1/merchants/{merchant_id}/settings
```

---

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

**Endpoints**:
```
POST   /v1/payments/authorize
POST   /v1/payments/sale
POST   /v1/payments/capture
POST   /v1/payments/void
POST   /v1/payments/refund
GET    /v1/payments/{transaction_id}
POST   /v1/payments/hosted-page
```

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

**Database Tables**:
- `card_tokens` (encrypted card data)
- `encryption_keys` (wrapped DEKs)

**Endpoints** (Internal Only):
```
POST   /internal/v1/tokenize
POST   /internal/v1/detokenize
DELETE /internal/v1/tokens/{token_id}
```

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

**Database Tables**:
- `transactions`
- `transaction_events`
- `transaction_amounts`
- `authorization_holds`

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

**Database Tables**:
- `fraud_rules`
- `fraud_scores`
- `blocked_cards`
- `blocked_ips`
- `transaction_velocity`

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

## Database Design

### PostgreSQL Schema Design

#### 1. Merchants Table
```sql
CREATE TABLE merchants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_code VARCHAR(50) UNIQUE NOT NULL,
    business_name VARCHAR(255) NOT NULL,
    legal_name VARCHAR(255),
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(50),
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, suspended, closed
    country_code CHAR(2) NOT NULL,
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',
    timezone VARCHAR(50) DEFAULT 'UTC',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_merchants_status ON merchants(status);
CREATE INDEX idx_merchants_code ON merchants(merchant_code);
```

#### 2. API Keys Table
```sql
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id),
    key_hash VARCHAR(255) NOT NULL UNIQUE, -- SHA-256 hash
    key_prefix VARCHAR(20) NOT NULL, -- pk_live_ or sk_test_
    name VARCHAR(100),
    scopes JSONB, -- ['payments:write', 'payments:read']
    is_active BOOLEAN DEFAULT true,
    last_used_at TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP
);

CREATE INDEX idx_api_keys_merchant ON api_keys(merchant_id);
CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);
```

#### 3. Card Tokens Table (Encrypted Vault)
```sql
CREATE TABLE card_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id),
    token VARCHAR(100) UNIQUE NOT NULL,
    -- Encrypted fields using pgcrypto or application-level encryption
    card_number_encrypted BYTEA NOT NULL,
    cardholder_name_encrypted BYTEA NOT NULL,
    expiry_month_encrypted BYTEA NOT NULL,
    expiry_year_encrypted BYTEA NOT NULL,
    -- Searchable metadata (not sensitive)
    card_brand VARCHAR(20), -- visa, mastercard, amex
    last_four CHAR(4) NOT NULL,
    bin CHAR(6), -- Bank Identification Number
    fingerprint VARCHAR(64), -- Hash for duplicate detection
    encryption_key_id UUID NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_tokens_merchant ON card_tokens(merchant_id);
CREATE INDEX idx_tokens_fingerprint ON card_tokens(fingerprint);
```

#### 4. Transactions Table
```sql
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id),
    transaction_code VARCHAR(50) UNIQUE NOT NULL, -- txn_xxx
    idempotency_key VARCHAR(100),
    
    -- Transaction details
    type VARCHAR(20) NOT NULL, -- auth, sale, capture, void, refund
    status VARCHAR(20) NOT NULL, -- pending, authorized, captured, settled, voided, refunded, failed
    
    -- Amounts
    amount DECIMAL(19, 4) NOT NULL,
    currency CHAR(3) NOT NULL,
    authorized_amount DECIMAL(19, 4),
    captured_amount DECIMAL(19, 4),
    refunded_amount DECIMAL(19, 4) DEFAULT 0,
    
    -- Payment method
    payment_method VARCHAR(20) DEFAULT 'card',
    card_token_id UUID REFERENCES card_tokens(id),
    card_last_four CHAR(4),
    card_brand VARCHAR(20),
    
    -- References
    parent_transaction_id UUID REFERENCES transactions(id), -- For captures/refunds
    invoice_id UUID,
    customer_id UUID,
    
    -- 3DS
    three_ds_authenticated BOOLEAN DEFAULT false,
    three_ds_auth_id UUID,
    
    -- Fraud
    fraud_score INTEGER, -- 0-100
    fraud_status VARCHAR(20), -- approved, review, declined
    
    -- Metadata
    description TEXT,
    customer_email VARCHAR(255),
    customer_ip VARCHAR(45),
    user_agent TEXT,
    metadata JSONB,
    
    -- Processing details
    processor_response JSONB,
    failure_code VARCHAR(50),
    failure_message TEXT,
    
    -- Timestamps
    authorized_at TIMESTAMP,
    captured_at TIMESTAMP,
    settled_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_txn_merchant ON transactions(merchant_id);
CREATE INDEX idx_txn_status ON transactions(status);
CREATE INDEX idx_txn_created ON transactions(created_at DESC);
CREATE INDEX idx_txn_customer_email ON transactions(customer_email);
CREATE INDEX idx_txn_idempotency ON transactions(idempotency_key);
CREATE INDEX idx_txn_parent ON transactions(parent_transaction_id);
```

#### 5. Transaction Events Table (Event Sourcing)
```sql
CREATE TABLE transaction_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    event_type VARCHAR(50) NOT NULL, -- created, authorized, captured, etc.
    previous_status VARCHAR(20),
    new_status VARCHAR(20) NOT NULL,
    amount DECIMAL(19, 4),
    metadata JSONB,
    occurred_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_txn_events_transaction ON transaction_events(transaction_id);
CREATE INDEX idx_txn_events_type ON transaction_events(event_type);
```

#### 6. Batches Table
```sql
CREATE TABLE batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id),
    batch_code VARCHAR(50) UNIQUE NOT NULL,
    type VARCHAR(20) NOT NULL, -- manual, automatic
    status VARCHAR(20) NOT NULL DEFAULT 'open', -- open, closed, settled
    
    -- Totals
    transaction_count INTEGER DEFAULT 0,
    total_amount DECIMAL(19, 4) DEFAULT 0,
    currency CHAR(3) NOT NULL,
    
    -- Timestamps
    opened_at TIMESTAMP NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMP,
    settled_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_batches_merchant ON batches(merchant_id);
CREATE INDEX idx_batches_status ON batches(status);
```

#### 7. Batch Transactions (Join Table)
```sql
CREATE TABLE batch_transactions (
    batch_id UUID NOT NULL REFERENCES batches(id),
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    added_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (batch_id, transaction_id)
);

CREATE INDEX idx_batch_txn_batch ON batch_transactions(batch_id);
CREATE INDEX idx_batch_txn_transaction ON batch_transactions(transaction_id);
```

#### 8. Recurring Payment Plans
```sql
CREATE TABLE payment_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id),
    customer_id UUID,
    plan_code VARCHAR(50) UNIQUE NOT NULL,
    
    -- Plan type
    type VARCHAR(20) NOT NULL, -- subscription, installment
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, paused, cancelled, completed
    
    -- Payment details
    amount DECIMAL(19, 4) NOT NULL,
    currency CHAR(3) NOT NULL,
    card_token_id UUID NOT NULL REFERENCES card_tokens(id),
    
    -- Schedule
    frequency VARCHAR(20) NOT NULL, -- daily, weekly, monthly, yearly
    interval INTEGER DEFAULT 1, -- every N frequency periods
    total_cycles INTEGER, -- NULL for subscriptions, number for installments
    cycles_completed INTEGER DEFAULT 0,
    
    -- Dates
    start_date DATE NOT NULL,
    next_charge_date DATE NOT NULL,
    end_date DATE, -- For installments
    
    -- Metadata
    description TEXT,
    metadata JSONB,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    paused_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE INDEX idx_plans_merchant ON payment_plans(merchant_id);
CREATE INDEX idx_plans_status ON payment_plans(status);
CREATE INDEX idx_plans_next_charge ON payment_plans(next_charge_date);
```

#### 9. Invoices Table
```sql
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id),
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    
    -- Customer details
    customer_id UUID,
    customer_email VARCHAR(255) NOT NULL,
    customer_name VARCHAR(255),
    billing_address JSONB,
    
    -- Invoice details
    status VARCHAR(20) NOT NULL DEFAULT 'draft', -- draft, sent, paid, overdue, cancelled
    due_date DATE,
    
    -- Amounts
    subtotal DECIMAL(19, 4) NOT NULL,
    tax_amount DECIMAL(19, 4) DEFAULT 0,
    discount_amount DECIMAL(19, 4) DEFAULT 0,
    total_amount DECIMAL(19, 4) NOT NULL,
    paid_amount DECIMAL(19, 4) DEFAULT 0,
    currency CHAR(3) NOT NULL,
    
    -- Payment
    payment_link VARCHAR(255) UNIQUE,
    transaction_id UUID REFERENCES transactions(id),
    
    -- Metadata
    notes TEXT,
    terms TEXT,
    metadata JSONB,
    
    -- Timestamps
    issued_at TIMESTAMP,
    paid_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invoices_merchant ON invoices(merchant_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_customer_email ON invoices(customer_email);
```

#### 10. Invoice Items Table
```sql
CREATE TABLE invoice_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    description VARCHAR(500) NOT NULL,
    quantity DECIMAL(10, 2) NOT NULL DEFAULT 1,
    unit_price DECIMAL(19, 4) NOT NULL,
    amount DECIMAL(19, 4) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invoice_items_invoice ON invoice_items(invoice_id);
```

#### 11. 3DS Authentications Table
```sql
CREATE TABLE three_ds_authentications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id),
    transaction_id UUID REFERENCES transactions(id),
    
    -- 3DS details
    version VARCHAR(10) NOT NULL, -- 2.1.0, 2.2.0
    status VARCHAR(20) NOT NULL, -- initiated, challenged, authenticated, failed
    authentication_type VARCHAR(20), -- frictionless, challenge
    
    -- Results
    eci VARCHAR(2), -- Electronic Commerce Indicator
    cavv VARCHAR(100), -- Cardholder Authentication Verification Value
    xid VARCHAR(100), -- Transaction ID
    
    -- Challenge
    challenge_url TEXT,
    challenge_completed_at TIMESTAMP,
    
    -- Liability
    liability_shift BOOLEAN DEFAULT false,
    
    -- Metadata
    authentication_value JSONB,
    
    -- Timestamps
    initiated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    authenticated_at TIMESTAMP
);

CREATE INDEX idx_3ds_merchant ON three_ds_authentications(merchant_id);
CREATE INDEX idx_3ds_transaction ON three_ds_authentications(transaction_id);
```

#### 12. Fraud Rules Table
```sql
CREATE TABLE fraud_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID REFERENCES merchants(id), -- NULL for global rules
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL, -- velocity, amount_threshold, geolocation, etc.
    condition JSONB NOT NULL, -- Rule configuration
    score_impact INTEGER NOT NULL, -- Points to add to fraud score
    is_active BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fraud_rules_merchant ON fraud_rules(merchant_id);
CREATE INDEX idx_fraud_rules_active ON fraud_rules(is_active);
```

#### 13. Fraud Scores Table
```sql
CREATE TABLE fraud_scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    merchant_id UUID NOT NULL REFERENCES merchants(id),
    
    -- Scoring
    total_score INTEGER NOT NULL, -- 0-100
    risk_level VARCHAR(20) NOT NULL, -- low, medium, high
    decision VARCHAR(20) NOT NULL, -- approve, review, decline
    
    -- Rules triggered
    rules_triggered JSONB, -- Array of rule IDs and their impacts
    
    -- Signals
    signals JSONB, -- Fraud indicators found
    
    -- Timestamps
    calculated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fraud_scores_transaction ON fraud_scores(transaction_id);
CREATE INDEX idx_fraud_scores_merchant ON fraud_scores(merchant_id);
```

#### 14. Webhook Endpoints Table
```sql
CREATE TABLE webhook_endpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id),
    url VARCHAR(500) NOT NULL,
    secret VARCHAR(100) NOT NULL, -- For HMAC signing
    
    -- Event subscriptions
    events TEXT[] NOT NULL, -- ['transaction.authorized', 'invoice.paid']
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    failure_count INTEGER DEFAULT 0,
    last_failure_at TIMESTAMP,
    
    -- Metadata
    description TEXT,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhook_endpoints_merchant ON webhook_endpoints(merchant_id);
CREATE INDEX idx_webhook_endpoints_active ON webhook_endpoints(is_active);
```

#### 15. Webhook Deliveries Table
```sql
CREATE TABLE webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_endpoint_id UUID NOT NULL REFERENCES webhook_endpoints(id),
    merchant_id UUID NOT NULL REFERENCES merchants(id),
    
    -- Event details
    event_type VARCHAR(100) NOT NULL,
    event_id UUID NOT NULL,
    payload JSONB NOT NULL,
    
    -- Delivery
    status VARCHAR(20) NOT NULL, -- pending, delivered, failed
    attempt_count INTEGER DEFAULT 0,
    response_status INTEGER,
    response_body TEXT,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    delivered_at TIMESTAMP,
    next_retry_at TIMESTAMP
);

CREATE INDEX idx_webhook_deliveries_endpoint ON webhook_deliveries(webhook_endpoint_id);
CREATE INDEX idx_webhook_deliveries_status ON webhook_deliveries(status);
CREATE INDEX idx_webhook_deliveries_retry ON webhook_deliveries(next_retry_at) WHERE status = 'pending';
```

#### 16. Audit Logs Table
```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID REFERENCES merchants(id),
    user_id UUID,
    
    -- Event details
    action VARCHAR(100) NOT NULL, -- 'transaction.created', 'api_key.revoked'
    resource_type VARCHAR(50) NOT NULL, -- 'transaction', 'api_key', 'merchant'
    resource_id UUID,
    
    -- Changes
    changes JSONB, -- Before/after values
    
    -- Request context
    ip_address VARCHAR(45),
    user_agent TEXT,
    request_id VARCHAR(100),
    
    -- Timestamps
    occurred_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_merchant ON audit_logs(merchant_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_occurred ON audit_logs(occurred_at DESC);
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

## Infrastructure & DevOps

### Kubernetes Architecture

#### Namespace Structure
```yaml
namespaces:
  - payment-gateway-prod
  - payment-gateway-staging
  - payment-gateway-dev
```

#### Service Deployment Example
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: transaction-service
  namespace: payment-gateway-prod
spec:
  replicas: 3
  selector:
    matchLabels:
      app: transaction-service
  template:
    metadata:
      labels:
        app: transaction-service
        version: v1
    spec:
      containers:
      - name: transaction-service
        image: your-registry/transaction-service:v1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: postgres-secrets
              key: url
        - name: KAFKA_BROKERS
          value: "kafka-0.kafka-headless:9092,kafka-1.kafka-headless:9092"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

---

### Docker Compose (Local Development)

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: payment_gateway
      POSTGRES_USER: pguser
      POSTGRES_PASSWORD: pgpass
    ports:
      - "5432:5432"
    volumes:
      - postgres-data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  kafka:
    image: confluentinc/cp-kafka:7.5.0
    depends_on:
      - zookeeper
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
    ports:
      - "9092:9092"

  zookeeper:
    image: confluentinc/cp-zookeeper:7.5.0
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000
    ports:
      - "2181:2181"

  vault:
    image: vault:1.15
    cap_add:
      - IPC_LOCK
    environment:
      VAULT_DEV_ROOT_TOKEN_ID: root-token
      VAULT_DEV_LISTEN_ADDRESS: 0.0.0.0:8200
    ports:
      - "8200:8200"

  auth-service:
    build: ./services/auth-service
    ports:
      - "8001:8080"
    environment:
      DATABASE_URL: postgres://pguser:pgpass@postgres:5432/payment_gateway
      REDIS_URL: redis://redis:6379
      JWT_PRIVATE_KEY_PATH: /secrets/jwt-private.key
    depends_on:
      - postgres
      - redis

  merchant-service:
    build: ./services/merchant-service
    ports:
      - "8002:8080"
    environment:
      DATABASE_URL: postgres://pguser:pgpass@postgres:5432/payment_gateway
      KAFKA_BROKERS: kafka:9092
    depends_on:
      - postgres
      - kafka

  payment-api:
    build: ./services/payment-api
    ports:
      - "8003:8080"
    environment:
      DATABASE_URL: postgres://pguser:pgpass@postgres:5432/payment_gateway
      KAFKA_BROKERS: kafka:9092
      REDIS_URL: redis://redis:6379
    depends_on:
      - postgres
      - kafka
      - redis

  tokenization-service:
    build: ./services/tokenization-service
    ports:
      - "8004:8080"
    environment:
      DATABASE_URL: postgres://pguser:pgpass@postgres:5432/payment_gateway
      VAULT_ADDR: http://vault:8200
      VAULT_TOKEN: root-token
    depends_on:
      - postgres
      - vault

  transaction-service:
    build: ./services/transaction-service
    ports:
      - "8005:8080"
    environment:
      DATABASE_URL: postgres://pguser:pgpass@postgres:5432/payment_gateway
      KAFKA_BROKERS: kafka:9092
    depends_on:
      - postgres
      - kafka

  fraud-detection-service:
    build: ./services/fraud-detection-service
    ports:
      - "8006:8080"
    environment:
      DATABASE_URL: postgres://pguser:pgpass@postgres:5432/payment_gateway
      KAFKA_BROKERS: kafka:9092
      REDIS_URL: redis://redis:6379
    depends_on:
      - postgres
      - kafka
      - redis

  three-ds-service:
    build: ./services/three-ds-service
    ports:
      - "8007:8080"
    environment:
      DATABASE_URL: postgres://pguser:pgpass@postgres:5432/payment_gateway
    depends_on:
      - postgres

  settlement-service:
    build: ./services/settlement-service
    ports:
      - "8008:8080"
    environment:
      DATABASE_URL: postgres://pguser:pgpass@postgres:5432/payment_gateway
      KAFKA_BROKERS: kafka:9092
    depends_on:
      - postgres
      - kafka

  recurring-service:
    build: ./services/recurring-service
    ports:
      - "8009:8080"
    environment:
      DATABASE_URL: postgres://pguser:pgpass@postgres:5432/payment_gateway
      KAFKA_BROKERS: kafka:9092
    depends_on:
      - postgres
      - kafka

  invoice-service:
    build: ./services/invoice-service
    ports:
      - "8010:8080"
    environment:
      DATABASE_URL: postgres://pguser:pgpass@postgres:5432/payment_gateway
      KAFKA_BROKERS: kafka:9092
    depends_on:
      - postgres
      - kafka

  notification-service:
    build: ./services/notification-service
    ports:
      - "8011:8080"
    environment:
      KAFKA_BROKERS: kafka:9092
      SMTP_HOST: mailhog
      SMTP_PORT: 1025
    depends_on:
      - kafka

  webhook-service:
    build: ./services/webhook-service
    ports:
      - "8012:8080"
    environment:
      DATABASE_URL: postgres://pguser:pgpass@postgres:5432/payment_gateway
      KAFKA_BROKERS: kafka:9092
    depends_on:
      - postgres
      - kafka

  audit-service:
    build: ./services/audit-service
    ports:
      - "8013:8080"
    environment:
      DATABASE_URL: postgres://pguser:pgpass@postgres:5432/payment_gateway
      KAFKA_BROKERS: kafka:9092
    depends_on:
      - postgres
      - kafka

  reporting-service:
    build: ./services/reporting-service
    ports:
      - "8014:8080"
    environment:
      DATABASE_URL: postgres://pguser:pgpass@postgres:5432/payment_gateway
    depends_on:
      - postgres

  virtual-terminal:
    build: ./frontend/virtual-terminal
    ports:
      - "3000:3000"
    environment:
      NEXT_PUBLIC_API_URL: http://localhost:8080
    depends_on:
      - payment-api

  api-gateway:
    image: kong:3.4
    environment:
      KONG_DATABASE: "off"
      KONG_DECLARATIVE_CONFIG: /kong/declarative/kong.yml
      KONG_PROXY_ACCESS_LOG: /dev/stdout
      KONG_ADMIN_ACCESS_LOG: /dev/stdout
      KONG_PROXY_ERROR_LOG: /dev/stderr
      KONG_ADMIN_ERROR_LOG: /dev/stderr
    ports:
      - "8080:8000"
      - "8443:8443"
      - "8001:8001"
    volumes:
      - ./kong.yml:/kong/declarative/kong.yml

  mailhog:
    image: mailhog/mailhog
    ports:
      - "1025:1025"
      - "8025:8025"

volumes:
  postgres-data:

networks:
  default:
    name: payment-gateway-network
```

---

### CI/CD Pipeline

#### GitHub Actions Workflow
```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: testpass
          POSTGRES_DB: test_db
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests
        run: |
          go test ./... -v -cover -coverprofile=coverage.out
      
      - name: Security scan
        run: |
          go install github.com/securego/gosec/v2/cmd/gosec@latest
          gosec ./...
      
      - name: Dependency check
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...

  build:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Build Docker images
        run: |
          docker build -t payment-gateway/auth-service:${{ github.sha }} ./services/auth-service
          docker build -t payment-gateway/transaction-service:${{ github.sha }} ./services/transaction-service
          # ... other services
      
      - name: Push to registry
        run: |
          echo "${{ secrets.DOCKER_PASSWORD }}" | docker login -u "${{ secrets.DOCKER_USERNAME }}" --password-stdin
          docker push payment-gateway/auth-service:${{ github.sha }}
          # ... other services

  deploy:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    
    steps:
      - name: Deploy to Kubernetes
        run: |
          kubectl set image deployment/auth-service auth-service=payment-gateway/auth-service:${{ github.sha }}
          # ... other deployments
```

---

### Monitoring & Observability

#### Prometheus Metrics
```go
// Example metrics in Go services
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    TransactionsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "payment_transactions_total",
            Help: "Total number of payment transactions",
        },
        []string{"status", "type", "merchant_id"},
    )

    TransactionDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "payment_transaction_duration_seconds",
            Help:    "Transaction processing duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"type"},
    )

    FraudScore = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "fraud_detection_score",
            Help:    "Fraud detection scores",
            Buckets: []float64{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
        },
        []string{"merchant_id"},
    )

    ActivePaymentPlans = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "recurring_payment_plans_active",
            Help: "Number of active recurring payment plans",
        },
        []string{"merchant_id", "type"},
    )
)
```

#### Logging with Structured Logs
```go
package logging

import (
    "go.uber.org/zap"
)

func InitLogger() (*zap.Logger, error) {
    config := zap.NewProductionConfig()
    config.EncoderConfig.TimeKey = "timestamp"
    config.EncoderConfig.MessageKey = "message"
    
    logger, err := config.Build()
    if err != nil {
        return nil, err
    }
    
    return logger, nil
}

// Usage
logger.Info("Transaction processed",
    zap.String("transaction_id", txnID),
    zap.String("merchant_id", merchantID),
    zap.Float64("amount", amount),
    zap.String("status", "authorized"),
    zap.String("trace_id", traceID),
)
```

#### Distributed Tracing with OpenTelemetry
```go
package tracing

import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/trace"
)

func InitTracer(serviceName string) error {
    exporter, err := jaeger.New(jaeger.WithCollectorEndpoint())
    if err != nil {
        return err
    }

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(/* ... */),
    )
    
    otel.SetTracerProvider(tp)
    return nil
}
```

---

### Security Scanning & Compliance

#### Container Security Scanning
```yaml
# Trivy scan in CI/CD
- name: Run Trivy vulnerability scanner
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: 'payment-gateway/transaction-service:${{ github.sha }}'
    format: 'sarif'
    output: 'trivy-results.sarif'
    severity: 'CRITICAL,HIGH'
```

#### SAST (Static Application Security Testing)
```bash
# SonarQube scan
sonar-scanner \
  -Dsonar.projectKey=payment-gateway \
  -Dsonar.sources=. \
  -Dsonar.host.url=http://sonarqube:9000 \
  -Dsonar.login=$SONAR_TOKEN
```

---

## Compliance Considerations

### PCI DSS Learning Implementation

#### 1. Network Segmentation
```
┌─────────────────────────────────────────────────┐
│              DMZ (Cardholder Data)              │
│  - Tokenization Service                         │
│  - Card Vault (Encrypted)                       │
│                                                 │
│  Firewall Rules: Deny all except specific      │
└─────────────────────────────────────────────────┘
                      ↕
┌─────────────────────────────────────────────────┐
│         Application Zone (Processing)           │
│  - Transaction Service                          │
│  - Payment API                                  │
│  - Fraud Detection                              │
└─────────────────────────────────────────────────┘
                      ↕
┌─────────────────────────────────────────────────┐
│            Public Zone (Frontend)               │
│  - Virtual Terminal                             │
│  - Hosted Payment Page                          │
└─────────────────────────────────────────────────┘
```

#### 2. Data Masking Rules
```go
// Mask PAN (Primary Account Number)
func MaskPAN(pan string) string {
    if len(pan) < 4 {
        return "****"
    }
    return "****" + pan[len(pan)-4:]
}

// Mask Email
func MaskEmail(email string) string {
    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return "***@***.com"
    }
    return parts[0][:2] + "***@" + parts[1]
}

// Never log these fields
var SensitiveFields = []string{
    "card_number",
    "cvv",
    "cvc",
    "pin",
    "password",
    "expiry",
}
```

#### 3. Access Control Matrix
```
Role: MERCHANT_ADMIN
Permissions:
  - transactions:read
  - transactions:create
  - transactions:refund
  - transactions:void
  - invoices:*
  - recurring:*
  - settings:*
  - api_keys:*

Role: MERCHANT_USER
Permissions:
  - transactions:read
  - transactions:create
  - invoices:read
  - invoices:create

Role: MERCHANT_VIEWER
Permissions:
  - transactions:read
  - invoices:read
  - reports:read

Role: DEVELOPER
Permissions:
  - api_keys:read
  - api_keys:create
  - webhooks:*
  - sandbox:*
```

#### 4. Encryption Key Management
```go
package encryption

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "io"
)

type EncryptionService struct {
    masterKey []byte // From Vault
}

// Encrypt data with AES-256-GCM
func (s *EncryptionService) Encrypt(plaintext []byte) (string, error) {
    block, err := aes.NewCipher(s.masterKey)
    if err != nil {
        return "", err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }

    ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt data
func (s *EncryptionService) Decrypt(encodedCiphertext string) ([]byte, error) {
    ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
    if err != nil {
        return nil, err
    }

    block, err := aes.NewCipher(s.masterKey)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonceSize := gcm.NonceSize()
    nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
    
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return nil, err
    }

    return plaintext, nil
}
```

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

## Go Project Structure

```
payment-gateway/
├── cmd/                                # Service entry points
│   ├── auth-service/
│   │   └── main.go
│   ├── merchant-service/
│   │   └── main.go
│   ├── payment-api/
│   │   └── main.go
│   └── ... (other services)
│
├── internal/                           # Private application code
│   ├── auth/
│   │   ├── handler/                    # HTTP handlers
│   │   ├── service/                    # Business logic
│   │   ├── repository/                 # Database access
│   │   └── model/                      # Domain models
│   ├── merchant/
│   ├── transaction/
│   ├── tokenization/
│   └── ... (other domains)
│
├── pkg/                                # Public shared libraries
│   ├── database/
│   │   ├── postgres.go
│   │   └── migrations/
│   ├── kafka/
│   │   ├── producer.go
│   │   ├── consumer.go
│   │   └── topics.go
│   ├── encryption/
│   │   ├── aes.go
│   │   └── vault.go
│   ├── auth/
│   │   ├── jwt.go
│   │   └── apikey.go
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── logging.go
│   │   ├── ratelimit.go
│   │   └── cors.go
│   ├── validator/
│   │   └── card.go
│   ├── logger/
│   │   └── zap.go
│   └── metrics/
│       └── prometheus.go
│
├── api/                                # API definitions
│   ├── openapi/
│   │   └── payment-gateway.yaml
│   └── proto/                          # gRPC (optional)
│       └── transaction.proto
│
├── frontend/                           # Next.js application
│   └── virtual-terminal/
│       ├── src/
│       │   ├── app/
│       │   ├── components/
│       │   ├── lib/
│       │   └── types/
│       ├── public/
│       ├── package.json
│       └── next.config.js
│
├── migrations/                         # Database migrations
│   ├── 001_create_merchants.up.sql
│   ├── 001_create_merchants.down.sql
│   ├── 002_create_transactions.up.sql
│   └── ...
│
├── deployments/                        # Deployment configurations
│   ├── docker/
│   │   ├── Dockerfile.auth
│   │   ├── Dockerfile.transaction
│   │   └── ...
│   ├── kubernetes/
│   │   ├── base/
│   │   └── overlays/
│   │       ├── dev/
│   │       ├── staging/
│   │       └── prod/
│   └── docker-compose.yml
│
├── scripts/                            # Utility scripts
│   ├── generate-keys.sh
│   ├── setup-vault.sh
│   ├── seed-data.sh
│   └── run-tests.sh
│
├── test/                               # Test files
│   ├── integration/
│   ├── e2e/
│   └── testdata/
│
├── docs/                               # Documentation
│   ├── architecture.md
│   ├── api.md
│   ├── security.md
│   └── deployment.md
│
├── .github/                            # GitHub Actions
│   └── workflows/
│       ├── ci.yml
│       └── cd.yml
│
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## Example Service Implementation

### Transaction Service (Go)

```go
// cmd/transaction-service/main.go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gorilla/mux"
    "payment-gateway/internal/transaction"
    "payment-gateway/pkg/database"
    "payment-gateway/pkg/kafka"
    "payment-gateway/pkg/logger"
    "payment-gateway/pkg/middleware"
)

func main() {
    // Initialize logger
    appLogger, err := logger.InitLogger()
    if err != nil {
        log.Fatal("Failed to initialize logger:", err)
    }
    defer appLogger.Sync()

    // Connect to database
    db, err := database.Connect(os.Getenv("DATABASE_URL"))
    if err != nil {
        appLogger.Fatal("Failed to connect to database", zap.Error(err))
    }
    defer db.Close()

    // Initialize Kafka producer
    kafkaProducer, err := kafka.NewProducer([]string{os.Getenv("KAFKA_BROKERS")})
    if err != nil {
        appLogger.Fatal("Failed to initialize Kafka producer", zap.Error(err))
    }
    defer kafkaProducer.Close()

    // Initialize repository
    repo := transaction.NewPostgresRepository(db)

    // Initialize service
    svc := transaction.NewService(repo, kafkaProducer, appLogger)

    // Initialize handler
    handler := transaction.NewHandler(svc, appLogger)

    // Set up router
    router := mux.NewRouter()
    
    // Middleware
    router.Use(middleware.RequestID)
    router.Use(middleware.Logging(appLogger))
    router.Use(middleware.Recovery(appLogger))

    // Health checks
    router.HandleFunc("/health", healthCheck).Methods("GET")
    router.HandleFunc("/ready", readinessCheck(db)).Methods("GET")

    // Internal API routes
    api := router.PathPrefix("/internal/v1").Subrouter()
    api.HandleFunc("/transactions", handler.CreateTransaction).Methods("POST")
    api.HandleFunc("/transactions/{id}", handler.GetTransaction).Methods("GET")
    api.HandleFunc("/transactions/{id}/capture", handler.CaptureTransaction).Methods("POST")
    api.HandleFunc("/transactions/{id}/void", handler.VoidTransaction).Methods("POST")
    api.HandleFunc("/transactions/{id}/refund", handler.RefundTransaction).Methods("POST")

    // Start server
    srv := &http.Server{
        Addr:         ":8080",
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Graceful shutdown
    go func() {
        appLogger.Info("Starting transaction service on :8080")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            appLogger.Fatal("Server failed to start", zap.Error(err))
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    appLogger.Info("Shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        appLogger.Fatal("Server forced to shutdown", zap.Error(err))
    }

    appLogger.Info("Server exited")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func readinessCheck(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := db.Ping(); err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            return
        }
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Ready"))
    }
}
```

```go
// internal/transaction/service.go
package transaction

import (
    "context"
    "encoding/json"
    "errors"
    "time"

    "github.com/google/uuid"
    "go.uber.org/zap"
    "payment-gateway/pkg/kafka"
)

type Service struct {
    repo     Repository
    producer *kafka.Producer
    logger   *zap.Logger
}

func NewService(repo Repository, producer *kafka.Producer, logger *zap.Logger) *Service {
    return &Service{
        repo:     repo,
        producer: producer,
        logger:   logger,
    }
}

func (s *Service) AuthorizeTransaction(ctx context.Context, req *AuthorizeRequest) (*Transaction, error) {
    // Validate request
    if err := req.Validate(); err != nil {
        return nil, err
    }

    // Create transaction
    txn := &Transaction{
        ID:            uuid.New(),
        MerchantID:    req.MerchantID,
        Type:          "authorize",
        Status:        "pending",
        Amount:        req.Amount,
        Currency:      req.Currency,
        CardTokenID:   req.CardTokenID,
        CustomerEmail: req.CustomerEmail,
        Description:   req.Description,
        Metadata:      req.Metadata,
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }

    // Save to database
    if err := s.repo.Create(ctx, txn); err != nil {
        s.logger.Error("Failed to create transaction", zap.Error(err))
        return nil, err
    }

    // Publish event to Kafka
    event := map[string]interface{}{
        "event_type":     "transaction.created",
        "transaction_id": txn.ID.String(),
        "merchant_id":    txn.MerchantID.String(),
        "amount":         txn.Amount,
        "currency":       txn.Currency,
        "timestamp":      time.Now().Format(time.RFC3339),
    }

    eventBytes, _ := json.Marshal(event)
    if err := s.producer.Publish("transaction.created", txn.ID.String(), eventBytes); err != nil {
        s.logger.Error("Failed to publish event", zap.Error(err))
        // Don't fail the transaction, but log the error
    }

    s.logger.Info("Transaction authorized",
        zap.String("transaction_id", txn.ID.String()),
        zap.String("merchant_id", txn.MerchantID.String()),
        zap.Float64("amount", txn.Amount),
    )

    return txn, nil
}

func (s *Service) CaptureTransaction(ctx context.Context, txnID uuid.UUID, amount float64) error {
    // Get transaction
    txn, err := s.repo.GetByID(ctx, txnID)
    if err != nil {
        return err
    }

    // Validate state
    if txn.Status != "authorized" {
        return errors.New("transaction must be authorized to capture")
    }

    if amount > txn.AuthorizedAmount {
        return errors.New("capture amount exceeds authorized amount")
    }

    // Update transaction
    txn.Status = "captured"
    txn.CapturedAmount = amount
    txn.CapturedAt = timePtr(time.Now())
    txn.UpdatedAt = time.Now()

    if err := s.repo.Update(ctx, txn); err != nil {
        return err
    }

    // Create event
    event := &TransactionEvent{
        ID:            uuid.New(),
        TransactionID: txnID,
        EventType:     "captured",
        PreviousStatus: "authorized",
        NewStatus:     "captured",
        Amount:        amount,
        OccurredAt:    time.Now(),
    }

    if err := s.repo.CreateEvent(ctx, event); err != nil {
        s.logger.Error("Failed to create event", zap.Error(err))
    }

    // Publish to Kafka
    kafkaEvent := map[string]interface{}{
        "event_type":     "transaction.captured",
        "transaction_id": txnID.String(),
        "amount":         amount,
        "timestamp":      time.Now().Format(time.RFC3339),
    }

    eventBytes, _ := json.Marshal(kafkaEvent)
    s.producer.Publish("transaction.captured", txnID.String(), eventBytes)

    return nil
}

// Helper function
func timePtr(t time.Time) *time.Time {
    return &t
}
```

---

## Testing Strategy

### Unit Tests
```go
// internal/transaction/service_test.go
package transaction_test

import (
    "context"
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "payment-gateway/internal/transaction"
)

type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, txn *transaction.Transaction) error {
    args := m.Called(ctx, txn)
    return args.Error(0)
}

func TestAuthorizeTransaction(t *testing.T) {
    // Arrange
    mockRepo := new(MockRepository)
    mockProducer := &kafka.MockProducer{}
    logger, _ := zap.NewDevelopment()
    
    svc := transaction.NewService(mockRepo, mockProducer, logger)

    req := &transaction.AuthorizeRequest{
        MerchantID:  uuid.New(),
        Amount:      99.99,
        Currency:    "USD",
        CardTokenID: uuid.New(),
    }

    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*transaction.Transaction")).
        Return(nil)

    // Act
    txn, err := svc.AuthorizeTransaction(context.Background(), req)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, txn)
    assert.Equal(t, "authorize", txn.Type)
    assert.Equal(t, "pending", txn.Status)
    assert.Equal(t, req.Amount, txn.Amount)
    mockRepo.AssertExpectations(t)
}
```

### Integration Tests
```go
// test/integration/transaction_test.go
package integration_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/suite"
    "payment-gateway/internal/transaction"
    "payment-gateway/pkg/database"
)

type TransactionTestSuite struct {
    suite.Suite
    db   *sql.DB
    repo *transaction.PostgresRepository
}

func (suite *TransactionTestSuite) SetupSuite() {
    // Setup test database
    db, err := database.Connect(os.Getenv("TEST_DATABASE_URL"))
    suite.Require().NoError(err)
    
    suite.db = db
    suite.repo = transaction.NewPostgresRepository(db)
}
