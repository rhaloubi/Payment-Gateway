# 🏪 Merchant Service

**Merchant Management & Onboarding**

The Merchant Service is the central hub for managing merchant accounts, business profiles, team members, and configuration settings. It handles the complete lifecycle of a merchant on the platform.

---

## 📚 Table of Contents

1. [Overview](#overview)
2. [Features](#features)
3. [Architecture](#architecture)
4. [Setup & Installation](#setup--installation)
5. [API Documentation](#api-documentation)
6. [Database Schema](#database-schema)

---

## Overview

The Merchant Service enables users to create and manage multiple business entities (merchants). It provides granular access control for team members, manages technical settings (webhooks, API keys), and handles business verification data.

### Key Capabilities

- ✅ **Multi-Merchant Support**: One user can own or join multiple merchants
- ✅ **Team Management**: Invite members with specific roles (Admin, Manager, Staff)
- ✅ **Role-Based Access**: Granular permissions per merchant context
- ✅ **Configuration**: Manage payment methods, currencies, and branding
- ✅ **API Key Management**: Generate and manage keys for the Payment API
- ✅ **Invitation System**: Secure email-based invitation flow

---

## Features

### 1. Merchant Management
- **Onboarding**: Create new merchant profiles with business details
- **Business Profile**: Manage legal name, website, phone, and address
- **Multi-Tenancy**: Users can switch between different merchant contexts

### 2. Team Collaboration
- **Roles**:
  - **Owner**: Full access, cannot be removed
  - **Admin**: Manage team and settings
  - **Manager**: Manage operations
  - **Staff**: View-only or limited operational access
- **Invitations**: Send email invitations to join the team
- **Member Management**: Update roles or remove members

### 3. Technical Settings
- **API Keys**: Generate and revoke keys for API access (integrated with Auth Service)
- **Webhooks**: Configure URLs and secrets for event notifications
- **Payment Settings**: Toggle payment methods and currencies

### 4. Branding & Localization
- **Branding**: Set logos and brand colors (planned)
- **Localization**: Default currency and timezone settings

---

## Architecture

### Tech Stack

- **Language**: Go 1.23+
- **Framework**: Gin (HTTP router)
- **Database**: PostgreSQL 14+
- **Cache**: Redis 7+
- **ORM**: GORM
- **Communication**: gRPC (to Auth Service)

### Service Dependencies

- **Auth Service**:
  - Validates user tokens
  - Manages underlying user accounts
  - Handles API key generation (via gRPC)
  - Manages Roles & Permissions

---

## Setup & Installation

### Prerequisites

- Go 1.23+
- PostgreSQL 14+
- Redis 7+
- Auth Service (running)

### Environment Variables

Create a `.env` file:

```bash
# Server
PORT=8002
GIN_MODE=debug

# Database
DATABASE_DSN=postgresql://user:password@localhost:5432/merchant_db?sslmode=disable

# Redis
REDIS_DSN=redis://localhost:6379/0

# Auth Service Integration
AUTH_SERVICE_URL=http://localhost:8001
AUTH_SERVICE_GRPC_URL=localhost:50051

# JWT (for validation)
JWT_SECRET_KEY=your-super-secret-jwt-key
```

### Installation Steps

```bash
# 1. Clone repository
cd payment-gateway/merchant-service

# 2. Install dependencies
go mod download

# 3. Run migrations
go run internal/migrations/migrate.go

# 4. Start service
go run cmd/main.go
```

---

## API Documentation

### Base URL
`http://localhost:8002/api/v1`

### Authentication
All endpoints require a valid JWT token from the Auth Service:
```
Authorization: Bearer <access_token>
```

### 🏢 Merchant Endpoints

#### Create Merchant
**POST** `/merchants`
```json
{
  "business_name": "Acme Corp",
  "email": "contact@acme.com",
  "business_type": "corporation"
}
```

#### Get My Merchants
**GET** `/merchants`
Returns a list of merchants the current user belongs to.

#### Get Merchant Details
**GET** `/merchants/:id`
Returns full profile of a specific merchant.

#### Update Merchant
**PATCH** `/merchants/:id`
```json
{
  "business_name": "Acme Inc.",
  "website": "https://acme.com"
}
```

### 👥 Team Endpoints

#### List Team Members
**GET** `/merchants/:id/team`

#### Invite Member
**POST** `/merchants/:id/team/invite`
```json
{
  "email": "colleague@acme.com",
  "role_name": "manager",
  "role_id": "id"
}
```

#### Update Member Role
**PATCH** `/merchants/:id/team/:user_id`
```json
{
  "role_name": "admin",
  "role_id": "id"
}
```

#### Remove Member
**DELETE** `/merchants/:id/team/:user_id`

### 🔑 API Key Endpoints

#### Create API Key
**POST** `/merchants/api-keys`
```json
{
  "merchant_id": "uuid",
  "name": "Production Key"
}
```

#### List API Keys
**GET** `/merchants/api-keys/merchant/:merchant_id`

#### Deactivate API Key
**PATCH** `/merchants/api-keys/:id/deactivate`

#### Delete API Key
**DELETE** `/merchants/api-keys/:id`

### ⚙️ Settings Endpoints

#### Get Settings
**GET** `/merchants/:id/settings`

#### Update Settings
**PATCH** `/merchants/:id/settings`
```json
{
  "default_currency": "USD",
  "webhook_url": "https://api.acme.com/webhooks"
}
```

---

## Database Schema

### Core Tables

#### `merchants`
- `id` (UUID, PK)
- `owner_id` (UUID) - Reference to Auth Service user
- `merchant_code` (VARCHAR) - Unique identifier (e.g., mch_...)
- `business_name` (VARCHAR)
- `status` (ENUM: pending_review, active, suspended)
- `country_code` (CHAR(2))
- `created_at`, `updated_at`

#### `merchant_users`
- `id` (UUID, PK)
- `merchant_id` (UUID, FK)
- `user_id` (UUID) - Reference to Auth Service user
- `role_id` (UUID) - Reference to Auth Service role
- `status` (ENUM: active, pending, suspended)
- `joined_at` (TIMESTAMP)

#### `merchant_settings`
- `id` (UUID, PK)
- `merchant_id` (UUID, FK)
- `payment_methods` (JSONB)
- `currencies` (JSONB)
- `webhook_url` (VARCHAR)
- `webhook_secret` (VARCHAR)

#### `merchant_invitations`
- `id` (UUID, PK)
- `merchant_id` (UUID, FK)
- `email` (VARCHAR)
- `role_id` (UUID)
- `token` (VARCHAR)
- `expires_at` (TIMESTAMP)

---

## Support

For issues and questions:
- GitHub: https://github.com/rhaloubi/Payment-Gateway-Microservices
- Email: redahaloubi8@gmail.com
