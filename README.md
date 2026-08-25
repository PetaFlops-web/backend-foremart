# Smart Commerce Backend API

**RESTful API documentation for the ForeMart smart commerce backend system.**

## 🏗️ Architecture Overview

This project uses a **Modular Monolith** architecture with Go (Fiber + GORM) running in Docker containers. The modular design separates concerns while maintaining operational simplicity without microservices complexity.

### Key Principles

1. **Data Isolation**: Each module manages its own database tables. Cross-module references use plain IDs (UUIDs), not database Foreign Keys.
2. **Client Interfaces**: Modules communicate via public interfaces (`<module>-client`) instead of direct table queries or JOINs across modules.
3. **Standardized Responses**: All API endpoints return structured JSON (`WebResponse[T]`) for consistent client parsing.

---

## 🚀 Quick Setup

### Prerequisites

- Docker & Docker Compose
- MySQL 8.0+ (provided via container)
- Git

### Installation Steps

1. **Clone Repository**
   ```bash
   git clone https://github.com/PetaFlops-web/backend-shop-smbk.git
   cd backend-shop-smbk
   ```

2. **Configuration**
   
   Copy environment template:
   ```bash
   cp .env.example .env
   ```
   
   Configure `config.json` to connect to services:
   - Database host should be `aic_mysql` for Docker Compose
   - ML service URL defaults to `http://host.docker.internal:8000`

3. **Run Services**
   ```bash
   docker compose up --build -d
   ```

Wait for MySQL and backend to initialize (check health status).

---

## 🔌 API Endpoints

### Base URL
```
http://localhost:8080
```

### Swagger Documentation
Interactive API docs at:
👉 **[http://localhost:8080/swagger/](http://localhost:8080/swagger/)**

---

## 📡 Authentication

Protected endpoints require JWT Bearer token in header:

```http
Authorization: Bearer <token>
```

Obtain tokens by POSTing to `/api/users/_login`:
```json
{
  "username": "admin",
  "password": "admin123"
}
```

---

## 🔮 Prediction APIs

### 1. Restock Prediction

Generates inventory replenishment recommendations based on ML-predicted daily sales.

#### Generate Predictions
```http
POST /api/restock-predictions/_generate
Authorization: Bearer ***
Content-Type: application/json
```

**Request Body:**
| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `store_id` | string | Yes | — | Store UUID |
| `product_id` | string | No | — | Product ID (omit = all products) |
| `forecast_date` | RFC3339 date | Yes | Tomorrow | Target prediction date |
| `history_days` | int | Yes | 30 | Minimum 30 days of history |

**Example (single product):**
```json
{
  "store_id": "4d967828-082d-4fe6-91a5-fd620b609d9e",
  "product_id": "prod_0000",
  "forecast_date": "2026-09-10T00:00:00Z",
  "history_days": 30
}
```

**Response:**
```json
{
  "data": {
    "generated_count": 1,
    "skipped_count": 0,
    "items": [{
      "id": "restock_xxx",
      "store_id": "xxx",
      "product_name": "Product Name",
      "unit": "kg",
      "forecast_date": "2026-09-10",
      "daily_sales": 10,
      "current_stock": 18,
      "recommended_restock_qty": 212,
      "created_at": 1787067747182
    }],
    "skipped": []
  },
  "message": "Successfully generated restock predictions",
  "success": true
}
```

**Skipped Products Reasons:**
- `riwayat_tidak_cukup`: Insufficient transaction history
- `kesalahan_ml`: ML service error  
- `restock_tidak_diperlukan`: Current stock adequate

---

### 2. Survival Prediction (Customer Retention)

Predicts when a customer will repurchase a specific product based on historical patterns.

```http
POST /api/predict-survival
Authorization: Bearer ***
Content-Type: application/json
```

**Request Body:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `store_id` | string | Yes | Store UUID |
| `customer_id` | int | Yes | Customer ID (numeric from DB) |
| `product_id` | string | Yes | Product UUID |

**Example:**
```json
{
  "store_id": "4d967828-082d-4fe6-91a5-fd620b609d9e",
  "customer_id": 1,
  "product_id": "prod_0000"
}
```

**Response:**
```json
{
  "data": {
    "customer_id": 1,
    "stock_code": "Product Name",
    "predicted_restock_date": "2026-08-25",
    "pred_days_left": 3,
    "pred_median_survival_days": 12.5,
    "days_since_last_buy": 9,
    "prob_buy_within_7d": 0.32,
    "prob_buy_within_14d": 0.61,
    "prob_buy_within_30d": 0.88,
    "partial_hazard": 1.24
  },
  "message": "Survival prediction successful",
  "success": true
}
```

**Key Fields:**
- `pred_days_left`: Days until predicted purchase (≤3 triggers notifications)
- `days_since_last_buy`: Days since customer's most recent purchase
- `prob_buy_within_*d`: Probability scores (0.0-1.0)

**Error Codes:**
| Code | Message | Cause |
|------|---------|-------|
| 400 | Invalid request format | Malformed JSON |
| 403 | Product not in your store | Mismatched store_id |
| 404 | No purchase history | Customer never bought this product |
| 500 | ML service unreachable | Connection timeout |

---

## 🔔 Notification System

### Fonnte WhatsApp Gateway Setup

Integration with [Fonnte](https://md.fonnte.com/) for WhatsApp delivery.

#### Configuration
Add to `config.json`:
```json
{
  "fonnte": {
    "token": "YOUR_FONNTE_API_TOKEN",
    "target": ""
  },
  "notification": {
    "schedule": "0 8 * * *"
  }
}
```

- Token obtained from Fonnte dashboard → Device menu
- Schedule: Cron expression (default: daily at 8 AM)
- Empty token = log-only mode (no actual SMS sent)

### Trigger Notifications Manually

```http
POST /api/notifications/_send
Authorization: Bearer ***
```

No body required. Triggers full notification pipeline.

**Response:**
```json
{
  "data": 5,
  "message": "Successfully executed notification sending",
  "success": true
}
```

### View Notification Logs

```http
GET /api/notifications/logs?store_id=<uuid>
Authorization: Bearer ***
```

**Sample Log Entry:**
```json
{
  "id": "notif_xxx",
  "store_id": "uuid",
  "customer_id": 1,
  "product_id": "prod_xxx",
  "channel": "whatsapp",
  "message": "Halo Budi, persediaan Gula Pasir 1 kg Anda...",
  "predicted_restock_date": "2026-08-25",
  "rule_triggered": "REPEAT_3X",
  "status": "sent",
  "period": "2026-08-22"
}
```

**Promotion Rules:**
- ≥5 purchases → 30% discount
- ≥3 purchases → 20% discount
- Otherwise → reminder only

---

## 💾 Transaction Mock Seeding

Development helper to generate realistic test data.

### Seed Transactions
```http
POST /api/transactions/_seed/mock
Authorization: Bearer ***
Content-Type: application/json
```

**Request Body:**
| Field | Type | Required | Values | Description |
|-------|------|----------|--------|-------------|
| `store_id` | string | Yes | UUID | Store identifier |
| `mode` | string | Yes | `restock` or `survival` | Mode type |

#### Restock Mode
- Creates backdated transactions over varied intervals (not daily)
- Applies random quantities for demand simulation  
- Automatically decrements stock levels
- Generates timestamp-based transaction IDs (`txn_{ms}_{random}`)

**Example Response:**
```json
{
  "success": true,
  "message": "Mock creation successful: 22 transactions | products: 2 | customers: 1",
  "data": {
    "productsAffected": ["prod_0000", "prod_0001"],
    "transactions_created": 22,
    "timeRange": {
      "startDate": "2026-05-15",
      "endDate": "2026-08-10"
    }
  }
}
```

#### Survival Mode  
- Backdates purchases strategically: 1, 45, and 95 days ago
- Ensures `days_since_last_buy` meets threshold logic
- Creates 3 transactions per customer for median calculation
- Triggers promotional offers based on purchase frequency

---

## 🎯 Standard Response Format

All endpoints follow unified response structure:

### Success Response
```json
{
  "data": { ... },
  "message": "Optional success message",
  "success": true
}
```

### Paginated Response
```json
{
  "data": [...],
  "message": "Success",
  "success": true,
  "paging": {
    "page": 1,
    "size": 10,
    "total_item": 25,
    "total_page": 3
  }
}
```

### Error Response
```json
{
  "data": null,
  "message": "Clear error description",
  "success": false,
  "statusCode": 400
}
```

---

## 🧪 Testing Workflows

### Complete Mock → Prediction → Notification Flow

1. **Create fresh environment** (if needed):
   ```bash
   curl -X POST http://localhost:8080/api/users/_login \\
     -H 'Content-Type: application/json' \\
     -d '{"username":"admin","password":"admin123"}'
   ```

2. **Seed mock data**:
   ```bash
   # Get bearer token from login response
   TOKEN="eyJhbG..."
   STORE_ID="4d967828-082d-4fe6-91a5-fd620b609d9e"
   
   # Create survival history
   curl -X POST http://localhost:8080/api/transactions/_seed/mock \\
     -H "Authorization: Bearer $TOKEN" \\
     -H 'Content-Type: application/json' \\
     -d "{\"store_id\":\"$STORE_ID\",\"mode\":\"survival\"}"
   ```

3. **Run prediction**:
   ```bash
   CUSTOMER_ID=1
   PRODUCT_ID="prod_0000"
   
   curl -X POST http://localhost:8080/api/predict-survival \\
     -H "Authorization: Bearer $TOKEN" \\
     -H 'Content-Type: application/json' \\
     -d "{\"store_id\":\"$STORE_ID\",\"customer_id\":$CUSTOMER_ID,\"product_id\":\"$PRODUCT_ID\"}"
   ```

4. **Trigger notifications**:
   ```bash
   curl -X POST http://localhost:8080/api/notifications/_send \\
     -H "Authorization: Bearer $TOKEN"
   ```

5. **Verify results**:
   ```bash
   curl -H "Authorization: Bearer $TOKEN" \\
     "http://localhost:8080/api/notifications/logs?store_id=$STORE_ID&page=1&size=5"
   ```

---

## 🛠️ Recent Enhancements

### Transaction ID Collision Prevention
- **Before**: Sequential IDs `txn_%04d` caused collisions during mass seeding
- **After**: Timestamp+random format `txn_{timestamp_ms}_{random4}` guarantees uniqueness

### Realistic Data Patterns
- Randomized transaction gaps instead of uniform daily spacing
- Variable quantity simulation for better ML training
- Automatic stock decrement synchronized with simulated sales

### Enhanced Debugging
- Comprehensive console logging in frontend components
- Detailed click/hover event tracking for bug identification
- State change visualization for complex flows

### Auto-Prediction Integration
- One-click workflow: Mock → ML Prediction → Notification
- Integrated UI buttons on `/prediksi` and `/prakiraan` pages
- Automatic selection of first available customer/product after seed

---

## 📚 Additional Resources

- **Design Documents**: [`backend/docs/design/`](./docs/design/)
- **Process Guides**: [`backend/docs/process/`](./docs/process/)
- **Product Requirements**: [`backend/docs/product/PRD.md`](./docs/product/PRD.md)
- **System Architecture**: [`SYSTEM_MAP.md`](./docs/design/SYSTEM_MAP.md)

---

## ⚠️ Important Notes

1. **ML Service Dependency**: Some endpoints require ML service running on `host.docker.internal:8000`
2. **MySQL Schema**: Must have `smart_commerce` database initialized before startup
3. **CORS Policy**: Only allows configured origins in production
4. **JWT Expiration**: Tokens expire after 24 hours by default
5. **Backup Strategy**: Always backup before bulk operations

---

© ForeMart Development Team - All Rights Reserved
