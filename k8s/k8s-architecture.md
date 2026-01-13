# Kubernetes Architecture Documentation

## Executive Summary

This document describes a production-ready Kubernetes architecture for a microservices-based payment gateway platform. The architecture implements zero-trust networking, centralized secret management, comprehensive observability, and secure ingress routing across four isolated namespaces.

### Key Features
- **Zero-Trust Networking**: Network policies enforce strict pod-to-pod communication rules
- **Secure Ingress**: NGINX Ingress Controller with external tunnel support
- **Centralized Secrets**: HashiCorp Vault for secure secret management
- **Full Observability**: Prometheus & Grafana for metrics and monitoring
- **Multi-Namespace Isolation**: Services segregated by function and security level

---

## Architecture Overview

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Internet                                 │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                    ┌──────▼──────┐
                    │  Cloudflare │
                    │   Tunnel    │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │  localhost  │
                    │   :30080    │
                    └──────┬──────┘
                           │
┌──────────────────────────┼──────────────────────────────────────┐
│  Kubernetes Cluster      │                                       │
│                          │                                       │
│  ┌───────────────────────▼────────────────────┐                 │
│  │     ingress-nginx Namespace                │                 │
│  │  ┌──────────────────────────────────────┐  │                 │
│  │  │   NGINX Ingress Controller           │  │                 │
│  │  │   - Routing                          │  │                 │
│  │  │   - SSL Termination                  │  │                 │
│  │  │   - Load Balancing                   │  │                 │
│  │  └──────────────┬───────────────────────┘  │                 │
│  └─────────────────┼──────────────────────────┘                 │
│                    │                                             │
│  ┌─────────────────▼──────────────────────────┐                 │
│  │     gateway Namespace                      │                 │
│  │  ┌──────────────────────────────────────┐  │                 │
│  │  │   API Gateway (2 replicas)           │  │                 │
│  │  │   - Request routing                  │  │                 │
│  │  │   - Service aggregation              │  │                 │
│  │  └──────────┬───┬───────────┬───────────┘  │                 │
│  └─────────────┼───┼───────────┼──────────────┘                 │
│                │   │           │                                 │
│  ┌─────────────▼───▼───────────▼──────────────┐                 │
│  │     services Namespace                     │                 │
│  │  ┌──────────────┐  ┌──────────────────┐    │                 │
│  │  │ Auth Service │  │ Merchant Service │    │                 │
│  │  │ (2 replicas) │  │   (2 replicas)   │    │                 │
│  │  │ HTTP + gRPC  │  │      HTTP        │    │                 │
│  │  └──────┬───────┘  └────────┬─────────┘    │                 │
│  │         │                   │               │                 │
│  │  ┌──────▼───────────────────▼─────────┐    │                 │
│  │  │   Payment API Service              │    │                 │
│  │  │      (2 replicas)                  │    │                 │
│  │  │         HTTP                       │    │                 │
│  │  └──────────┬─────────────────────────┘    │                 │
│  └─────────────┼──────────────────────────────┘                 │
│                │                                                 │
│  ┌─────────────▼──────────────────────────────┐                 │
│  │     internal Namespace                     │                 │
│  │  ┌──────────────────┐  ┌────────────────┐  │                 │
│  │  │ Tokenization Svc │  │ Transaction Svc│  │                 │
│  │  │  (2 replicas)    │  │  (2 replicas)  │  │                 │
│  │  │      gRPC        │◄─┤      gRPC      │  │                 │
│  │  └────────┬─────────┘  └────────┬───────┘  │                 │
│  └───────────┼─────────────────────┼──────────┘                 │
│              │                     │                             │
│  ┌───────────▼─────────────────────▼──────────┐                 │
│  │     databases Namespace                    │                 │
│  │  ┌──────────────┐      ┌────────────────┐  │                 │
│  │  │  PostgreSQL  │      │     Redis      │  │                 │
│  │  │ (StatefulSet)│      │ (StatefulSet)  │  │                 │
│  │  └──────────────┘      └────────────────┘  │                 │
│  └────────────────────────────────────────────┘                 │
│                                                                  │
│  ┌──────────────────────────────────────────────┐               │
│  │     monitoring Namespace                     │               │
│  │  ┌────────────┐      ┌──────────────────┐    │               │
│  │  │ Prometheus │      │     Grafana      │    │               │
│  │  │  (Metrics) │      │  (Visualization) │    │               │
│  │  └────────────┘      └──────────────────┘    │               │
│  └──────────────────────────────────────────────┘               │
│                                                                  │
│  ┌──────────────────────────────────────────────┐               │
│  │     vault Namespace                          │               │
│  │  ┌──────────────────────────────────────┐    │               │
│  │  │   HashiCorp Vault                    │    │               │
│  │  │   (Secret Management)                │    │               │
│  │  └──────────────────────────────────────┘    │               │
│  └──────────────────────────────────────────────┘               │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

---

## Namespace Architecture

The cluster is divided into **6 namespaces**, each with specific responsibilities:

### 1. **ingress-nginx** Namespace
- **Purpose**: External traffic entry point
- **Components**: NGINX Ingress Controller
- **Isolation**: Can communicate with gateway namespace only
- **Network Type**: NodePort (30080, 30443)

### 2. **gateway** Namespace
- **Purpose**: API Gateway layer
- **Components**: API Gateway (2 replicas)
- **Isolation**: Receives from ingress-nginx, talks to services namespace
- **Ports**: 8080 (HTTP)

### 3. **services** Namespace
- **Purpose**: Business logic microservices
- **Components**: 
  - Auth Service (HTTP + gRPC)
  - Merchant Service (HTTP)
  - Payment API Service (HTTP)
- **Isolation**: Receives from gateway, talks to internal and databases
- **Ports**: 8001, 8002, 8004, 50051 (gRPC)

### 4. **internal** Namespace
- **Purpose**: Internal-only gRPC services
- **Components**:
  - Tokenization Service (gRPC)
  - Transaction Service (gRPC)
- **Isolation**: Receives from services namespace only, no external access
- **Ports**: 50052, 50053 (gRPC)

### 5. **databases** Namespace
- **Purpose**: Data persistence layer
- **Components**:
  - PostgreSQL (StatefulSet with 5Gi storage)
  - Redis (StatefulSet with 5Gi storage)
- **Isolation**: Only accessible by services and internal namespaces
- **Ports**: 5432 (PostgreSQL), 6379 (Redis)

### 6. **monitoring** Namespace
- **Purpose**: Observability stack
- **Components**:
  - Prometheus (metrics collection)
  - Grafana (visualization)
  - Node Exporter (node metrics)
  - Kube State Metrics (cluster metrics)
- **Isolation**: Can scrape all namespaces, no inbound traffic
- **Ports**: 9090 (Prometheus), 3000 (Grafana)

### 7. **vault** Namespace
- **Purpose**: Secret management
- **Components**: HashiCorp Vault (StatefulSet)
- **Isolation**: Accessible by all service namespaces
- **Ports**: 8200 (API), 8201 (Cluster)

---

## Traffic Flow Deep-Dive

### 1. External Request Flow

```mermaid
sequenceDiagram
    participant User
    participant Cloudflare
    participant NGINX
    participant Gateway
    participant AuthSvc
    participant MerchantSvc
    participant PaymentSvc
    participant TokenSvc
    participant TransSvc
    participant DB

    User->>Cloudflare: HTTPS Request
    Cloudflare->>NGINX: localhost:30080
    NGINX->>Gateway: http://api-gateway:80
    Gateway->>AuthSvc: Validate JWT
    AuthSvc->>Gateway: Token Valid
    Gateway->>MerchantSvc: Get Merchant Details
    MerchantSvc->>Gateway: Merchant Data
    Gateway->>PaymentSvc: Process Payment
    PaymentSvc->>TokenSvc: Tokenize Card (gRPC)
    TokenSvc->>DB: Store Token
    DB->>TokenSvc: Success
    TokenSvc->>PaymentSvc: Token ID
    PaymentSvc->>TransSvc: Create Transaction (gRPC)
    TransSvc->>DB: Store Transaction
    DB->>TransSvc: Success
    TransSvc->>PaymentSvc: Transaction ID
    PaymentSvc->>Gateway: Payment Result
    Gateway->>NGINX: Response
    NGINX->>Cloudflare: Response
    Cloudflare->>User: HTTPS Response
```

### 2. Network Policy Enforcement

Each request must pass through network policy checks:

```
┌─────────────────────────────────────────────────────────────┐
│  Request: User → API Gateway                                 │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  1. NGINX Ingress Controller                                 │
│     ✓ Allows: ALL traffic (external entry point)            │
│     ✓ Network Policy: ingress-nginx-controller-ingress      │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  2. NGINX → API Gateway                                      │
│     ✓ Allows: ingress-nginx namespace → gateway namespace   │
│     ✓ Network Policy: ingress-nginx-controller-egress       │
│     ✓ Network Policy: api-gateway-ingress                   │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  3. API Gateway → Auth Service                               │
│     ✓ Allows: gateway namespace → services namespace        │
│     ✓ Network Policy: api-gateway-egress                    │
│     ✓ Network Policy: auth-service-ingress                  │
│     ✓ Port: 8001 (HTTP) or 50051 (gRPC)                     │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  4. Auth Service → PostgreSQL                                │
│     ✓ Allows: services namespace → databases namespace      │
│     ✓ Network Policy: auth-service-egress                   │
│     ✓ Network Policy: postgres-ingress                      │
│     ✓ Port: 5432                                            │
└─────────────────────────────────────────────────────────────┘
```

### 3. Service-to-Service Communication

#### HTTP Communication (REST APIs)
- **API Gateway → Auth Service**: Port 8001
- **API Gateway → Merchant Service**: Port 8002  
- **API Gateway → Payment API Service**: Port 8004

#### gRPC Communication (Internal)
- **Payment API → Tokenization Service**: Port 50052
- **Payment API → Transaction Service**: Port 50053
- **Transaction Service → Tokenization Service**: Port 50052
- **Services → Auth Service (gRPC)**: Port 50051

#### Database Connections
- **All Services → PostgreSQL**: Port 5432
- **All Services → Redis**: Port 6379

### 4. DNS Resolution Flow

All pods use CoreDNS for service discovery:

```
Service Name Format: <service-name>.<namespace>.svc.cluster.local

Examples:
- auth-service.services.svc.cluster.local → 10.43.224.225:8001
- postgres.databases.svc.cluster.local → 10.43.xxx.xxx:5432
- vault.vault.svc.cluster.local → 10.43.xxx.xxx:8200

Shortened Format (same namespace):
- auth-service → resolves to auth-service.services.svc.cluster.local

Cross-namespace Format:
- auth-service.services → resolves to auth-service.services.svc.cluster.local
```

---

## Network Security Architecture

### Zero-Trust Network Model

The architecture implements a **deny-all-by-default** approach with explicit allow rules:

#### Phase 1: Default Deny
Every namespace starts with:
```yaml
# Deny all ingress traffic
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-all-ingress
  namespace: <namespace>
spec:
  podSelector: {}
  policyTypes:
  - Ingress
```

#### Phase 2: DNS Access
All pods need DNS resolution:
```yaml
# Allow DNS queries to kube-system
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-dns
spec:
  podSelector: {}
  policyTypes:
  - Egress
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: kube-system
      podSelector:
        matchLabels:
          k8s-app: kube-dns
    ports:
    - protocol: UDP
      port: 53
```

#### Phase 3: Service-Specific Rules
Each service has ingress and egress policies:

**Example: Auth Service**
```yaml
# Ingress: Who can talk to Auth Service?
- From: gateway namespace (API Gateway)
  Ports: 8001, 50051
- From: services namespace (other services)
  Ports: 8001, 50051
- From: internal namespace (Tokenization)
  Ports: 8001

# Egress: Who can Auth Service talk to?
- To: databases namespace (PostgreSQL, Redis)
  Ports: 5432, 6379
```

### Network Policy Matrix

| Source Namespace | Target Namespace | Allowed Ports | Purpose |
|-----------------|------------------|---------------|---------|
| ingress-nginx | gateway | 8080 | Route external traffic |
| gateway | services | 8001, 8002, 8004 | API calls |
| services | services | 8001, 50051 | Inter-service gRPC |
| services | internal | 50052, 50053 | gRPC to internal services |
| services | databases | 5432, 6379 | Database access |
| internal | internal | 50052, 50053 | Internal gRPC |
| internal | databases | 5432, 6379 | Database access |
| internal | services | 8001 | Auth validation |
| monitoring | ALL | ALL | Metrics scraping |
| ALL | vault | 8200 | Secret retrieval |

### Security Boundaries

```
┌──────────────────────────────────────────────────────────────┐
│  Security Zone: External                                      │
│  - Cloudflare Tunnel (SSL Termination)                        │
│  - NGINX Ingress (WAF, Rate Limiting)                         │
└────────────────────────┬─────────────────────────────────────┘
                         │ Encrypted (SSL)
                         │
┌────────────────────────▼─────────────────────────────────────┐
│  Security Zone: DMZ (gateway namespace)                       │
│  - API Gateway (Request Validation, JWT Verification)         │
│  - Network Policy: Only from ingress-nginx                    │
└────────────────────────┬─────────────────────────────────────┘
                         │ Internal (Cluster Network)
                         │
┌────────────────────────▼─────────────────────────────────────┐
│  Security Zone: Application (services namespace)              │
│  - Auth, Merchant, Payment Services                           │
│  - Network Policy: Only from gateway                          │
│  - Secrets: From Vault                                        │
└────────────────────────┬─────────────────────────────────────┘
                         │
                         │
┌────────────────────────▼─────────────────────────────────────┐
│  Security Zone: Internal (internal namespace)                 │
│  - Tokenization, Transaction Services                         │
│  - Network Policy: Only from services namespace               │
│  - No external access                                         │
│  - Secrets: From Vault                                        │
└────────────────────────┬─────────────────────────────────────┘
                         │
                         │
┌────────────────────────▼─────────────────────────────────────┐
│  Security Zone: Data (databases namespace)                    │
│  - PostgreSQL, Redis                                          │
│  - Network Policy: Only from services + internal              │
│  - StatefulSets with persistent volumes                       │
│  - No external access                                         │
└──────────────────────────────────────────────────────────────┘
```

---

## Component Details

### NGINX Ingress Controller

**Deployment Type**: Deployment (1 replica)  
**Image**: registry.k8s.io/ingress-nginx/controller:v1.11.1  
**Service Type**: NodePort  
**Ports**: 30080 (HTTP), 30443 (HTTPS)

**Responsibilities**:
- Route incoming requests to API Gateway
- SSL/TLS termination
- Host-based routing (paymentgateway.redahaloubi.com)
- Request buffering and timeouts
- Load balancing across API Gateway replicas

**Configuration**:
- `rewrite-target: /` - Preserve request path
- `ssl-redirect: false` - Allow HTTP (tunnel handles SSL)
- `proxy-body-size: 10m` - Max request size
- Timeouts: 30s for connect/send/read

**RBAC**:
- ClusterRole with access to Ingresses, Services, EndpointSlices
- Required for service discovery and routing updates

### API Gateway

**Deployment Type**: Deployment (2 replicas)  
**Image**: rhaloubi8/payment-gateway-api:latest  
**Service Type**: ClusterIP  
**Port**: 80 → 8080

**Responsibilities**:
- Central entry point for all API requests
- Request routing to backend services
- Service aggregation
- Health check aggregation

**Environment Variables**:
- `AUTH_SERVICE_URL`: http://auth-service.services:8001
- `MERCHANT_SERVICE_URL`: http://merchant-service.services:8002
- `PAYMENT_SERVICE_URL`: http://payment-api-service.services:8004

**Health Checks**:
- Liveness: /health every 20s (start after 40s)
- Readiness: /health every 15s (start after 20s)

**Resource Limits**:
- Request: 100m CPU, 128Mi RAM
- Limit: 200m CPU, 256Mi RAM

### Auth Service

**Deployment Type**: Deployment (2 replicas)  
**Image**: rhaloubi8/payment-gateway-auth:latest  
**Service Type**: ClusterIP  
**Ports**: 8001 (HTTP), 50051 (gRPC)

**Responsibilities**:
- User authentication (JWT)
- Token validation
- User management
- gRPC auth validation for other services

**Dependencies**:
- PostgreSQL: micro_auth database
- Redis: Session storage
- Vault: JWT secret, database credentials

**Database Migration**:
- Job: auth-migrate (runs on deployment)
- Rollback: auth-rollback job available

### Merchant Service

**Deployment Type**: Deployment (2 replicas)  
**Image**: rhaloubi8/payment-gateway-merchant:latest  
**Service Type**: ClusterIP  
**Port**: 8002

**Responsibilities**:
- Merchant account management
- Merchant verification
- Email notifications (Mailtrap)

**Dependencies**:
- PostgreSQL: micro_merchant database
- Redis: Caching
- Auth Service: JWT validation (gRPC)
- Vault: Database credentials, Mailtrap credentials

### Payment API Service

**Deployment Type**: Deployment (2 replicas)  
**Image**: rhaloubi8/payment-gateway-payment-api:latest  
**Service Type**: ClusterIP  
**Port**: 8004

**Responsibilities**:
- Payment processing orchestration
- Checkout session management
- Payment webhook handling

**Dependencies**:
- PostgreSQL: micro_payment database
- Redis: Session storage
- Tokenization Service: Card tokenization (gRPC)
- Transaction Service: Transaction recording (gRPC)
- Auth Service: Authentication (gRPC)
- Vault: Database credentials, checkout URL

**Resource Limits** (Higher than others):
- Request: 150m CPU, 128Mi RAM
- Limit: 300m CPU, 256Mi RAM

### Tokenization Service

**Deployment Type**: Deployment (2 replicas)  
**Image**: rhaloubi8/payment-gateway-tokenization:latest  
**Service Type**: ClusterIP  
**Port**: 50052 (gRPC only)

**Responsibilities**:
- Tokenize sensitive card data
- PCI-DSS compliance layer
- Token validation and retrieval

**Dependencies**:
- PostgreSQL: micro_tokenization database
- Redis: Token caching
- Auth Service: Validation
- Vault: Database credentials

**Security**:
- No HTTP endpoint (gRPC only)
- Internal namespace (no external access)
- Network policy restricted to services namespace

### Transaction Service

**Deployment Type**: Deployment (2 replicas)  
**Image**: rhaloubi8/payment-gateway-transaction:latest  
**Service Type**: ClusterIP  
**Port**: 50053 (gRPC only)

**Responsibilities**:
- Transaction record management
- Transaction history
- Transaction status tracking

**Dependencies**:
- PostgreSQL: micro_transaction database
- Redis: Caching
- Tokenization Service: Token lookup (gRPC)
- Vault: Database credentials

### PostgreSQL

**Deployment Type**: StatefulSet (1 replica)  
**Image**: postgres:15-alpine  
**Service Type**: ClusterIP (Headless)  
**Port**: 5432

**Databases**:
- micro_auth
- micro_merchant
- micro_payment
- micro_tokenization
- micro_transaction

**Storage**: 5Gi persistent volume (ReadWriteOnce)  
**Credentials**: postgres/reda2001 (stored in Vault)

**Bootstrap**: postgres-bootstrap Job creates all databases on first run

### Redis

**Deployment Type**: StatefulSet (1 replica)  
**Image**: redis:7-alpine  
**Service Type**: ClusterIP (Headless)  
**Port**: 6379

**Configuration**:
- AOF persistence enabled
- Append-only mode for durability

**Storage**: 5Gi persistent volume (ReadWriteOnce)  
**Use Cases**: Session storage, caching, rate limiting

---

## Observability Stack

### Prometheus

**Deployment Type**: StatefulSet (managed by kube-prometheus-stack)  
**Service Type**: NodePort (30090)  
**Port**: 9090

**Responsibilities**:
- Metrics collection from all services
- Time-series data storage
- Alert evaluation

**ServiceMonitors**:
Each service exposes `/metrics` endpoint scraped every 30s:
- API Gateway: gateway/api-gateway
- Auth Service: services/auth-service
- Merchant Service: services/merchant-service
- Payment API: services/payment-api-service
- Tokenization: internal/tokenization-service
- Transaction: internal/transaction-service

**Configuration**:
- `serviceMonitorSelectorNilUsesHelmValues=false`: Discover all ServiceMonitors across namespaces
- Scrape interval: 30 seconds
- Retention: 15 days (default)

**Network Access**:
Prometheus can scrape ALL namespaces via network policies

### Grafana

**Deployment Type**: Deployment (managed by kube-prometheus-stack)  
**Service Type**: NodePort (30300)  
**Port**: 3000

**Responsibilities**:
- Metrics visualization
- Dashboard management
- Alert visualization

**Default Credentials**: admin/admin

**Pre-installed Dashboards**:
- Kubernetes cluster monitoring
- Node metrics
- Pod metrics
- Persistent volume metrics

**Data Source**: Prometheus (auto-configured)

### Node Exporter

**Deployment Type**: DaemonSet (runs on every node)  
**Responsibilities**:
- Node-level metrics (CPU, memory, disk, network)
- Hardware metrics
- OS metrics

### Kube State Metrics

**Deployment Type**: Deployment  
**Responsibilities**:
- Kubernetes object metrics
- Pod status, deployments, services
- Resource quotas and limits

---

## Secret Management with Vault

### Vault Architecture

**Deployment Type**: StatefulSet (1 replica)  
**Image**: hashicorp/vault:1.15.4  
**Service Type**: NodePort (30820)  
**Ports**: 8200 (API), 8201 (Cluster)

**Storage Backend**: File storage (1Gi persistent volume)  
**UI**: Enabled at http://localhost:30820/ui

### Initialization Flow

```
┌────────────────────────────────────────────────────────────┐
│  1. Vault Initialization                                   │
│     - Generate unseal key (1 share, threshold 1)          │
│     - Generate root token                                  │
│     - Save keys securely (vault-keys.txt)                 │
└────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────┐
│  2. Vault Unsealing                                        │
│     - Vault starts sealed (encrypted)                      │
│     - Unseal with unseal key                              │
│     - Vault becomes operational                            │
└────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────┐
│  3. Kubernetes Auth Configuration                          │
│     - Enable kubernetes auth method                        │
│     - Configure K8s API endpoint                           │
│     - Vault can now validate service account tokens       │
└────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────┐
│  4. Policy Creation                                        │
│     - Create policy per service (least privilege)         │
│     - auth-service: access to auth/* secrets              │
│     - merchant-service: access to merchant/* secrets      │
│     - payment-api-service: access to payment/* secrets    │
│     - tokenization-service: access to tokenization/*      │
│     - transaction-service: access to transaction/*        │
└────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────┐
│  5. Role Binding                                           │
│     - Bind each service account to its policy             │
│     - Set namespace restrictions                           │
│     - Set TTL (24 hours)                                  │
└────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌────────────────────────────────────────────────────────────┐
│  6. Secret Migration                                       │
│     - Read secrets from Kubernetes Secrets                │
│     - Write to Vault at secret/* path                     │
│     - KV v2 engine (versioned secrets)                    │
└────────────────────────────────────────────────────────────┘
```

### Secret Organization

```
secret/
├── database/
│   ├── auth-dsn
│   ├── merchant-dsn
│   ├── payment-api-dsn
│   ├── tokenization-dsn
│   ├── transaction-dsn
│   └── redis-dsn
└── app/
    ├── jwt-secret
    ├── checkout-url
    ├── mailtrap-username
    └── mailtrap-password
```

### Service Authentication Flow

```
┌────────────────────────────────────────────────────────────┐
│  Service Pod Starts                                        │
└────────────────────────┬───────────────────────────────────┘
                         │
                         ▼
┌────────────────────────────────────────────────────────────┐
│  1. Init Container Runs                                    │
│     - Read K8s service account JWT token                   │
│     - Token location: /var/run/secrets/kubernetes.io/     │
│       serviceaccount/token                                 │
└────────────────────────┬───────────────────────────────────┘
                         │
                         ▼
┌────────────────────────────────────────────────────────────┐
│  2. Authenticate with Vault                                │
│     POST /v1/auth/kubernetes/login                         │
│     Body: { role: "auth-service", jwt: "<token>" }        │
└────────────────────────┬───────────────────────────────────┘
                         │
                         ▼
┌────────────────────────────────────────────────────────────┐
│  3. Vault Validates Token                                  │
│     - Vault calls K8s TokenReview API                      │
│     - Verifies token signature                             │
│     - Checks service account and namespace                 │
└────────────────────────┬───────────────────────────────────┘
                         │
                         ▼
┌────────────────────────────────────────────────────────────┐
│  4. Vault Returns Client Token                             │
│     - Token with attached policy                           │
│     - TTL: 24 hours                                        │
└────────────────────────┬───────────────────────────────────┘
                         │
                         ▼
┌────────────────────────────────────────────────────────────┐
│  5. Fetch Secrets                                          │
│     GET /v1/secret/data/database/auth-dsn                  │
│     Header: X-Vault-Token: <client-token>                 │
└────────────────────────┬───────────────────────────────────┘
                         │
                         ▼
┌────────────────────────────────────────────────────────────┐
│  6. Write Secrets to Files                                 │
│     - /secrets/database-dsn                                │
│     - /secrets/redis-dsn                                   │
│     - /secrets/jwt-secret                                  │
└────────────────────────┬───────────────────────────────────┘
                         │
                         ▼
┌────────────────────────────────────────────────────────────┐
│  7. Main Container Starts                                  │
│     - Reads secrets from files                             │
│     - Example: DATABASE_DSN=file:///secrets/database-dsn   │
└────────────────────────────────────────────────────────────┘
```

### Policy Example

Auth Service Policy:
```hcl
path "secret/data/auth/*" {
  capabilities = ["read", "list"]
}

path "secret/data/database/auth-dsn" {
  capabilities = ["read"]
}

path "secret/data/database/redis-dsn" {
  capabilities = ["read"]
}

path "secret/data/app/jwt-secret" {
  capabilities = ["read"]
}
```

This policy:
- Grants read-only access to auth-specific secrets
- Allows access to shared database DSN
- Allows access to shared JWT secret
- Denies access to all other secrets

---

## Database Architecture

### Database Isolation

Each microservice has its own dedicated database:

```
PostgreSQL Instance
├── micro_auth (Auth Service)
├── micro_merchant (Merchant Service)
├── micro_payment (Payment API Service)
├── micro_tokenization (Tokenization Service)
└── micro_transaction (Transaction Service)
```

**Benefits**:
- **Data isolation**: Service failures don't affect others
- **Independent scaling**: Each DB can be optimized separately
- **Migration independence**: Schema changes don't cascade
- **Security**: Each service has credentials only to its DB

### Connection String Format

```
postgresql://<user>:<password>@postgres.databases:5432/<database>?sslmode=disable

Examples:
- Auth: postgresql://redahaloubi:reda2001@postgres.databases:5432/micro_auth
- Merchant: postgresql://redahaloubi:reda2001@postgres.databases:5432/micro_merchant
```

Stored in Vault as: `secret/database/<service>-dsn`

### Migration Strategy

Each service has two Jobs:
1. **Migration Job**: Runs on deployment, applies schema changes
2. **Rollback Job**: Available for manual rollback if needed

**Migration Flow**:
```
1. Build Docker image with migration binary
2. Job pulls image
3. Job connects to database
4. Job runs migrations (up)
5. Job completes
6. Service deployment starts
```

### Redis Usage

Single Redis instance shared across services:

**Use Cases**:
- **Session storage**: Auth service JWT sessions
- **Rate limiting**: API Gateway rate limits
- **Caching**: Merchant and payment data caching
- **Pub/Sub**: Future real-time notifications

**Connection**: `redis://redis.databases:6379`

---

## Deployment Strategy

### Pod Distribution

```
Namespace: gateway (2 pods)
├── api-gateway-5c5c9b58c6-bk9fn
└── api-gateway-5c5c9b58c6-w7bdx

Namespace: services (6 pods)
├── auth-service-5fbb7db85c-clr8t
├── auth-service-5fbb7db85c-s8sqt
├── merchant-service-7df5597d68-fdbq5
├── merchant-service-7df5597d68-jwzrg
├── payment-api-service-64588486dd-dldjj
└── payment-api-service-64588486dd-jtwn2

Namespace: internal (4 pods)
├── tokenization-service-64f6c6956-bbz7x
├── tokenization-service-64f6c6956-mwvz2
├── transaction-service-7467986c77-7b2nv
└── transaction-service-7467986c77-g8gzf

Namespace: databases (2 pods)
├── postgres-0
└── redis-0

Namespace: ingress-nginx (1 pod)
└── ingress-nginx-controller-78b65c64d-d4sm7

Namespace: monitoring (multiple pods)
├── prometheus-kube-prometheus-prometheus-0
├── grafana-xxx
├── kube-state-metrics-xxx
└── node-exporter-xxx (DaemonSet)

Namespace: vault (1 pod)
└── vault-0
```

### Replica Strategy

**2 Replicas** for all stateless services:
- Provides high availability
- Allows zero-downtime deployments
- Load distribution via ClusterIP services

**1 Replica** for stateful services:
- PostgreSQL (StatefulSet)
- Redis (StatefulSet)
- Vault (StatefulSet)
- NGINX Ingress Controller

**Rationale**: StatefulSets require careful scaling. Start with 1 replica, scale when load requires.

### Rolling Update Strategy

Default Deployment strategy: RollingUpdate
```yaml
strategy:
  type: RollingUpdate
  rollingUpdate:
    maxSurge: 1        # Allow 1 extra pod during update
    maxUnavailable: 0  # Keep all pods running during update
```

**Update Flow**:
1. Create 1 new pod with new version
2. Wait for readiness probe to pass
3. Terminate 1 old pod
4. Repeat for remaining pods

**Zero Downtime**: Always 2 healthy pods serving traffic

### Resource Allocation

Total Cluster Resources Required:

**CPU**:
- Services (6 pods × 100m): 600m
- Gateway (2 pods × 100m): 200m
- Internal (4 pods × 100m): 400m
- Databases (2 pods × 100m): 200m
- NGINX (1 pod × 100m): 100m
- Monitoring (~500m): 500m
- Vault (1 pod × 100m): 100m
- **Total Request**: ~2.1 CPU cores

**Memory**:
- Services (6 pods × 128Mi): 768Mi
- Gateway (2 pods × 128Mi): 256Mi
- Internal (4 pods × 128Mi): 512Mi
- Databases (2 pods × 256Mi): 512Mi
- NGINX (1 pod × 90Mi): 90Mi
- Monitoring (~2Gi): 2048Mi
- Vault (1 pod × 128Mi): 128Mi
- **Total Request**: ~4.3Gi

**Recommended Cluster**: 4 CPU cores, 8Gi RAM minimum

---

## Health Checks & Reliability

### Health Check Strategy

All services implement two types of probes:

#### Liveness Probe
- **Purpose**: Detect if pod is alive
- **Failure Action**: Restart pod
- **Path**: `/health`
- **Initial Delay**: 40 seconds (allow startup time)
- **Period**: 20 seconds
- **Failure Threshold**: 3 (60 seconds before restart)

#### Readiness Probe
- **Purpose**: Detect if pod can serve traffic
- **Failure Action**: Remove from service endpoints
- **Path**: `/health`
- **Initial Delay**: 20 seconds
- **Period**: 15 seconds
- **Failure Threshold**: 3 (45 seconds before removal)

### Service Health Check Flow

```
Pod Starts
    │
    ▼
Wait 20s (readinessProbe.initialDelaySeconds)
    │
    ▼
Readiness Check #1 → /health
    │
    ├─ Success → Pod added to Service endpoints
    │              (starts receiving traffic)
    │
    └─ Failure → Wait 15s, retry
                  (pod not added to Service)
                  
After 40s from start (livenessProbe.initialDelaySeconds)
    │
    ▼
Liveness Check #1 → /health
    │
    ├─ Success → Pod marked healthy
    │
    └─ Failure → Wait 20s, retry
                  (3 failures = restart pod)
```

### API Gateway Health Check

Special aggregated health check:
```json
{
  "service": "api-gateway",
  "status": "ok",
  "version": "1.0.0",
  "services": {
    "auth": "open",
    "merchant": "open", 
    "payment": "open"
  }
}
```

Status per backend:
- `open`: Service reachable and healthy
- `closed`: Service unreachable or unhealthy

---

## Scalability Considerations

### Horizontal Scaling

Services ready for horizontal scaling:
- ✅ API Gateway (stateless)
- ✅ Auth Service (stateless with JWT)
- ✅ Merchant Service (stateless)
- ✅ Payment API Service (stateless)
- ✅ Tokenization Service (stateless)
- ✅ Transaction Service (stateless)

Scale with:
```bash
kubectl scale deployment <service-name> -n <namespace> --replicas=<N>
```

### Stateful Services

Require careful scaling:
- PostgreSQL: Replication setup needed
- Redis: Sentinel or Cluster mode needed
- Vault: HA mode with Raft/Consul needed

### Load Distribution

All ClusterIP services automatically load balance across pods:
```
Client → Service (ClusterIP) → iptables rules → Random pod selection
```

Session affinity: None (stateless design)

### Database Scaling

Current: Single PostgreSQL instance with multiple databases

**Future Options**:
1. **Vertical Scaling**: Increase resources (CPU, RAM, storage)
2. **Read Replicas**: Add read-only replicas for read-heavy services
3. **Database per Service**: Separate PostgreSQL instance per service
4. **Managed Database**: Move to AWS RDS, Cloud SQL, etc.

---

## Security Hardening

### Network Security Layers

1. **External Layer**: Cloudflare (DDoS protection, WAF)
2. **Ingress Layer**: NGINX (rate limiting, request validation)
3. **Application Layer**: JWT authentication, authorization
4. **Network Layer**: Kubernetes NetworkPolicies (pod-to-pod)
5. **Data Layer**: Encrypted secrets (Vault), encrypted at rest

### Pod Security

All pods run with security context:
```yaml
securityContext:
  allowPrivilegeEscalation: false
  runAsUser: 101  # Non-root
  capabilities:
    drop:
    - ALL
    add:
    - NET_BIND_SERVICE  # Only for NGINX
```

### Secret Security

1. **Never in Git**: No secrets in YAML files
2. **Vault Storage**: All secrets in Vault
3. **File-based**: Services read from files (not env vars)
4. **Rotation**: Vault supports automatic rotation
5. **Audit**: Vault logs all secret access

### RBAC

Each component has minimal required permissions:
- NGINX: Read Ingresses, Services, EndpointSlices
- Vault: Read ServiceAccounts, Create TokenReviews
- Services: None (use default ServiceAccount)

---

## Disaster Recovery

### Backup Strategy

**StatefulSets** (require regular backups):
1. **PostgreSQL**: 
   - Daily pg_dump of all databases
   - Store in persistent volume or S3
2. **Redis**:
   - AOF persistence enabled
   - Backup /data volume
3. **Vault**:
   - Backup /vault/data directory
   - Store unseal keys and root token securely

**ConfigMaps/Secrets**:
- Stored in Git (Infrastructure as Code)
- Vault secrets backed up with Vault data

### Recovery Procedures

**Database Recovery**:
```
1. Scale PostgreSQL StatefulSet to 0
2. Delete PVC
3. Restore data to new PVC
4. Scale StatefulSet to 1
5. Verify data integrity
```

**Vault Recovery**:
```
1. Deploy Vault
2. Restore data volume
3. Unseal with unseal key
4. Verify secrets accessible
5. Update service deployments if needed
```

**Service Recovery**:
```
1. Services are stateless
2. Delete deployment
3. Re-apply YAML
4. Pods auto-start with health checks
```

---

## Monitoring & Alerting

### Key Metrics

**Service Level Indicators (SLIs)**:
1. **Availability**: `up` metric from /health endpoints
2. **Latency**: Request duration histograms
3. **Error Rate**: HTTP 5xx responses
4. **Throughput**: Requests per second

**Infrastructure Metrics**:
1. **CPU Usage**: Per pod and per node
2. **Memory Usage**: Per pod and per node
3. **Network I/O**: Bytes sent/received
4. **Disk I/O**: Read/write operations

**Database Metrics**:
1. **Connection Pool**: Active connections
2. **Query Performance**: Slow queries
3. **Replication Lag**: If replicas exist
4. **Storage Usage**: Disk space remaining

### Grafana Dashboards

**Recommended Dashboards**:
1. **Cluster Overview**: 10856 (Kubernetes Cluster Monitoring)
2. **Pod Resources**: 6417 (Kubernetes Pod Resources)
3. **Go Applications**: 10826 (Go Metrics)
4. **PostgreSQL**: 9628 (PostgreSQL Database)
5. **Redis**: 11835 (Redis Dashboard)
6. **NGINX Ingress**: 9614 (NGINX Ingress Controller)

### Alert Rules (Examples)

```yaml
- name: service_availability
  rules:
  - alert: ServiceDown
    expr: up{job="<service>"} == 0
    for: 5m
    annotations:
      summary: "Service {{ $labels.job }} is down"

- name: high_error_rate
  rules:
  - alert: HighErrorRate
    expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
    for: 5m
    annotations:
      summary: "High error rate on {{ $labels.job }}"

- name: high_latency
  rules:
  - alert: HighLatency
    expr: histogram_quantile(0.95, http_request_duration_seconds) > 1
    for: 5m
    annotations:
      summary: "High latency on {{ $labels.job }}"
```

---

## Performance Optimization

### Current Optimizations

1. **Connection Pooling**: All services use connection pools for PostgreSQL/Redis
2. **In-Memory Caching**: Redis for frequently accessed data
3. **gRPC for Internal**: Lower latency than HTTP/REST
4. **Horizontal Scaling**: 2 replicas for load distribution
5. **Health Check Tuning**: Balanced intervals to avoid overhead

### Future Optimizations

1. **Horizontal Pod Autoscaler (HPA)**:
   ```yaml
   apiVersion: autoscaling/v2
   kind: HorizontalPodAutoscaler
   metadata:
     name: api-gateway-hpa
   spec:
     scaleTargetRef:
       apiVersion: apps/v1
       kind: Deployment
       name: api-gateway
     minReplicas: 2
     maxReplicas: 10
     metrics:
     - type: Resource
       resource:
         name: cpu
         target:
           type: Utilization
           averageUtilization: 70
   ```

2. **CDN Integration**: Static assets via Cloudflare
3. **Database Read Replicas**: For read-heavy services
4. **Service Mesh**: Istio for advanced traffic management
5. **Caching Layer**: Redis Cluster or Memcached

---

## Technology Stack Summary

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| Container Runtime | Docker | Latest | Container execution |
| Orchestration | Kubernetes (k3d) | 1.28+ | Container orchestration |
| CNI | Calico | Latest | Network policies |
| Ingress Controller | NGINX | 1.11.1 | Traffic routing |
| Service Mesh | None | - | (Future: Istio) |
| Metrics | Prometheus | Latest | Metrics collection |
| Visualization | Grafana | Latest | Dashboards |
| Secret Management | Vault | 1.15.4 | Centralized secrets |
| Database | PostgreSQL | 15-alpine | Relational data |
| Cache | Redis | 7-alpine | Caching, sessions |
| Programming Language | Go | 1.21+ | Microservices |
| API Protocol | REST + gRPC | - | Communication |
| DNS | CoreDNS | Latest | Service discovery |
| Storage | Local Path | - | Persistent volumes |

---

## Conclusion

This Kubernetes architecture provides a production-ready foundation for a microservices payment gateway with:

✅ **Security**: Zero-trust networking, centralized secrets, RBAC  
✅ **Observability**: Full metrics coverage, real-time monitoring  
✅ **Reliability**: Health checks, 2 replicas, persistent data  
✅ **Scalability**: Horizontal scaling ready, stateless design  
✅ **Maintainability**: Infrastructure as Code, clear namespace separation

**Current State**: 
- ✅ Phase 1: Network Policies (Zero-Trust)
- ✅ Phase 2: NGINX Ingress (Secure Routing)
- ✅ Phase 3: Prometheus + Grafana (Observability)
- ✅ Phase 4: HashiCorp Vault (Secret Management)

**Future Enhancements**:
- Phase 5: CI/CD Automation (GitHub Actions)
- Phase 6: AWS Deployment (Terraform)
- Phase 7: Advanced Features (HPA, Service Mesh, Multi-region)

---

**Document Version**: 1.0  
**Last Updated**: January 2026  
**Architecture Status**: Production-Ready (Local k3d)