# Auth Service - Complete Documentation

## 📚 Table of Contents

1. [Overview](#overview)
2. [Features](#features)
3. [Architecture](#architecture)
4. [Setup & Installation](#setup--installation)
5. [API Documentation](#api-documentation)
6. [Testing](#testing)
7. [Security Features](#security-features)
8. [Database Schema](#database-schema)

---

## Overview

The Auth Service handles authentication, authorization, and access control for the payment gateway platform. It provides secure user registration, login, session management, role-based access control (RBAC), and API key management.

### Key Features

- ✅ User registration and authentication
- ✅ JWT-based session management
- ✅ Role-based access control (RBAC)
- ✅ Permission-based authorization
- ✅ API key management
- ✅ Redis caching for performance
- ✅ Account security (lockout, password requirements)
- ✅ Multi-tenant support

---

## Features

### 1. Authentication

- **User Registration**: Email-based registration with password hashing (bcrypt)
- **Login**: Secure authentication with JWT tokens
- **Token Refresh**: Automatic token renewal without re-login
- **Logout**: Session revocation (single device or all devices)
- **Password Change**: Secure password updates with re-authentication

### 2. Authorization (RBAC)

- **Roles**: Predefined roles (Owner, Admin, Manager, Staff)
- **Permissions**: Granular permissions (resource:action format)
- **Multi-tenant**: Different roles per merchant for same user
- **Dynamic Permission Checking**: Fast permission validation with Redis caching

### 3. API Key Management

- **API Key Generation**: Secure random key generation
- **Key Validation**: SHA-256 hashed storage
- **Usage Tracking**: Last used timestamp tracking
- **Key Management**: Activate/deactivate/delete keys

### 4. Security Features

- **Password Hashing**: bcrypt with default cost (10)
- **Account Lockout**: 5 failed attempts = 30-minute lock
- **JWT Security**: HS256 signing, 24h expiry for access tokens
- **Session Tracking**: IP address and user agent logging
- **Redis Caching**: Fast permission checks and session validation

---

## Architecture

### Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin (HTTP router)
- **Database**: PostgreSQL 14+
- **Cache**: Redis 7+
- **ORM**: GORM
- **Authentication**: JWT (golang-jwt/jwt)

### Layers

```
┌─────────────────────────────────────┐
│         HTTP Handlers               │  (REST API endpoints)
│  - auth_handler.go                  │
│  - role_handler.go                  │
│  - api_key_handler.go               │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│         Service Layer               │  (Business logic)
│  - auth_service.go                  │
│  - user_service.go                  │
│  - role_service.go                  │
│  - api_key_service.go               │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│      Repository Layer               │  (Data access)
│  - user_repository.go               │
│  - role_repository.go               │
│  - session_repository.go            │
│  - api_key_repository.go            │
└──────────────┬──────────────────────┘
               │
      ┌────────┴────────┐
      │                 │
┌─────▼──────┐  ┌──────▼─────┐
│ PostgreSQL │  │   Redis    │
└────────────┘  └────────────┘
```

---

## Setup & Installation

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 14+
- Redis 7+
- Docker (optional)

### Environment Variables

Create a `.env` file:

```bash
# Database
DATABASE_DSN=postgresql://user:password@localhost:5432/payment_gateway?sslmode=disable

# Redis
REDIS_DSN=redis://localhost:6379/0

# JWT
JWT_SECRET_KEY=your-super-secret-jwt-key-minimum-32-characters

# Server
PORT=8001
GIN_MODE=release
```

### Installation Steps

#### 1. Clone and Install Dependencies

```bash
cd payment-gateway/auth-service
go mod download
```

#### 2. Start Infrastructure (Docker)

```bash
# Start PostgreSQL and Redis
docker-compose up -d postgres redis
```

#### 3. Run Migrations

```bash
go run cmd/migrate up

```
### rollback 

```bash
go run cmd/migrate down

```

#### 4. Start Server

````bash
air init && air

#### Or
```bash
go run cmd/main.go



# Output:
# ✅ Database connected
# ✅ Redis connected
# 🚀 Auth service starting on :8001
````

### Health Check

```bash
curl http://localhost:8001/health

# Response:
{
  "status": "ok",
  "service": "auth-service"
}
```

---

## API Documentation

### Base URL

```
http://localhost:8001/api/v1
```

### Authentication

Most endpoints require authentication via JWT token:

```
Authorization: Bearer <access_token>
```

Or API key for server-to-server:

```
X-API-Key: <api_key>
```

---

### 📋 Auth Endpoints

#### 1. Register User

**POST** `/auth/register`

Register a new user account.

**Request Body:**

```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "SecurePass123!"
}
```

**Response:** `201 Created`

```json
{
  "success": true,
  "data": {
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "John Doe",
      "email": "john@example.com",
      "email_verified": false,
      "status": "pending_verification",
      "created_at": "2025-11-08T10:00:00Z"
    }
  },
  "message": "Registration successful. Please verify your email."
}
```

**Validation Rules:**

- `name`: Required
- `email`: Required, valid email format
- `password`: Required, minimum 8 characters

---

#### 2. Login

**POST** `/auth/login`

Authenticate user and receive JWT tokens.

**Request Body:**

```json
{
  "email": "john@example.com",
  "password": "SecurePass123!"
}
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "John Doe",
      "email": "john@example.com",
      "email_verified": false,
      "status": "active"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400
  }
}
```

**Error Responses:**

- `401 Unauthorized`: Invalid credentials
- `401 Unauthorized`: Account locked (too many failed attempts)
- `401 Unauthorized`: Account suspended

---

#### 3. Refresh Token

**POST** `/auth/refresh`

Get a new access token using refresh token.

**Request Body:**

```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400
  }
}
```

---

#### 4. Get Profile

**GET** `/auth/profile`

Get authenticated user's profile.

**Headers:**

```
Authorization: Bearer <access_token>
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "John Doe",
      "email": "john@example.com",
      "email_verified": false,
      "status": "active",
      "created_at": "2025-11-08T10:00:00Z"
    }
  }
}
```

---

#### 5. Logout

**POST** `/auth/logout`

Revoke current session.

**Headers:**

```
Authorization: Bearer <access_token>
```

**Response:** `200 OK`

```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

---

#### 6. Change Password

**POST** `/auth/change-password`

Change user's password (forces logout from all devices).

**Headers:**

```
Authorization: Bearer <access_token>
```

**Request Body:**

```json
{
  "old_password": "SecurePass123!",
  "new_password": "NewSecurePass456!"
}
```

**Response:** `200 OK`

```json
{
  "success": true,
  "message": "Password changed successfully. Please login again."
}
```

---

#### 7. Get Sessions

**GET** `/auth/sessions`

Get all active sessions for the user.

**Headers:**

```
Authorization: Bearer <access_token>
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "sessions": [
      {
        "id": "session-uuid",
        "ip_address": "192.168.1.1",
        "user_agent": "Mozilla/5.0...",
        "created_at": "2025-11-08T10:00:00Z",
        "expires_at": "2025-11-09T10:00:00Z"
      }
    ]
  }
}
```

---

### 👥 Role & Permission Endpoints

#### 8. Get All Roles

**GET** `/roles`

Get list of all available roles.

**Headers:**

```
Authorization: Bearer <access_token>
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "roles": [
      {
        "id": "role-uuid",
        "name": "Owner",
        "description": "Merchant owner - full access"
      },
      {
        "id": "role-uuid",
        "name": "Admin",
        "description": "Full access to payments, invoices, team, and settings"
      },
      {
        "id": "role-uuid",
        "name": "Manager",
        "description": "Can manage payments and invoices"
      },
      {
        "id": "role-uuid",
        "name": "Staff",
        "description": "Can only view and create transactions"
      }
    ]
  }
}
```

---

#### 9. Get Role Details

**GET** `/roles/:id`

Get role with all its permissions.

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "role": {
      "id": "role-uuid",
      "name": "Manager",
      "description": "Can manage payments and invoices",
      "permissions": [
        {
          "id": "perm-uuid",
          "resource": "transactions",
          "action": "read",
          "description": "View transaction details"
        },
        {
          "id": "perm-uuid",
          "resource": "transactions",
          "action": "create",
          "description": "Create new transactions"
        }
      ]
    }
  }
}
```

---

#### 10. Assign Role to User

**POST** `/roles/assign`

Assign a role to a user for a specific merchant.

**Request Body:**

```json
{
  "user_id": "user-uuid",
  "role_id": "role-uuid",
  "merchant_id": "merchant-uuid"
}
```

**Response:** `200 OK`

```json
{
  "success": true,
  "message": "Role assigned successfully"
}
```

---

#### 11. Remove Role from User

**DELETE** `/roles/assign`

Remove a role from a user.

**Request Body:**

```json
{
  "user_id": "user-uuid",
  "role_id": "role-uuid",
  "merchant_id": "merchant-uuid"
}
```

**Response:** `200 OK`

```json
{
  "success": true,
  "message": "Role removed successfully"
}
```

---

#### 12. Get User Roles

**GET** `/roles/user/:user_id/merchant/:merchant_id`

Get all roles for a user in a specific merchant.

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "roles": [
      {
        "id": "role-uuid",
        "name": "Admin",
        "description": "Full access"
      }
    ]
  }
}
```

---

#### 13. Get User Permissions

**GET** `/roles/user/:user_id/merchant/:merchant_id/permissions`

Get all permissions for a user in a specific merchant.

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "permissions": [
      {
        "resource": "transactions",
        "action": "read"
      },
      {
        "resource": "transactions",
        "action": "create"
      },
      {
        "resource": "invoices",
        "action": "read"
      }
    ]
  }
}
```

---

### 🔑 API Key Endpoints

#### 14. Create API Key

**POST** `/api-keys`

Generate a new API key for a merchant.

**Request Body:**

```json
{
  "merchant_id": "merchant-uuid",
  "name": "Production API Key"
}
```

**Response:** `201 Created`

```json
{
  "success": true,
  "data": {
    "api_key": {
      "id": "key-uuid",
      "name": "Production API Key",
      "key_prefix": "pk_",
      "created_at": "2025-11-08T10:00:00Z"
    },
    "plain_key": "pk_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  },
  "message": "⚠️ Save this API key! It won't be shown again."
}
```

**⚠️ Important**: The `plain_key` is only shown once. Store it securely!

---

#### 15. List Merchant API Keys

**GET** `/api-keys/merchant/:merchant_id`

Get all API keys for a merchant.

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "api_keys": [
      {
        "id": "key-uuid",
        "name": "Production API Key",
        "key_prefix": "pk_",
        "is_active": true,
        "last_used_at": "2025-11-08T12:00:00Z",
        "created_at": "2025-11-08T10:00:00Z"
      }
    ]
  }
}
```

---

#### 16. Deactivate API Key

**PATCH** `/api-keys/:id/deactivate`

Deactivate an API key (doesn't delete it).

**Response:** `200 OK`

```json
{
  "success": true,
  "message": "API key deactivated successfully"
}
```

---

#### 17. Delete API Key

**DELETE** `/api-keys/:id`

Permanently delete an API key.

**Response:** `200 OK`

```json
{
  "success": true,
  "message": "API key deleted successfully"
}
```

---

## Testing

### Unit Tests

Run all tests:

```bash
go test ./internal/auth/... -v -cover
```

Run specific test:

```bash
go test ./internal/auth/service -v -run TestRegister
```

Generate coverage report:

```bash
go test ./internal/auth/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Integration Tests

```bash
# Start test database
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
go test ./test/integration/... -v

# Cleanup
docker-compose -f docker-compose.test.yml down
```

### Test Coverage Goals

- Service Layer: > 80%
- Repository Layer: > 70%
- Handlers: > 75%

---

## Security Features

### 1. Password Security

- **Algorithm**: bcrypt with cost 10
- **Requirements**: Minimum 8 characters
- **Storage**: Only hashed passwords stored
- **Salt**: Automatic per-password salt

### 2. Account Protection

- **Failed Login Limit**: 5 attempts
- **Lockout Duration**: 30 minutes
- **Unlock**: Automatic after timeout or manual by admin

### 3. JWT Security

- **Algorithm**: HS256 (HMAC-SHA256)
- **Access Token Expiry**: 24 hours
- **Refresh Token Expiry**: 7 days
- **Token Storage**: Hashed in database (SHA-256)
- **Session Tracking**: IP address and user agent logged

### 4. API Key Security

- **Generation**: Cryptographically secure random
- **Storage**: SHA-256 hashed
- **Format**: `pk_{32_random_chars}`
- **Exposure**: Plain key shown only once

### 5. Data Protection

- **SQL Injection**: Parameterized queries (GORM)
- **XSS**: Input sanitization
- **CSRF**: Token-based protection
- **Rate Limiting**: Per-endpoint limits (future)

---

## Database Schema

### Core Tables

#### users

```sql
- id (UUID, PK)
- name (VARCHAR)
- email (VARCHAR, UNIQUE)
- email_verified (BOOLEAN)
- password_hash (VARCHAR)
- status (ENUM: active, suspended, pending_verification)
- failed_login_attempts (INTEGER)
- locked_until (TIMESTAMP)
- last_login_at (TIMESTAMP)
- last_login_ip (VARCHAR)
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)
- deleted_at (TIMESTAMP) -- Soft delete
```

#### roles

```sql
- id (UUID, PK)
- name (VARCHAR, UNIQUE)
- description (TEXT)
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)
```

#### permissions

```sql
- id (UUID, PK)
- resource (VARCHAR) -- e.g., 'transactions'
- action (VARCHAR)   -- e.g., 'read', 'create'
- description (TEXT)
- created_at (TIMESTAMP)
```

#### user_roles (Junction Table)

```sql
- user_id (UUID, FK)
- role_id (UUID, FK)
- merchant_id (UUID) -- From merchant service
- assigned_by (UUID, FK)
- assigned_at (TIMESTAMP)
- PRIMARY KEY (user_id, role_id)
```

#### sessions

```sql
- id (UUID, PK)
- user_id (UUID, FK)
- jwt_token (TEXT, HASHED)
- ip_address (VARCHAR)
- user_agent (TEXT)
- expires_at (TIMESTAMP)
- is_revoked (BOOLEAN)
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)
```

#### api_keys

```sql
- id (UUID, PK)
- merchant_id (UUID)
- key_hash (VARCHAR, UNIQUE, HASHED)
- key_prefix (VARCHAR)
- name (VARCHAR)
- is_active (BOOLEAN)
- expires_at (TIMESTAMP)
- last_used_at (TIMESTAMP)
- created_by (UUID, FK)
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)
```

---

## Performance & Caching

### Redis Caching Strategy

**Cached Data:**

- User profiles (TTL: 15 min)
- Roles (TTL: 30 min)
- Permissions (TTL: 10 min)
- Sessions (TTL: based on expiry)
- API key validation (TTL: 30 min)

**Cache Keys:**

```
user:id:{uuid}
user:email:{email}
role:id:{uuid}
role:name:{name}
user:roles:{user_id}:{merchant_id}
user:permissions:{user_id}:{merchant_id}
session:token:{hash}
apikey:hash:{hash}
```

**Cache Invalidation:**

- On update/delete operations
- On role/permission changes
- On session revocation

---

## Error Handling

### Standard Error Response

```json
{
  "success": false,
  "error": "descriptive error message"
}
```

### HTTP Status Codes

- `200 OK`: Success
- `201 Created`: Resource created
- `400 Bad Request`: Invalid input
- `401 Unauthorized`: Authentication required/failed
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

---

## Monitoring & Logging

### Structured Logging

All logs include:

- Timestamp
- Log level (INFO, WARN, ERROR)
- User ID (if authenticated)
- Request ID
- IP address
- Action performed

### Metrics (Future)

- Login success/failure rate
- API key usage
- Permission check latency
- Cache hit/miss rate
- Session count

---

## Development

### Project Structure

```
internal/
├── handler/           # HTTP handlers
├── service/           # Business logic
├── repository/        # Data access
├── model/             # Database models
├── middleware/        # HTTP middleware
├── migration/         # Database migrations
└── routes/            # Route definitions
```

### Adding New Permissions

```sql
INSERT INTO permissions (resource, action, description)
VALUES ('reports', 'export', 'Export report data');
```

### Adding Permissions to Role

```go
roleService := service.NewRoleService()
roleService.AssignPermissionToRole(roleID, permissionID)
```

---

## Troubleshooting

### Common Issues

**1. Database Connection Failed**

```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Check connection string
echo $DATABASE_DSN
```

**2. Redis Connection Failed**

```bash
# Check Redis is running
docker ps | grep redis

# Test connection
redis-cli ping
```

**3. JWT Token Invalid**

- Check `JWT_SECRET_KEY` environment variable
- Ensure token hasn't expired
- Verify token format: `Bearer <token>`

**4. Permission Denied**

- Check user has correct role assigned
- Verify role has required permissions
- Check merchant_id in context

---

## Support

For issues and questions:

- GitHub : https://github.com/rhaloubi/Payment-Gateway-Microservices
- Email: redahaloubi8@gmail.com
