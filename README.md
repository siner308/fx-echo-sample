# FX Echo Sample

> A modern Go web application built with Echo framework, Uber FX dependency injection, and item reward system

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Echo](https://img.shields.io/badge/Echo-v4-brightgreen.svg)](https://echo.labstack.com)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## 🚀 Overview

FX Echo Sample is a production-ready RESTful API server that demonstrates modern Go web development practices. It features a comprehensive item reward system with multi-tier authentication, modular architecture, and industry-standard API design patterns.

### ✨ Key Features

- **🔐 Multi-Token Authentication**: Access, Refresh, and Admin JWT tokens
- **🏢 Enterprise SSO**: Keycloak integration for admin authentication
- **🎁 Item Reward System**: Coupon redemption and payment-based item distribution
- **🏗️ Modular Architecture**: Clean domain separation with dependency injection
- **🛡️ Security First**: Argon2id password hashing, type-safe contexts
- **📊 Stripe-Style APIs**: Industry-standard response formats

## 🏛️ Architecture

Built with modern software engineering principles:

- **Framework**: Echo v4 (High-performance HTTP router)
- **Dependency Injection**: Uber FX (Lifecycle management)
- **Authentication**: JWT + Keycloak SSO
- **Security**: Argon2id password hashing
- **Design Pattern**: Domain-Driven Design (DDD)
- **API Style**: Stripe-inspired response format

## 🛠️ Quick Start

### Prerequisites

- Go 1.21 or higher
- Git

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/siner308/fx-echo-sample.git
   cd fx-echo-sample
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Run the application**
   ```bash
   go run main.go
   ```

The server will start on `http://localhost:8080`

## 🔧 Environment Configuration

Create a `.env` file with the following required variables:

```bash
# JWT Secrets (use strong, random values)
ACCESS_TOKEN_SECRET=your-access-token-secret-32-chars-min
REFRESH_TOKEN_SECRET=your-refresh-token-secret-32-chars-min
ADMIN_TOKEN_SECRET=your-admin-token-secret-32-chars-min

# Server Configuration
PORT=8080
GIN_MODE=debug

# Keycloak Configuration (Optional)
KEYCLOAK_BASE_URL=http://localhost:8080
KEYCLOAK_REALM=your-realm
KEYCLOAK_CLIENT_ID=your-client-id
KEYCLOAK_CLIENT_SECRET=your-client-secret

# Database (Future)
# DATABASE_URL=postgres://user:pass@localhost/dbname
```

## 📚 API Documentation

### Authentication Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `POST` | `/api/v1/auth/user/login` | User login | ❌ |
| `POST` | `/api/v1/auth/user/refresh` | Refresh access token | ❌ |
| `GET`  | `/api/v1/auth/admin/sso/auth-url` | Get Keycloak auth URL | ❌ |
| `POST` | `/api/v1/auth/admin/sso/callback` | Handle SSO callback | ❌ |
| `GET`  | `/api/v1/auth/admin/me` | Get admin info | 🔑 Admin |

### User Management

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `POST` | `/api/v1/users` | Create user account | ❌ |
| `GET`  | `/api/v1/users/me` | Get my profile | 🔑 User |
| `GET`  | `/api/v1/users/:id` | Get user by ID | 🔑 User |
| `PUT`  | `/api/v1/users/:id` | Update user | 🔑 User |
| `DELETE` | `/api/v1/users/:id` | Delete user | 🔑 User |
| `GET`  | `/api/v1/users` | List all users | 🔑 Admin |

### Item System

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `GET`  | `/api/v1/items` | List items | 🔑 User |
| `GET`  | `/api/v1/items/:id` | Get item details | 🔑 User |
| `GET`  | `/api/v1/items/types` | Get item types | 🔑 User |
| `GET`  | `/api/v1/users/:id/inventory` | Get user inventory | 🔑 User |
| `POST` | `/api/v1/admin/items` | Create item | 🔑 Admin |
| `PUT`  | `/api/v1/admin/items/:id` | Update item | 🔑 Admin |
| `DELETE` | `/api/v1/admin/items/:id` | Delete item | 🔑 Admin |

### Payment & Rewards

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `POST` | `/api/v1/payments` | Process payment | 🔑 User |
| `GET`  | `/api/v1/payments/:id` | Get payment | 🔑 User |
| `POST` | `/api/v1/coupons/redeem` | Redeem coupon | 🔑 User |
| `POST` | `/api/v1/admin/rewards/grant` | Grant rewards | 🔑 Admin |

## 📋 API Response Format

This API follows Stripe-style response patterns for consistency and developer experience.

### Success Responses

**Single Resource:**
```json
{
  "id": 123,
  "name": "Magic Sword",
  "type": "equipment",
  "rarity": "legendary"
}
```

**List Resources:**
```json
{
  "object": "list",
  "data": [
    {"id": 1, "name": "Item 1"},
    {"id": 2, "name": "Item 2"}
  ],
  "has_more": false,
  "total_count": 2
}
```

**Delete Operation:**
```json
{
  "deleted": true,
  "id": "123"
}
```

### Error Responses

```json
{
  "error": {
    "type": "validation_error",
    "code": "invalid_parameter",
    "message": "Name is required",
    "param": "name"
  }
}
```

**Error Types:**
- `validation_error` - Input validation failed
- `authentication_error` - Invalid credentials
- `invalid_request_error` - Malformed request
- `api_error` - Internal server error

## 🏗️ Project Structure

```
fx-echo-sample/
├── docs/                    # Comprehensive documentation
│   ├── ARCHITECTURE.md      # System architecture
│   ├── API_REFERENCE.md     # Detailed API docs
│   ├── FX_CONCEPTS.md       # Uber FX patterns
│   └── SECURITY_GUIDE.md    # Security best practices
├── modules/                 # Domain modules
│   ├── auth/               # Authentication (user + admin)
│   ├── user/               # User management
│   ├── item/               # Item system
│   ├── payment/            # Payment processing
│   ├── coupon/             # Coupon system
│   └── reward/             # Reward distribution
├── pkg/                    # Shared packages
│   ├── dto/                # Data transfer objects
│   ├── jwt/                # JWT utilities
│   ├── security/           # Password hashing
│   └── validator/          # Input validation
├── middleware/             # HTTP middleware
└── server/                 # Server setup
```

## 🧪 Testing

Run the test suite:

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific module tests
go test ./modules/user/...

# Run integration tests
go test ./modules/auth/user/ -run Integration
```

### Test Coverage

- ✅ **Service Layer**: Business logic unit tests
- ✅ **Integration**: JWT token validation
- ✅ **Security**: Password hashing benchmarks
- ⚠️ **Handler Layer**: HTTP response testing (TODO)

## 📖 Learning Resources

### For New Team Members

1. **Start Here**: [`docs/FX_CONCEPTS.md`](docs/FX_CONCEPTS.md) - Understand Uber FX basics
2. **Architecture**: [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) - System design overview
3. **Development**: [`.claude/CLAUDE.md`](.claude/CLAUDE.md) - Latest changes and patterns

### Key Concepts

- **Dependency Injection**: How FX manages application lifecycle
- **Type-Safe Contexts**: Secure data passing in HTTP handlers
- **Domain Separation**: Clean boundaries between business modules
- **Authentication Flow**: Multi-token JWT strategy

## 🔒 Security Features

- **🔐 Argon2id Hashing**: Industry-standard password security
- **🎫 JWT Multi-Token**: Separate access/refresh/admin tokens
- **🏢 Enterprise SSO**: Keycloak integration
- **🛡️ Type Safety**: Compile-time context key validation
- **⏱️ Timing Attack Prevention**: Consistent password verification time

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow the existing modular structure
- Add tests for new functionality
- Update documentation for API changes
- Use the established error response patterns

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Echo Framework](https://echo.labstack.com/) - High-performance HTTP router
- [Uber FX](https://uber-go.github.io/fx/) - Dependency injection framework
- [Stripe API](https://stripe.com/docs/api) - API design inspiration
- [Keycloak](https://www.keycloak.org/) - Identity and access management

## 📞 Support

- 📖 **Documentation**: [`docs/`](docs/) directory
- 🐛 **Issues**: [GitHub Issues](https://github.com/siner308/fx-echo-sample/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/siner308/fx-echo-sample/discussions)

---

**Built with ❤️ using modern Go practices and industry-proven patterns.**