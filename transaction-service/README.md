# 🔄 Transaction Service

**Core Payment Transaction Engine**

The Transaction Service is the heart of payment processing, managing the complete transaction lifecycle from authorization to settlement.

---

## ✨ Features

### Core Transaction Operations
- ✅ **Authorization** - Hold funds on customer's card (7-day expiry)
- ✅ **Capture** - Charge previously authorized funds (full or partial)
- ✅ **Void** - Cancel authorization before capture
- ✅ **Refund** - Return funds to customer (full or partial)

### Financial Management
- ✅ **Multi-Currency Support** - USD, EUR, MAD with automatic conversion
- ✅ **Exchange Rate Management** - Hourly rate updates (currently using default rates)
- ✅ **Processing Fees** - Automatic calculation (2.9% + $0.30 converted to MAD)
- ✅ **Settlement Processing** - Daily batch creation at midnight (T+2 settlement)

### Security & Compliance
- ✅ **Card Simulator** - Test card processing for development
- ✅ **Chargeback Management** - Complete dispute handling workflow
- ✅ **Audit Logging** - All transaction state changes tracked
- ✅ **Transaction Events** - Complete history of all operations

### Background Workers
- ✅ **Settlement Worker** - Runs daily at midnight
- ✅ **Auto-Void Worker** - Expires old authorizations (runs hourly)
- ✅ **Currency Update Worker** - Updates exchange rates (runs hourly)

---

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────────┐
│          TRANSACTION SERVICE (gRPC)                  │
│                Port 50053                            │
├──────────────────────────────────────────────────────┤
│                                                      │
│  ┌────────────────────────────────────────────┐    │
│  │       gRPC Server (Internal Only)          │    │
│  │  - Authorize                               │    │
│  │  - Capture                                 │    │
│  │  - Void                                    │    │
│  │  - Refund                                  │    │
│  │  - GetTransaction                          │    │
│  │  - ListTransactions                        │    │
│  └──────────────┬─────────────────────────────┘    │
│                 │                                   │
│  ┌──────────────▼─────────────────────────────┐    │
│  │       Transaction Service Layer            │    │
│  │  - State machine management                │    │
│  │  - Business logic                          │    │
│  │  - Currency conversion                     │    │
│  │  - Fee calculation                         │    │
│  └──────────────┬─────────────────────────────┘    │
│                 │                                   │
│        ┌────────┼────────┐                         │
│        │        │        │                         │
│   ┌────▼───┐ ┌─▼──┐ ┌──▼────┐                     │
│   │ Token  │ │Card│ │Settle │                     │
│   │Service │ │Sim │ │Worker │                     │
│   │(gRPC)  │ │    │ │       │                     │
│   └────────┘ └────┘ └───────┘                     │
│                                                     │
│  ┌──────────────────────────────────────────┐     │
│  │    PostgreSQL + Redis Storage            │     │
│  │  - Transactions                          │     │
│  │  - Transaction Events                    │     │
│  │  - Settlement Batches                    │     │
│  │  - Exchange Rates                        │     │
│  │  - Chargebacks                           │     │
│  └──────────────────────────────────────────┘     │
└──────────────────────────────────────────────────────┘
```

---

## 🔄 Transaction State Machine

```
PENDING
   ├─→ AUTHORIZED (7 days expiry)
   │      ├─→ CAPTURED
   │      │     ├─→ SETTLED (T+2)
   │      │     └─→ REFUNDED / PARTIALLY_REFUNDED
   │      ├─→ VOIDED (manual or auto-void)
   │      └─→ EXPIRED (auto-void after 7 days)
   └─→ FAILED (declined by issuer/fraud)
```

---

## 💱 Multi-Currency Processing

### Supported Currencies
- **USD** - US Dollar
- **EUR** - Euro
- **MAD** - Moroccan Dirham (base currency)

### Currency Conversion
All amounts are converted to MAD for processing:
- **USD → MAD**: 1 USD = 10 MAD
- **EUR → MAD**: 1 EUR = 11 MAD
- **MAD → MAD**: No conversion

Exchange rates are updated daily (configurable).

### Example Flow
```
Merchant processes $99.99 USD
↓
Convert to MAD: 99.99 * 10 = 999.90 MAD
↓
Calculate fee: 999.90 * 0.029 + 300 = 329 MAD ($0.30 base fee)
↓
Net amount: 999.90 - 329 = 670.90 MAD (merchant receives)
```

---

## 💰 Processing Fees

### Fee Structure
- **Percentage**: 2.9%
- **Fixed Fee**: $0.30 (converted to MAD = 300 MAD cents)

### Calculation
```
Total Fee = (Amount * 0.029) + Base Fee
Net Amount = Amount - Total Fee
```

### Examples
```
$10.00 → Fee: $0.59 → Net: $9.41
$100.00 → Fee: $3.20 → Net: $96.80
$1,000.00 → Fee: $29.30 → Net: $970.70
```

---

## 📅 Settlement Process

### Daily Settlement (Runs at Midnight)
1. **Batch Creation**
   - Collects all captured transactions from previous day
   - Groups by merchant
   - Calculates gross amount, fees, refunds
   - Creates settlement batch

2. **T+2 Settlement**
   - Batches settle 2 business days after capture
   - Funds transferred to merchant's bank account
   - Settlement confirmation sent

3. **Settlement Report**
   - CSV file with transaction details
   - Breakdown by currency
   - Fee summary
   - Net payout amount

---

## 🛡️ Chargeback Management

### Chargeback Lifecycle
```
Customer disputes → Chargeback created (NEEDS_RESPONSE)
   ↓
Merchant has 7 days to respond
   ↓
Merchant submits evidence → UNDER_REVIEW
   ↓
Bank/Network decision → WON or LOST
```

### Chargeback Fee
- **Fee**: $15.00 per chargeback
- **Charged even if merchant wins**

---

## 🧪 Test Cards (Card Simulator)

| Card Number (Last 4) | Result | Response Code | Use Case |
|----------------------|--------|---------------|----------|
| 4242 | ✅ Approved | 00 | Visa success |
| 4444 | ✅ Approved | 00 | Mastercard success |
| 0002 | ❌ Declined | 05 | Generic decline |
| 9995 | ❌ Declined | 51 | Insufficient funds |
| 0069 | ❌ Declined | 54 | Expired card |
| 0127 | ❌ Declined | N7 | CVV mismatch |
| 0119 | ❌ Declined | 96 | Processing error |

---

## 📦 Installation

```bash
# 1. Create database
psql -U postgres -c "CREATE DATABASE transaction_db;"

# 2. Run migrations
cd transaction-service
go run cmd/migrate/migrate.go

# 3. Generate gRPC code (if proto modified)
protoc \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/transaction.proto

# 4. Start service
go run cmd/main.go
```

**Service runs on:**
- gRPC: `localhost:50053`

---

## 🔌 gRPC API

### Authorize
```protobuf
rpc Authorize(AuthorizeRequest) returns (AuthorizeResponse);
```

### Capture
```protobuf
rpc Capture(CaptureRequest) returns (CaptureResponse);
```

### Void
```protobuf
rpc Void(VoidRequest) returns (VoidResponse);
```

### Refund
```protobuf
rpc Refund(RefundRequest) returns (RefundResponse);
```

---

## 🔧 Background Workers

### 1. Settlement Worker
- **Frequency**: Daily at midnight
- **Tasks**:
  - Create settlement batches
  - Process T+2 settlements
  - Generate settlement reports

### 2. Auto-Void Worker
- **Frequency**: Every hour
- **Tasks**:
  - Find authorizations > 7 days old
  - Auto-void expired authorizations
  - Send notifications

### 3. Currency Update Worker
- **Frequency**: Every hour
- **Tasks**:
  - Fetch latest exchange rates
  - Update database
  - (Currently uses default rates)

---

## 📊 Database Schema

### Core Tables
- **transactions** - All payment transactions
- **transaction_events** - State change history
- **settlement_batches** - Daily settlement batches
- **exchange_rates** - Currency conversion rates
- **chargebacks** - Dispute records
- **issuer_responses** - Debug logs

---

## ⚙️ Configuration

```bash
# Server
GRPC_PORT=50053

# Database
DATABASE_DSN=postgresql://user:pass@localhost/transaction_db

# Redis
REDIS_DSN=redis://localhost:6379/5

# External Services
TOKENIZATION_SERVICE_GRPC=localhost:50052

# Logging
LOG_LEVEL=info
```

---

## 🐛 Troubleshooting

### Issue: "Failed to connect to tokenization service"
**Solution:** Ensure tokenization service is running on port 50051

### Issue: "Currency conversion failed"
**Solution:** Check exchange_rates table has data. Run migration to seed default rates.

### Issue: "Settlement batch not created"
**Solution:** Check settlement worker logs. Verify transactions exist with status "captured"

---
## Support

For issues and questions:

- GitHub : https://github.com/rhaloubi/Payment-Gateway-Microservices
- Email: redahaloubi8@gmail.com
---

## 📄 License

Copyright © 2025 Payment Gateway. All rights reserved.

---

**Service Version:** v1.0.0  
**gRPC Version:** v1  
**Last Updated:** December 2025