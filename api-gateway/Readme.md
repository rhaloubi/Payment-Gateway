# 🌐 API Gateway

**Unified Entry Point for Payment Processing System**

The API Gateway serves as the single point of entry for all client requests in the payment processing system. It handles cross-cutting concerns like routing, rate limiting, CORS, logging, and circuit breaking before forwarding requests to the appropriate microservices.

---

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Setup & Installation](#setup--installation)
- [Configuration](#configuration)
- [Middleware](#middleware)
- [Routing](#routing)
- [Circuit Breaker](#circuit-breaker)
- [Rate Limiting](#rate-limiting)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)

---

## 🎯 Overview

The API Gateway acts as a reverse proxy and aggregation layer, providing:

- **Single Entry Point** - Unified API endpoint for all clients
- **Service Abstraction** - Hides internal service topology
- **Cross-Cutting Concerns** - Centralized handling of common requirements
- **Traffic Management** - Rate limiting and circuit breaking
- **Observability** - Request logging and metrics collection

### Why Use an API Gateway?

✅ **Simplified Client Integration** - Clients interact with one endpoint  
✅ **Security** - Centralized authentication and authorization checks  
✅ **Performance** - Caching, rate limiting, and load balancing  
✅ **Monitoring** - Unified logging and metrics  
✅ **Flexibility** - Easy to add new services without client changes

---

## ✨ Features

### Traffic Management
- ✅ **Request Routing** - Route requests to appropriate backend services
- ✅ **Rate Limiting** - Per-client and per-endpoint limits
- ✅ **Circuit Breaking** - Prevent cascading failures
- ✅ **Timeouts** - Configurable per-service timeouts

### Security
- ✅ **CORS** - Cross-origin resource sharing configuration
- ✅ **Request Validation** - Header and payload validation
- ✅ **IP Tracking** - Client IP forwarding to backend services

### Observability
- ✅ **Request Logging** - JSON-structured logs
- ✅ **Metrics** - Prometheus metrics endpoint
- ✅ **Request Tracing** - Unique request ID generation
- ✅ **Health Checks** - Gateway and backend service health

### Resilience
- ✅ **Graceful Shutdown** - Proper connection draining
- ✅ **Panic Recovery** - Automatic error recovery
- ✅ **Retry Logic** - Circuit breaker with automatic recovery

---

## 🏗️ Architecture

### High-Level Flow

```
┌──────────────────────────────────────────────────────────────┐
│                        CLIENT REQUEST                         │
│              (Dashboard, Mobile App, CLI Tool)                │
└────────────────────────────┬─────────────────────────────────┘
                             │
                             │ HTTPS
                             │
┌────────────────────────────▼─────────────────────────────────┐
│                      API GATEWAY (Port 8080)                  │
├───────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │               MIDDLEWARE CHAIN                       │    │
│  │                                                       │    │
│  │  1. Logger       → Log request details              │    │
│  │  2. Recovery     → Catch panics                     │    │
│  │  3. CORS         → Handle cross-origin              │    │
│  │  4. Request ID   → Generate unique ID               │    │
│  │  5. Rate Limiter → Check request limits             │    │
│  │  6. Circuit Breaker → Check service health          │    │
│  │  7. Proxy        → Forward to backend               │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
└────────────────────────────┬─────────────────────────────────┘
                             │
            ┌────────────────┼────────────────┐
            │                │                │
    ┌───────▼──────┐  ┌─────▼──────┐  ┌─────▼──────┐
    │              │  │            │  │            │
    │ Auth Service │  │  Merchant  │  │  Payment   │
    │  (Port 8001) │  │  Service   │  │    API     │
    │              │  │ (Port 8002)│  │ (Port 8004)│
    └──────────────┘  └────────────┘  └────────────┘
```

### Request Flow

```
1. Client sends request → https://api.gateway.com/api/v1/payments/authorize
   ↓
2. CORS middleware → Validate origin and set headers
   ↓
3. Request ID middleware → Generate/extract request ID
   ↓
4. Rate limiter → Check if client is within limits
   ↓
5. Circuit breaker → Check if payment service is healthy
   ↓
6. Proxy handler → Forward to http://localhost:8004/api/v1/payments/authorize
   ↓
7. Backend service processes request
   ↓
8. Response flows back through middleware chain
   ↓
9. Client receives response with added headers (X-Request-ID, X-RateLimit-*, etc.)
```

---

## 📦 Setup & Installation

### Prerequisites

- Go 1.23+
- Backend services running:
  - Auth Service (port 8001)
  - Merchant Service (port 8002)
  - Payment API Service (port 8004)

### Installation Steps

```bash
# 1. Navigate to gateway directory
cd api-gateway

# 2. Install dependencies
go mod download

# 3. Copy configuration
cp configs/config.yaml configs/config.local.yaml
# Edit config.local.yaml with your service URLs

# 4. Build the gateway
go build -o bin/gateway cmd/main.go

# 5. Run
./bin/gateway

# Or use Air for hot reload
air
```

### Using Docker

```bash
# Build Docker image
docker build -t payment-gateway:latest .

# Run container
docker run -p 8080:8080 \
  -v $(pwd)/configs:/configs \
  -e CONFIG_PATH=/configs/config.yaml \
  payment-gateway:latest
```

---

## ⚙️ Configuration

### Configuration File: `configs/config.yaml`

```yaml
server:
  port: 8080
  read_timeout: 10s
  write_timeout: 10s
  idle_timeout: 60s

services:
  auth:
    url: "http://localhost:8001"
    timeout: 5s
    
  merchant:
    url: "http://localhost:8002"
    timeout: 10s
    
  payment:
    url: "http://localhost:8004"
    timeout: 30s

rate_limiting:
  enabled: true
  storage: "memory"  # or "redis"
  
  global:
    requests_per_hour: 1000
    
  endpoints:
    - pattern: "/api/v1/auth/login"
      requests_per_minute: 5
      by: "ip"
      
    - pattern: "/api/v1/auth/register"
      requests_per_hour: 3
      by: "ip"
      
    - pattern: "/api/v1/payments/*"
      requests_per_second: 20
      by: "api_key"

circuit_breaker:
  enabled: true
  
  auth_service:
    failure_threshold: 5      # Open after 5 failures
    timeout: 30s              # Stay open for 30s
    success_threshold: 2      # Close after 2 successes
    
  merchant_service:
    failure_threshold: 5
    timeout: 30s
    success_threshold: 2
    
  payment_service:
    failure_threshold: 3
    timeout: 15s
    success_threshold: 3

logging:
  level: "info"
  format: "json"
  output: "stdout"

metrics:
  enabled: true
  port: 9090
  path: "/metrics"
```

### Environment Variables

You can override configuration using environment variables:

```bash
# Server
export PORT=8080

# Services
export AUTH_SERVICE_URL=http://localhost:8001
export MERCHANT_SERVICE_URL=http://localhost:8002
export PAYMENT_SERVICE_URL=http://localhost:8004

# Config file path
export CONFIG_PATH=configs/config.yaml
```

---

## 🛡️ Middleware

### 1. CORS Middleware

Handles cross-origin requests for browser-based clients.

**Configuration:**
```go
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization, X-API-Key, Idempotency-Key, X-Client-Secret
```

**Preflight Requests:**
```bash
curl -X OPTIONS http://localhost:8080/api/v1/payments \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST"
```

---

### 2. Request ID Middleware

Generates or extracts unique request identifiers for tracing.

**Headers:**
- **Request**: `X-Request-ID` (optional - will be generated if missing)
- **Response**: `X-Request-ID` (always present)

**Example:**
```bash
curl http://localhost:8080/api/v1/auth/profile \
  -H "X-Request-ID: custom-trace-123"
```

---

### 3. Logger Middleware

Logs all requests in structured JSON format.

**Log Format:**
```json
{
  "time": "2025-12-31T10:00:00Z",
  "method": "POST",
  "path": "/api/v1/payments/authorize",
  "query": "",
  "ip": "192.168.1.100",
  "status": 200,
  "latency": "245ms"
}
```

---

### 4. Rate Limiter Middleware

Controls request rate per client or API key.

**Response Headers:**
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 995
X-RateLimit-Reset: 1735656000
```

**Rate Limit Exceeded Response:**
```json
{
  "success": false,
  "error": "rate limit exceeded",
  "retry_after": 3600
}
```

---

### 5. Recovery Middleware

Catches panics and returns graceful error responses.

**Panic Response:**
```json
{
  "success": false,
  "error": "internal server error"
}
```

---

## 🗺️ Routing

### Public Endpoints (No Authentication)

#### Health Check
```
GET /health
```

**Response:**
```json
{
  "status": "ok",
  "service": "api-gateway",
  "version": "1.0.0",
  "services": {
    "auth": "closed",
    "merchant": "closed",
    "payment": "closed"
  }
}
```

#### Metrics
```
GET /metrics
```
Returns Prometheus-formatted metrics.

---

### Auth Service Routes

Base path: `/api/v1/auth/*`  
Target: `http://localhost:8001`

```
POST   /api/v1/auth/register         → Register user
POST   /api/v1/auth/login            → Login
POST   /api/v1/auth/refresh          → Refresh token
GET    /api/v1/auth/profile          → Get user profile
POST   /api/v1/auth/logout           → Logout
POST   /api/v1/auth/change-password  → Change password
GET    /api/v1/auth/sessions         → List sessions
```

**Rate Limits:**
- Register: 3 requests/hour per IP
- Login: 5 requests/minute per IP

---

### Roles Service Routes

Base path: `/api/v1/roles/*`  
Target: `http://localhost:8001`

```
GET    /api/v1/roles                                        → List all roles
GET    /api/v1/roles/:id                                    → Get role details
POST   /api/v1/roles/assign                                 → Assign role
DELETE /api/v1/roles/assign                                 → Remove role
GET    /api/v1/roles/user/:user_id/merchant/:merchant_id   → Get user roles
GET    /api/v1/roles/user/:user_id/merchant/:merchant_id/permissions → Get permissions
```

---

### Merchant Service Routes

Base path: `/api/v1/merchants/*`  
Target: `http://localhost:8002`

```
POST   /api/v1/merchants                        → Create merchant
GET    /api/v1/merchants                        → List merchants
GET    /api/v1/merchants/:id                    → Get merchant
PUT    /api/v1/merchants/:id                    → Update merchant
PATCH  /api/v1/merchants/:id                    → Partial update
DELETE /api/v1/merchants/:id                    → Delete merchant

GET    /api/v1/merchants/:id/team               → List team members
POST   /api/v1/merchants/:id/team/invite        → Invite member
PATCH  /api/v1/merchants/:id/team/:user_id      → Update member role
DELETE /api/v1/merchants/:id/team/:user_id      → Remove member

GET    /api/v1/merchants/:id/settings           → Get settings
PATCH  /api/v1/merchants/:id/settings           → Update settings

POST   /api/v1/merchants/api-keys               → Create API key
GET    /api/v1/merchants/api-keys/merchant/:id  → List API keys
PATCH  /api/v1/merchants/api-keys/:id/deactivate → Deactivate key
DELETE /api/v1/merchants/api-keys/:id            → Delete key
```

---

### Payment Service Routes

Base path: `/api/v1/payments/*`  
Target: `http://localhost:8004`

```
POST   /api/v1/payments/authorize       → Authorize payment
POST   /api/v1/payments/sale            → Process sale
POST   /api/v1/payments/:id/capture     → Capture payment
POST   /api/v1/payments/:id/void        → Void payment
POST   /api/v1/payments/:id/refund      → Refund payment
GET    /api/v1/payments/:id             → Get payment details
GET    /api/v1/payments                 → List payments

GET    /api/v1/transactions             → List transactions
GET    /api/v1/transactions/:id         → Get transaction

POST   /api/v1/payment-intents          → Create payment intent
POST   /api/v1/payment-intents/:id/cancel → Cancel intent
```

**Rate Limit:** 20 requests/second per API key

---

### Public Payment Intent Routes

Base path: `/api/public/payment-intents/*`  
Target: `http://localhost:8004`

```
GET    /api/public/payment-intents/:id          → Get intent (client secret auth)
POST   /api/public/payment-intents/:id/confirm  → Confirm payment
```

---

## 🔄 Circuit Breaker

The circuit breaker prevents cascading failures by temporarily blocking requests to failing services.

### States

1. **Closed** (Normal)
   - All requests pass through
   - Failures are counted

2. **Open** (Service Down)
   - All requests are blocked immediately
   - Returns 503 Service Unavailable
   - Stays open for configured timeout

3. **Half-Open** (Testing Recovery)
   - Limited requests allowed through
   - If successful, transitions to Closed
   - If failed, transitions back to Open

### State Transitions

```
CLOSED → (5 failures) → OPEN
  ↑                        ↓
  └─ (2 successes) ← HALF-OPEN (after 30s timeout)
```

### Configuration Example

```yaml
circuit_breaker:
  payment_service:
    failure_threshold: 3    # Open after 3 consecutive failures
    timeout: 15s            # Stay open for 15 seconds
    success_threshold: 3    # Close after 3 consecutive successes
```

### Error Response (Circuit Open)

```json
{
  "success": false,
  "error": "service temporarily unavailable: payment"
}
```

---

## 🚦 Rate Limiting

### Global Rate Limit

Applied to all requests by default:

```yaml
global:
  requests_per_hour: 1000
```

### Endpoint-Specific Limits

```yaml
endpoints:
  - pattern: "/api/v1/auth/login"
    requests_per_minute: 5
    by: "ip"
  
  - pattern: "/api/v1/payments/*"
    requests_per_second: 20
    by: "api_key"
```

### Rate Limit Headers

**Response includes:**
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 847
X-RateLimit-Reset: 1735660800
X-RateLimit-Endpoint: payments
```

### Rate Limit Exceeded

**Status:** 429 Too Many Requests

**Response:**
```json
{
  "success": false,
  "error": "rate limit exceeded for payments",
  "retry_after": 45.2
}
```

### Identifier Strategy

- **By IP** - For public endpoints (login, register)
- **By API Key** - For authenticated endpoints (payments)

---

## 📊 Monitoring

### Health Check Endpoint

```bash
curl http://localhost:8080/health
```

**Response:**
```json
{
  "status": "ok",
  "service": "api-gateway",
  "version": "1.0.0",
  "services": {
    "auth": "closed",
    "merchant": "closed",
    "payment": "half-open"
  }
}
```

### Prometheus Metrics

```bash
curl http://localhost:8080/metrics
```

**Available Metrics:**
- Request count by endpoint
- Request duration histogram
- Circuit breaker state changes
- Rate limit hits
- Error rates by service

**Example Metrics:**
```
# HELP gateway_requests_total Total number of requests
# TYPE gateway_requests_total counter
gateway_requests_total{method="POST",path="/api/v1/payments",status="200"} 1523

# HELP gateway_request_duration_seconds Request duration in seconds
# TYPE gateway_request_duration_seconds histogram
gateway_request_duration_seconds_bucket{le="0.1"} 1200
gateway_request_duration_seconds_bucket{le="0.5"} 1500
```

### Logging

All requests are logged in JSON format:

```json
{
  "time": "2025-12-31T10:30:45Z",
  "method": "POST",
  "path": "/api/v1/payments/authorize",
  "query": "",
  "ip": "192.168.1.50",
  "status": 200,
  "latency": "342ms"
}
```

---

## 🔧 Advanced Configuration

### Custom Service Timeouts

Adjust timeouts based on service complexity:

```yaml
services:
  auth:
    timeout: 5s      # Fast service
  
  merchant:
    timeout: 10s     # Medium complexity
  
  payment:
    timeout: 30s     # May involve external calls
```

### Redis-Based Rate Limiting

For distributed deployments:

```yaml
rate_limiting:
  storage: "redis"
  redis_url: "redis://localhost:6379/0"
```

### Disable Features

```yaml
rate_limiting:
  enabled: false

circuit_breaker:
  enabled: false

metrics:
  enabled: false
```

---

## 🐛 Troubleshooting

### Issue: Gateway Won't Start

**Cause:** Port already in use

**Solution:**
```bash
# Check what's using port 8080
lsof -i :8080

# Change port in config
server:
  port: 8081
```

---

### Issue: "Service Temporarily Unavailable"

**Cause:** Circuit breaker is open

**Solution:**
```bash
# Check health endpoint
curl http://localhost:8080/health

# Response will show service state:
{
  "services": {
    "payment": "open"  # ← Circuit is open
  }
}

# Check backend service health
curl http://localhost:8004/health

# Wait for timeout (default 30s) or restart backend service
```

---

### Issue: Rate Limit Always Exceeded

**Cause:** Using same IP/API key across multiple clients

**Solution:**
```bash
# Check current limit
curl -I http://localhost:8080/api/v1/payments

# Response headers show:
X-RateLimit-Limit: 20
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1735660800

# Increase limit in config or wait for reset
```

---

### Issue: CORS Errors in Browser

**Cause:** Missing or incorrect CORS headers

**Solution:**

Check the browser console for specific error. Common fixes:

```yaml
# In middleware/cors.go, ensure headers include:
Access-Control-Allow-Origin: *  # Or specific origin
Access-Control-Allow-Headers: Content-Type, Authorization, X-API-Key
```

**Test preflight:**
```bash
curl -X OPTIONS http://localhost:8080/api/v1/payments \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST" \
  -v
```

---

### Issue: Backend Service Unreachable

**Cause:** Incorrect service URL or service not running

**Solution:**
```bash
# Test backend connectivity
curl http://localhost:8001/health  # Auth
curl http://localhost:8002/health  # Merchant
curl http://localhost:8004/health  # Payment

# Update URLs in config if needed
services:
  auth:
    url: "http://correct-host:8001"
```

---

### Issue: High Latency

**Cause:** Backend service slow or timeout too high

**Solution:**

1. Check gateway response time header:
```bash
curl -I http://localhost:8080/api/v1/payments/123
# Look for: X-Gateway-Response-Time: 2450ms
```

2. Reduce service timeout:
```yaml
services:
  payment:
    timeout: 10s  # Reduced from 30s
```

3. Monitor circuit breaker - it may be repeatedly trying failed requests

---

## 🔍 Request Tracing Example

```bash
# 1. Send request with custom trace ID
REQUEST_ID=$(uuidgen)
curl -X POST http://localhost:8080/api/v1/payments/authorize \
  -H "X-Request-ID: $REQUEST_ID" \
  -H "X-API-Key: pk_live_..." \
  -d '{"amount": 1000, "currency": "USD", ...}'

# 2. Response includes same trace ID
# X-Request-ID: <same-uuid>

# 3. Check logs across all services for this ID
# Gateway logs:
grep "$REQUEST_ID" logs/gateway.log

# Backend service logs will also have this ID
grep "$REQUEST_ID" logs/payment-api.log
```

---

## 📈 Performance Tuning

### Connection Pooling

```yaml
server:
  read_timeout: 10s
  write_timeout: 10s
  idle_timeout: 120s  # Keep connections alive longer
```

### Rate Limiter Cleanup

The in-memory rate limiter automatically cleans up old buckets every minute. For high traffic, consider Redis:

```yaml
rate_limiting:
  storage: "redis"
  redis_url: "redis://localhost:6379/0"
```

### Circuit Breaker Tuning

For more resilient services, increase thresholds:

```yaml
circuit_breaker:
  auth_service:
    failure_threshold: 10    # More lenient
    timeout: 60s             # Longer recovery time
    success_threshold: 5     # More proof needed
```

---

## 📄 License

Copyright © 2025 Payment Gateway. All rights reserved.

---

## Support

For issues and questions:

- GitHub: https://github.com/rhaloubi/Payment-Gateway-Microservices
- Email: redahaloubi8@gmail.com

---

**Service Version:** v1.0.0  
**Last Updated:** December 2025