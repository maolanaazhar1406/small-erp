# 🏢 Small ERP System

A fully functional, containerized Small ERP (Enterprise Resource Planning) application built with **Go** and **PostgreSQL**. This project demonstrates practical backend engineering skills including API development, database design, transaction management, business logic implementation, containerization, and testing.

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Technology Stack](#technology-stack)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [API Documentation](#api-documentation)
- [Demo Accounts](#demo-accounts)
- [Testing](#testing)
- [Deployment](#deployment)
- [Known Limitations](#known-limitations)

## Overview

This Small ERP system addresses common business operations for small to medium enterprises:

- **Procurement** — Purchase order management, goods receipt processing, supplier management
- **Sales** — Sales order management, goods delivery, customer management
- **Inventory** — Real-time stock tracking, movement history, stock adjustments
- **Finance** — Double-entry accounting, accounts payable/receivable, payment processing
- **Reporting** — Sales, purchase, inventory, P&L, balance sheet, and cash flow reports

## Features

### Authentication & Authorization
- JWT-based authentication
- Role-Based Access Control (RBAC) with 3 default roles: Admin, Manager, Staff
- 32 granular permissions covering all modules
- Secure password hashing with bcrypt

### Master Data Management
- Products with categories and units of measurement
- Customers and suppliers with contact information
- Full CRUD operations with validation

### Procurement Module
- Purchase order creation with multiple line items
- Status workflow: Draft → Confirmed → Received / Cancelled
- Goods receipt processing with automatic:
  - Inventory updates
  - Accounts Payable creation
  - Double-entry journal entries (Debit Inventory, Credit AP)
- Duplicate receipt prevention

### Inventory Management
- Real-time stock balance tracking
- Stock movement history with reference tracking
- Stock adjustments with audit trail
- Insufficient stock validation

### Sales Module
- Sales order creation with multiple line items
- Status workflow: Draft → Confirmed → Delivered / Cancelled
- Delivery processing with automatic:
  - Inventory deduction (with stock validation)
  - Accounts Receivable creation
  - Double-entry journal entries (Debit AR, Credit Revenue + Debit COGS, Credit Inventory)

### Finance & Accounting
- Chart of Accounts (COA) management
- Manual journal entry creation with balance validation
- Accounts Payable/Receivable tracking
- Payment processing (AP & AR) with:
  - Partial and full payment support
  - Automatic status updates
  - Journal entry creation (Debit/Credit Cash)

### Reporting
- Sales report (by date range)
- Purchase report (by date range)
- Inventory valuation report
- Stock movement report
- Accounts Payable aging report
- Accounts Receivable aging report
- Cash flow report
- Profit & Loss statement
- Balance Sheet (follows accounting equation: Assets = Liabilities + Equity)

## Technology Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.22+ |
| Database | PostgreSQL 16 |
| Router | chi v5 |
| DB Driver | pgx/v5 |
| Auth | JWT (golang-jwt/jwt/v5) |
| Password | bcrypt |
| Migrations | golang-migrate |
| Logging | slog (stdlib) |
| Container | Docker + Docker Compose |

## Architecture

```
┌──────────────────────────────────────────────────┐
│                   HTTP Client                     │
└──────────────────────┬───────────────────────────┘
                       │
┌──────────────────────▼───────────────────────────┐
│              Middleware Layer                      │
│  (Logger → Recovery → Auth → RBAC)               │
└──────────────────────┬───────────────────────────┘
                       │
┌──────────────────────▼───────────────────────────┐
│               Handler Layer                       │
│  (Request validation, response formatting)        │
└──────────────────────┬───────────────────────────┘
                       │
┌──────────────────────▼───────────────────────────┐
│             Repository Layer                      │
│  (Database queries, transactions, business logic) │
└──────────────────────┬───────────────────────────┘
                       │
┌──────────────────────▼───────────────────────────┐
│            PostgreSQL Database                    │
│  (20+ tables, constraints, indexes)              │
└──────────────────────────────────────────────────┘
```

## Project Structure

```
├── cmd/server/main.go           # Application entry point
├── internal/
│   ├── config/                  # Environment configuration
│   ├── database/                # PostgreSQL connection pool
│   ├── middleware/              # Auth, RBAC, logging, recovery
│   ├── model/                   # Data models and request/response types
│   ├── handler/                 # HTTP request handlers
│   ├── repository/              # Database operations and business logic
│   └── router/                  # Route definitions
├── migrations/                  # Database schema and seed data
├── tests/smoke/                 # Smoke tests
├── scripts/                     # Utility scripts
├── Dockerfile                   # Multi-stage Go build
├── compose.yaml                 # Docker Compose orchestration
├── Makefile                     # Common development commands
└── .env.example                 # Environment variable template
```

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.22+ (for local development)

### Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/user/small-erp.git
   cd small-erp
   ```

2. **Start the application**
   ```bash
   docker compose up --build
   ```
   This will:
   - Start PostgreSQL 16
   - Run database migrations (schema + seed data)
   - Build and start the Go application

3. **Verify the application is running**
   ```bash
   curl http://localhost:8080/health
   ```

4. **Login to get a JWT token**
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username": "admin", "password": "admin123"}'
   ```

### Local Development

```bash
# Copy environment file
cp .env.example .env

# Start only the database
docker compose up -d db

# Run migrations
make migrate-up

# Build and run locally
make run
```

## API Documentation

### Base URL
```
http://localhost:8080/api/v1
```

### Response Format
All endpoints return a consistent JSON format:
```json
{
  "success": true,
  "message": "Description of the result",
  "data": { },
  "meta": { "page": 1, "per_page": 20, "total": 100 }
}
```

### Authentication
Include the JWT token in the `Authorization` header:
```
Authorization: Bearer <token>
```

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/api/v1/auth/login` | User login |
| **Users & Roles** | | |
| GET/POST | `/api/v1/users` | List/Create users |
| GET/PUT/DELETE | `/api/v1/users/{id}` | Get/Update/Delete user |
| GET/POST | `/api/v1/roles` | List/Create roles |
| GET/PUT/DELETE | `/api/v1/roles/{id}` | Get/Update/Delete role |
| **Master Data** | | |
| GET/POST | `/api/v1/products` | List/Create products |
| GET/PUT/DELETE | `/api/v1/products/{id}` | Get/Update/Delete product |
| GET/POST | `/api/v1/product-categories` | List/Create categories |
| GET/POST | `/api/v1/units` | List/Create units |
| GET/POST | `/api/v1/customers` | List/Create customers |
| GET/POST | `/api/v1/suppliers` | List/Create suppliers |
| **Procurement** | | |
| GET/POST | `/api/v1/purchase-orders` | List/Create POs |
| GET/PUT | `/api/v1/purchase-orders/{id}` | Get/Update PO |
| PATCH | `/api/v1/purchase-orders/{id}/status` | Update PO status |
| GET/POST | `/api/v1/goods-receipts` | List/Create goods receipts |
| **Inventory** | | |
| GET | `/api/v1/inventory` | Current stock balances |
| GET | `/api/v1/inventory/movements` | Stock movement history |
| POST | `/api/v1/inventory/adjustments` | Create stock adjustment |
| **Sales** | | |
| GET/POST | `/api/v1/sales-orders` | List/Create SOs |
| GET/PUT | `/api/v1/sales-orders/{id}` | Get/Update SO |
| PATCH | `/api/v1/sales-orders/{id}/status` | Update SO status |
| **Finance** | | |
| GET/POST | `/api/v1/accounts` | Chart of Accounts |
| GET/POST | `/api/v1/journal-entries` | Journal entries |
| GET/POST | `/api/v1/payments` | Payment processing |
| GET | `/api/v1/accounts-payable` | Accounts Payable |
| GET | `/api/v1/accounts-receivable` | Accounts Receivable |
| **Reports** | | |
| GET | `/api/v1/reports/sales` | Sales report |
| GET | `/api/v1/reports/purchases` | Purchase report |
| GET | `/api/v1/reports/inventory` | Inventory report |
| GET | `/api/v1/reports/stock-movements` | Stock movements |
| GET | `/api/v1/reports/accounts-payable` | AP report |
| GET | `/api/v1/reports/accounts-receivable` | AR report |
| GET | `/api/v1/reports/cash-flow` | Cash flow |
| GET | `/api/v1/reports/profit-loss` | Profit & Loss |
| GET | `/api/v1/reports/balance-sheet` | Balance Sheet |

### Example: Full Procurement Workflow

```bash
# 1. Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.data.token')

# 2. Create Purchase Order
curl -X POST http://localhost:8080/api/v1/purchase-orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "supplier_id": 1,
    "order_date": "2024-01-15",
    "items": [
      {"product_id": 1, "quantity": 10, "unit_price": 12000000}
    ]
  }'

# 3. Confirm PO
curl -X PATCH http://localhost:8080/api/v1/purchase-orders/1/status \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": "confirmed"}'

# 4. Receive Goods (creates inventory + AP + journal entry)
curl -X POST http://localhost:8080/api/v1/goods-receipts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "po_id": 1,
    "receipt_date": "2024-01-16",
    "items": [
      {"po_item_id": 1, "product_id": 1, "quantity_received": 10}
    ]
  }'

# 5. Pay supplier
curl -X POST http://localhost:8080/api/v1/payments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "payment_type": "payable",
    "reference_id": 1,
    "amount": 120000000,
    "payment_date": "2024-01-20",
    "payment_method": "bank_transfer"
  }'
```

## Demo Accounts

| Username | Password | Role | Access |
|----------|----------|------|--------|
| `admin` | `admin123` | Admin | Full access to all modules |
| `manager` | `manager123` | Manager | Read/write except user management |
| `staff` | `staff123` | Staff | Read access + create sales/purchases |

## Testing

### Run Smoke Tests
```bash
# Start the application first
docker compose up --build -d

# Wait for startup, then run smoke tests
go test ./tests/smoke/... -v -tags=smoke -count=1
```

### Smoke tests cover:
- Health check & database connectivity
- Authentication (login, token validation)
- Authorization (RBAC permission checks)
- Master data CRUD operations
- Full procurement workflow (PO → GR → Inventory → AP)
- Full sales workflow (SO → Delivery → Inventory → AR)
- Payment processing (AP and AR)
- All 9 report endpoints
- Insufficient stock rejection
- Unauthorized access denial

## Deployment

### Docker (Recommended)
```bash
docker compose up --build -d
```

### Environment Variables
See `.env.example` for all configuration options. Key variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_PORT` | `8080` | Application port |
| `DB_HOST` | `db` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `erp_user` | Database user |
| `DB_PASSWORD` | `erp_password` | Database password |
| `DB_NAME` | `erp_db` | Database name |
| `JWT_SECRET` | - | JWT signing secret (**change in production!**) |
| `JWT_EXPIRY_HOURS` | `24` | Token expiry in hours |

## Known Limitations

- **Single currency** — All amounts are in a single currency (no multi-currency support)
- **No partial goods receipt** — Goods receipt receives the full PO in one transaction
- **No product variants** — Products don't have variants (size, color, etc.)
- **Basic reporting** — Reports use simple SQL queries without caching
- **No audit trail** — No comprehensive audit log beyond stock movements
- **No file uploads** — No support for document attachments
- **No email notifications** — No email/notification system

### Future Improvements
- Add multi-warehouse support
- Implement partial goods receipts
- Add product images and variants
- Implement audit logging
- Add data export (CSV, PDF)
- Implement caching layer for reports
- Add WebSocket for real-time inventory updates
- Implement multi-currency support

## License

MIT License
