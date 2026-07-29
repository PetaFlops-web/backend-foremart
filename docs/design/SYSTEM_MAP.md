# System Map — Smart Commerce Backend

_Peta sistem backend Smart Commerce: kondisi aktual codebase, arsitektur, data, API, dan gap terhadap PRD/SDD._

| Field              | Detail                                                              |
| ------------------ | ------------------------------------------------------------------- |
| Nama Proyek        | Smart Commerce                                                      |
| Kompetisi          | AI Innovation Challenge (AIC)                                       |
| Dokumen Referensi  | PRD v0.1 (08 Jul 2026), SDD v0.1                                   |
| Tanggal Dibuat     | 12 Juli 2026                                                        |
| Status Codebase    | Awal — 1 dari 7 modul fully implemented, 1 stub, 5 belum dimulai   |
| Source of Truth     | Dokumen ini mengikuti **kondisi aktual codebase**, bukan aspirasi dokumen |

---

## 1. Project Overview

### 1.1 Deskripsi Singkat

Smart Commerce adalah aplikasi web untuk membantu pedagang UMKM (toko sembako) mengelola toko: mencatat transaksi via suara/foto, melihat laporan untung-rugi harian, mendapat prediksi restock otomatis, dan promosi otomatis berbasis rule. Backend dibangun dengan arsitektur **Modular Monolith** menggunakan Go (Golang).

### 1.2 Tech Stack Aktual (dari `go.mod` & config)

| Komponen          | Teknologi                              | Versi (go.mod)      | Catatan                                    |
| ----------------- | -------------------------------------- | ------------------- | ------------------------------------------ |
| Bahasa            | Go (Golang)                            | 1.25.3              | SDD menyebut 1.22 — codebase lebih baru   |
| HTTP Framework    | Fiber v2                               | v2.52.14            | —                                          |
| ORM               | GORM                                   | v1.30.0             | —                                          |
| Database Driver   | MySQL only                             | v1.6.0              | SDD menyebut MySQL/PostgreSQL configurable |
| Validasi Input    | go-playground/validator                | v10.30.3            | —                                          |
| Konfigurasi       | Viper                                  | v1.21.0             | config.json + env vars                     |
| Logging           | Logrus                                 | v1.9.4              | JSON formatter                             |
| JWT               | golang-jwt/jwt v5                      | v5.3.1              | —                                          |
| Password Hashing  | golang.org/x/crypto (bcrypt)           | v0.52.0             | Tidak disebut di SDD                       |
| API Docs          | fiber-swagger                          | v1.3.0              | Tidak disebut di SDD tapi dipakai          |
| Containerization  | Docker + docker-compose                | —                   | Multi-stage build                          |
| Scheduler/Cron    | ❌ Belum ada                           | —                   | SDD: robfig/cron                           |

### 1.3 Status Keseluruhan

```
Modul Implemented : 2/7 (auth, prod)  ████████░░░░░░  ~28%
Modul Stub        : 0/7               ░░░░░░░░░░░░░░
Modul Not Started : 5/7               ░░░░░░░░░░░░░░
Tabel DB Aktual   : 2/7 (users, prod)
Endpoint Aktual   : 8/~14
```

---

## 2. Architecture Overview

### 2.1 Pola Arsitektur

**Modular Monolith** — satu deployment unit, kode dipecah menjadi modul-modul terisolasi:

- **Data Isolation**: setiap modul punya tabel sendiri. Referensi lintas modul berupa plain ID, bukan FK database.
- **Inter-module Communication**: lewat `*-client` interface, bukan query/JOIN langsung ke tabel modul lain.
- **Independent Migration**: setiap modul menjalankan `AutoMigrate` untuk tabelnya sendiri.
- **Module Contract**: semua modul mengimplementasikan interface `module.Module` (`Migrate()` + `RegisterRoutes()`).

> Ref SDD: Section 2.1–2.4

### 2.2 Module Map (Status Implementasi)

```mermaid
graph TB
    subgraph "Entrypoint"
        MAIN["cmd/web/main.go<br/>Fiber HTTP Server"]
    end

    subgraph "Modules — Implemented ✅"
        AUTH["auth<br/>✅ Fully Implemented<br/>Tabel: users"]
        PRODUCT["product<br/>✅ Fully Implemented<br/>Tabel: products"]
    end

    subgraph "Modules — Stub 🟡"
    end

    subgraph "Modules — Not Started ❌"
        STORE["store ❌"]
        TRANSACTION["transaction ❌"]
        REPORT["report ❌"]
        RESTOCK["restock ❌"]
        PROMOTION["promotion ❌"]
    end

    subgraph "Planned — Not Started ❌"
        JOBS["internal/jobs/ ❌<br/>restock_job, promotion_job"]
        MLCLIENT["internal/pkg/mlclient/ ❌<br/>ML Integration"]
    end

    subgraph "Shared Infrastructure ✅"
        CONFIG["shared/config<br/>Viper, GORM, Fiber, Logrus, Validator"]
        MW["shared/middleware<br/>JWT Auth"]
        REPO["shared/repository<br/>Generic Repository T"]
        RESP["shared/response<br/>WebResponse T"]
        UTILS["shared/utils<br/>ID Generator"]
        MOD["shared/module<br/>Module Interface"]
    end

    MAIN --> AUTH
    MAIN --> PRODUCT
    MAIN -.-> STORE
    MAIN -.-> TRANSACTION
    MAIN -.-> REPORT

    AUTH --> CONFIG
    AUTH --> MW
    AUTH --> REPO
    AUTH --> RESP
    AUTH --> UTILS

    TRANSACTION -.-> MLCLIENT
    RESTOCK -.-> JOBS
    PROMOTION -.-> JOBS

    style AUTH fill:#2d6a4f,color:#fff
    style PRODUCT fill:#2d6a4f,color:#fff
    style STORE fill:#e76f51,color:#fff
    style TRANSACTION fill:#e76f51,color:#fff
    style REPORT fill:#e76f51,color:#fff
    style RESTOCK fill:#e76f51,color:#fff
    style PROMOTION fill:#e76f51,color:#fff
    style JOBS fill:#e76f51,color:#fff
    style MLCLIENT fill:#e76f51,color:#fff
    style CONFIG fill:#264653,color:#fff
    style MW fill:#264653,color:#fff
    style REPO fill:#264653,color:#fff
    style RESP fill:#264653,color:#fff
    style UTILS fill:#264653,color:#fff
    style MOD fill:#264653,color:#fff
```

**Legenda:** ✅ Implemented | 🟡 Stub | ❌ Not Started

### 2.3 Entrypoint & Wiring

File: `cmd/web/main.go`

```
main()
  ├─ config.NewViper()          → load config.json
  ├─ config.NewLogger()         → Logrus (JSON)
  ├─ config.NewDatabase()       → GORM + MySQL
  ├─ config.NewValidator()      → go-playground/validator
  ├─ config.NewFiber()          → Fiber app + Swagger route
  ├─ middleware.AuthMiddleware() → JWT middleware
  ├─ auth.New()                 → satu-satunya modul aktif
  ├─ modules[].Migrate()        → GORM AutoMigrate
  ├─ modules[].RegisterRoutes() → register ke /api group
  └─ app.Listen(:8080)
```

> **Gap**: SDD menyebut `cmd/api/main.go` — codebase menggunakan `cmd/web/main.go`. Keputusan: **`cmd/web/` adalah standar** (OQ-4).

---

## 3. Module Dependency Diagram

Diagram berikut menunjukkan dependency antar modul **sesuai rencana SDD**, di-overlay dengan status implementasi aktual.

```mermaid
graph LR
    subgraph "✅ Implemented"
        AUTH["auth"]
        AUTH_C["auth-client"]
        PRODUCT["product"]
        PRODUCT_C["product-client"]
    end

    subgraph "🟡 Stub"
    end

    subgraph "❌ Not Started"
        STORE["store"]
        STORE_C["store-client"]
        TXN["transaction"]
        TXN_C["transaction-client"]
        REPORT["report"]
        RESTOCK["restock"]
        PROMO["promotion"]
        MLC["mlclient"]
    end

    AUTH --> AUTH_C
    PRODUCT --> PRODUCT_C

    %% Planned dependencies from SDD
    STORE -.->|uses| AUTH_C
    PRODUCT -.->|uses| STORE_C
    TXN -.->|uses| STORE_C
    TXN -.->|uses| PRODUCT_C
    TXN -.->|uses| MLC
    REPORT -.->|uses| TXN_C
    REPORT -.->|uses| PRODUCT_C
    RESTOCK -.->|uses| TXN_C
    RESTOCK -.->|uses| PRODUCT_C
    PROMO -.->|uses| TXN_C
    PROMO -.->|uses| PRODUCT_C

    style AUTH fill:#2d6a4f,color:#fff
    style AUTH_C fill:#2d6a4f,color:#fff
    style PRODUCT fill:#2d6a4f,color:#fff
    style PRODUCT_C fill:#2d6a4f,color:#fff
    style STORE fill:#e76f51,color:#fff
    style STORE_C fill:#e76f51,color:#fff
    style TXN fill:#e76f51,color:#fff
    style TXN_C fill:#e76f51,color:#fff
    style REPORT fill:#e76f51,color:#fff
    style RESTOCK fill:#e76f51,color:#fff
    style PROMO fill:#e76f51,color:#fff
    style MLC fill:#e76f51,color:#fff
```

### Client Interface Inventory

| Client Interface   | Method (SDD)                                                           | Status     |
| ------------------ | ---------------------------------------------------------------------- | ---------- |
| `auth-client`      | `GetUserByID(ctx, userID) → *UserDTO`                                  | ✅ Implemented |
| `product-client`   | `GetByID`, `ListByStoreID`, `DecrementStock`, `Search`                 | ✅ Implemented |
| `store-client`     | `GetStoreByOwnerID`, `GetStoreByID`                                    | ❌ Belum ada |
| `transaction-client` | `CreateTransaction`, `ListByStoreAndDate`, `ListItemsByStoreAndDateRange` | ❌ Belum ada |

> Ref SDD: Section 7

---

## 4. Project Structure (Aktual)

```
Backend-AIC/
├── cmd/
│   └── web/
│       └── main.go                              # Entrypoint (SDD: cmd/api/)
├── config.json                                  # Konfigurasi utama (Viper)
├── config.example.json                          # Template config
├── .env                                         # Env vars untuk docker-compose
├── internal/
│   ├── module/                                  # ⚠️ Singular — standar: modules/ (OQ-5)
│   │   ├── auth/                                # ✅ FULLY IMPLEMENTED
│   │   │   ├── module.go                        # Wiring: New(), Migrate(), Client()
│   │   │   ├── route.go                         # RegisterRoutes: /api/users/*
│   │   │   ├── client_impl.go                   # Implements auth-client.Client
│   │   │   └── src/
│   │   │       ├── controller/
│   │   │       │   └── auth_controller.go       # Register, Login, Current
│   │   │       ├── entity/
│   │   │       │   └── user_entity.go           # GORM struct → tabel users
│   │   │       ├── model/
│   │   │       │   ├── auth_request.go          # RegisterRequest, LoginRequest
│   │   │       │   ├── auth_response.go         # UserResponse, AuthResponse
│   │   │       │   └── converter/
│   │   │       │       └── auth_converter.go    # UserToResponse
│   │   │       ├── repository/
│   │   │       │   └── user_repository.go       # FindByUsername, CountByUsername
│   │   │       └── usecase/
│   │   │           └── auth_usecase.go          # Register, Login, Current logic
│   │   ├── auth-client/
│   │   │   └── client.go                        # ✅ Interface: Client{GetUserByID}
│   │   ├── product/                             # ✅ FULLY IMPLEMENTED
│   │   │   ├── module.go                        # Wiring module
│   │   │   ├── route.go                         # Register routes
│   │   │   ├── client_impl.go                   # Implements product-client.Client
│   │   │   └── src/
│   │   │       ├── controller/
│   │   │       │   └── product_controller.go    # Create, Get, List, Update, Delete
│   │   │       ├── entity/
│   │   │       │   └── product_entity.go        # GORM struct → tabel products
│   │   │       ├── model/
│   │   │       │   ├── product_request.go       # Request DTOs
│   │   │       │   ├── product_response.go      # Response DTOs
│   │   │       │   ├── product_params.go        # Query params DTOs
│   │   │       │   └── converter/
│   │   │       │       └── product_converter.go # ProductToResponse
│   │   │       ├── repository/
│   │   │       │   └── product_repository.go    # Custom queries & Search
│   │   │       └── usecase/
│   │   │           └── product_usecase.go       # Business logic
│   │   └── product-client/
│   │       └── client.go                        # ✅ Interface: Client
│   └── shared/
│       ├── config/
│       │   ├── viper.go                         # ✅ Load config.json
│       │   ├── gorm.go                          # ✅ MySQL connection + pool
│       │   ├── fiber.go                         # ✅ Fiber app + Swagger + ErrorHandler
│       │   ├── logrus.go                        # ✅ JSON structured logging
│       │   └── validator.go                     # ✅ go-playground/validator
│       ├── middleware/
│       │   ├── auth.go                          # ✅ Auth struct{ID}
│       │   ├── auth_middleware.go               # ✅ JWT Bearer validation
│       │   └── jwt.go                           # ✅ GenerateJWT (HS256, 72h expiry)
│       ├── module/
│       │   └── module.go                        # ✅ Module interface
│       ├── repository/
│       │   └── repository.go                    # ✅ Generic Repository[T]
│       ├── response/
│       │   └── response.go                      # ✅ WebResponse[T], PageMetadata, ApiErrorResponse
│       └── utils/
│           └── generate_user_id.go              # ✅ usr_{username}_{random4digit}
├── db/
│   └── migrations/
│       ├── 20260709160912_create_table_users.up.sql    # ✅ CREATE TABLE users
│       ├── 20260709160912_create_table_users.down.sql  # ✅ DROP TABLE users
│       ├── 20260715230900_create_table_products.up.sql # ✅ CREATE TABLE products
│       └── 20260715230900_create_table_products.down.sql # ✅ DROP TABLE products
├── docs/
│   ├── PRD.md
│   ├── SDD.md
│   └── SYSTEM_MAP.md                            # ← dokumen ini
├── Dockerfile                                   # ✅ Multi-stage build
├── docker-compose.yml                           # ✅ api + mysql
├── go.mod
├── go.sum
└── .gitignore
```

### Perbandingan Struktur: SDD vs Codebase

| Aspek | SDD (Rencana) | Codebase (Aktual) | Keputusan |
| --- | --- | --- | --- |
| Entrypoint | `cmd/api/main.go` | `cmd/web/main.go` | Codebase ✓ (OQ-4) |
| Module path | `internal/modules/` (plural) | `internal/module/` (singular) | Rename ke `modules/` (OQ-5) |
| Layer files | Flat: `handler.go`, `service.go`, `model.go` | Nested: `src/controller/`, `src/usecase/`, `src/entity/`, `src/model/converter/` | Codebase ✓ (OQ-6) |
| Jobs folder | `internal/jobs/` | ❌ Belum ada | Planned |
| ML client | `internal/pkg/mlclient/` | ❌ Belum ada | Planned |
| Extra: `src/model/params` | Tidak disebut | Ada di product (kosong) | Tambahan codebase |
| Extra: `converter/` | Tidak disebut | Ada, dipakai di auth | Tambahan codebase |

---

## 5. Module Inventory

| Modul | Status | Tabel DB | Endpoint Count | FR-ID Terkait | Ref SDD |
| --- | --- | --- | --- | --- | --- |
| `auth` | ✅ Implemented | `users` | 3 aktif | FR-01 (partial) | §3, §6.1, §8.1 |
| `store` | ❌ Not Started | `stores` (planned) | 0 / 2 planned | FR-02 | §3, §6.2, §8.2 |
| `product` | ✅ Implemented | `products` | 5 aktif | FR-03, FR-04, FR-05 | §3, §6.3, §8.3 |
| `transaction` | 🟡 Entity Only | `transactions`, `transaction_items` | 0 / 4 planned | FR-05–FR-15 | §3, §6.4–6.5, §8.4 |
| `report` | ❌ Not Started | — (query via client) | 0 / 1 planned | FR-16–FR-20 | §3, §8.5 |
| `restock` | ❌ Not Started | `restock_predictions` (planned) | 0 / 1 planned | FR-21–FR-22 | §3, §6.6, §8.6 |
| `promotion` | ❌ Not Started | `promotion_logs` (planned) | 0 / 1 planned | FR-23–FR-25 | §3, §6.7, §8.6 |

---

## 6. API Endpoint Inventory

### 6.1 Endpoint Aktual (Implemented)

| Method | Path | Modul | Fungsi | Auth | Status | FR-ID |
| --- | --- | --- | --- | --- | --- | --- |
| `POST` | `/api/users` | auth | Register (username/password) | — | ✅ Implemented | FR-01 (modified) |
| `POST` | `/api/users/_login` | auth | Login (username/password → JWT) | — | ✅ Implemented | FR-01 (modified) |
| `GET` | `/api/users/_current` | auth | Profil user yang sedang login | JWT | ✅ Implemented | FR-01 |
| `GET` | `/swagger/*` | shared/config | Swagger UI | — | ✅ Implemented | — |
| `POST` | `/api/products` | product | Tambah produk | JWT | ✅ Implemented | FR-03 |
| `GET` | `/api/products` | product | List produk (paginated) | JWT | ✅ Implemented | FR-03 |
| `GET` | `/api/products/:id` | product | Detail produk | JWT | ✅ Implemented | — |
| `PUT` | `/api/products/:id` | product | Update produk | JWT | ✅ Implemented | — |
| `DELETE` | `/api/products/:id` | product | Hapus produk | JWT | ✅ Implemented | — |

### 6.2 Endpoint Planned (Belum Ada di Codebase)

| Method | Path | Modul | Fungsi | Status | FR-ID | Ref SDD |
| --- | --- | --- | --- | --- | --- | --- |
| `POST` | `/api/stores` | store | Buat toko baru | ❌ | FR-02 | §8.2 |
| `GET` | `/api/stores` | store | Ambil toko milik user | ❌ | FR-02 | §8.2 |

| `POST` | `/api/transactions/extract/voice` | transaction | Upload audio → ML → preview | ❌ | FR-06–FR-10 | §8.4 |
| `POST` | `/api/transactions/extract/photo` | transaction | Upload foto → ML → preview | ❌ | FR-12–FR-15 | §8.4 |
| `POST` | `/api/transactions` | transaction | Konfirmasi & simpan transaksi | ❌ | FR-11, FR-05 | §8.4 |
| `GET` | `/api/transactions` | transaction | List riwayat transaksi | ❌ | — | §8.4 |
| `GET` | `/api/reports/daily` | report | Laporan harian | ❌ | FR-16–FR-20 | §8.5 |
| `GET` | `/api/restock-predictions` | restock | List prediksi restock | ❌ | FR-21–FR-22 | §8.6 |
| `GET` | `/api/promotions` | promotion | List promotion logs | ❌ | FR-23–FR-24 | §8.6 |

---

## 7. Data Schema

### 7.1 Tabel `users` — ✅ Implemented (Modul `auth`)

**Source**: GORM entity (`user_entity.go`) + SQL migration (`20260709160912_create_table_users.up.sql`)

| Field | Tipe (GORM Entity) | Tipe (SQL Migration) | GORM Tags | Catatan |
| --- | --- | --- | --- | --- |
| `id` | `string` | `VARCHAR(36)` | `primaryKey` | Custom format: `usr_{username}_{random4digit}` |
| `username` | `string` | `VARCHAR(50)` ⚠️ | `varchar(100);not null;unique` | ⚠️ GORM=100, migration=50 (OQ-9: migration benar) |
| `email` | `string` | `VARCHAR(100)` ⚠️ | `varchar(255);not null;unique` | ⚠️ GORM=255, migration=100 (OQ-9: migration benar) |
| `password` | `string` | `VARCHAR(255)` | `not null` | Bcrypt hash. ⚠️ Tidak ada di SDD (SDD rencana OAuth) |
| `created_at` | `int64` | `BIGINT` | `autoCreateTime:milli` | Milli-epoch |
| `updated_at` | `int64` | `BIGINT` | `autoUpdateTime:milli` | Milli-epoch |

**Index**: `UNIQUE idx_users_username (username)`, `UNIQUE idx_users_email (email)`

> **Konvensi Timestamp yang Disepakati (OQ-3)**:
> - Entity (GORM): `int64` dengan tag `autoCreateTime:milli` / `autoUpdateTime:milli`
> - Response DTO (Go): `*time.Time`
> - DB Column: `BIGINT`

> **Known Issue (OQ-9)**: GORM entity mendefinisikan `username varchar(100)` dan `email varchar(255)`, tapi SQL migration mendefinisikan `varchar(50)` dan `varchar(100)`. **Keputusan: SQL migration yang benar** — GORM entity perlu diupdate agar cocok.

### 7.1.2 Tabel `products` — ✅ Implemented (Modul `product`)

**Source**: GORM entity (`product_entity.go`) + SQL migration (`20260715230900_create_table_products.up.sql`)

| Field | Tipe (GORM Entity) | Tipe (SQL Migration) | GORM Tags | Catatan |
| --- | --- | --- | --- | --- |
| `id` | `string` | `VARCHAR(36)` | `primaryKey` | Custom format |
| `store_id` | `string` | `VARCHAR(36)` | `not null` | Reference ke stores.id (plain ID) |
| `product_name` | `string` | `VARCHAR(255)` | `not null` | Nama produk |
| `cost_price` | `int64` | `BIGINT` | `not null;default:0` | Harga modal (rupiah, tanpa desimal) |
| `selling_price` | `int64` | `BIGINT` | `not null;default:0` | Harga jual |
| `stock` | `int` | `INT` | `not null;default:0` | Stok saat ini |
| `unit` | `string` | `VARCHAR(50)` | `not null` | Satuan (kg, pcs, dll) |
| `created_at` | `int64` | `BIGINT` | `autoCreateTime:milli` | Milli-epoch |
| `updated_at` | `int64` | `BIGINT` | `autoUpdateTime:milli` | Milli-epoch |

### 7.2 Tabel yang Direncanakan (Belum Ada di Codebase)

Berikut schema yang direncanakan di SDD, belum ada kode/migration-nya:

#### Tabel `stores` (Modul `store` — ❌ Not Started)

| Field | Tipe | Keterangan |
| --- | --- | --- |
| `id` | string (custom format) | Primary key |
| `owner_id` | string | Reference ke users.id (plain ID, bukan FK) |
| `store_name` | string | Nama toko |
| `created_at` | BIGINT | Milli-epoch |
| `updated_at` | BIGINT | Milli-epoch |

> Ref SDD §6.2


#### Tabel `transactions` (Modul `transaction` — 🟡 Entity & Migration)

| Field | Tipe | Keterangan |
| --- | --- | --- |
| `id` | string (custom format) | Primary key |
| `store_id` | string | Reference ke stores.id (plain ID) |
| `transaction_date` | date | Tanggal transaksi |
| `source` | string (enum) | `voice` \| `photo` |
| `created_at` | BIGINT | Milli-epoch |

> Ref SDD §6.4

#### Tabel `transaction_items` (Modul `transaction` — 🟡 Entity & Migration)

| Field | Tipe | Keterangan |
| --- | --- | --- |
| `id` | string (custom format) | Primary key |
| `transaction_id` | string | FK ke transactions.id (intra-modul) |
| `product_id` | string | Reference ke products.id (lintas modul, plain ID) |
| `product_name_snapshot` | string | Snapshot nama produk saat transaksi |
| `qty` | int | Jumlah unit terjual |
| `cost_price_snapshot` | bigint | Snapshot harga modal saat transaksi |
| `selling_price_snapshot` | bigint | Snapshot harga jual saat transaksi |

> Ref SDD §6.5

#### Tabel `restock_predictions` (Modul `restock` — ❌ Not Started)

| Field | Tipe | Keterangan |
| --- | --- | --- |
| `id` | string (custom format) | Primary key |
| `store_id` | string | Reference ke stores.id |
| `product_id` | string | Reference ke products.id |
| `predicted_restock_date` | date | Estimasi tanggal restock |
| `predicted_qty` | int | Estimasi kuantitas restock |
| `avg_daily_sold` | decimal | Rata-rata penjualan harian (basis prediksi) |
| `created_at` | BIGINT | Milli-epoch |

> Ref SDD §6.6

#### Tabel `promotion_logs` (Modul `promotion` — ❌ Not Started)

| Field | Tipe | Keterangan |
| --- | --- | --- |
| `id` | string (custom format) | Primary key |
| `store_id` | string | Reference ke stores.id |
| `product_id` | string | Reference ke products.id |
| `rule_triggered` | string | Nama/kode rule, mis. `QTY_MONTHLY_THRESHOLD` |
| `period` | string | Periode evaluasi, mis. `2026-07` |
| `created_at` | BIGINT | Milli-epoch |

> Ref SDD §6.7

### 7.3 Entity-Relationship Diagram (Rencana Penuh)

```mermaid
erDiagram
    users {
        varchar id PK "usr_{username}_{rand}"
        varchar username UK
        varchar email UK
        varchar password
        bigint created_at
        bigint updated_at
    }

    stores {
        varchar id PK
        varchar owner_id "→ users.id (plain ref)"
        varchar store_name
        bigint created_at
        bigint updated_at
    }

    products {
        varchar id PK
        varchar store_id "→ stores.id (plain ref)"
        varchar product_name
        bigint cost_price
        bigint selling_price
        int stock
        varchar unit
        bigint created_at
        bigint updated_at
    }

    transactions {
        varchar id PK
        varchar store_id "→ stores.id (plain ref)"
        date transaction_date
        varchar source "voice | photo"
        bigint created_at
    }

    transaction_items {
        varchar id PK
        varchar transaction_id FK "→ transactions.id"
        varchar product_id "→ products.id (plain ref)"
        varchar product_name_snapshot
        int qty
        bigint cost_price_snapshot
        bigint selling_price_snapshot
    }

    restock_predictions {
        varchar id PK
        varchar store_id "→ stores.id"
        varchar product_id "→ products.id"
        date predicted_restock_date
        int predicted_qty
        decimal avg_daily_sold
        bigint created_at
    }

    promotion_logs {
        varchar id PK
        varchar store_id "→ stores.id"
        varchar product_id "→ products.id"
        varchar rule_triggered
        varchar period
        bigint created_at
    }

    users ||--o{ stores : "owns"
    stores ||--o{ products : "has"
    stores ||--o{ transactions : "has"
    transactions ||--o{ transaction_items : "contains"
    products ||--o{ transaction_items : "referenced by"
    products ||--o{ restock_predictions : "predicted for"
    products ||--o{ promotion_logs : "promoted"
    stores ||--o{ restock_predictions : "belongs to"
    stores ||--o{ promotion_logs : "belongs to"
```

> **Catatan**: Hanya tabel `users` yang sudah ada di codebase. Seluruh tabel lain adalah rencana dari SDD dan belum diimplementasikan. Relasi lintas modul menggunakan plain ID reference (bukan FK database), sesuai prinsip data isolation modular monolith.

---

## 8. Data Flow Diagrams

### 8.1 Flow yang SUDAH Berjalan di Codebase

#### 8.1.1 Register User

```mermaid
sequenceDiagram
    participant FE as Frontend
    participant F as Fiber Router
    participant C as AuthController
    participant U as AuthUseCase
    participant R as UserRepository
    participant DB as MySQL

    FE->>F: POST /api/users {username, email, password}
    F->>C: Register(ctx)
    C->>C: BodyParser → RegisterRequest
    C->>U: Register(ctx, request)
    U->>U: Validate(request)
    U->>DB: BEGIN TRANSACTION
    U->>R: CountByUsername(tx, username)
    R->>DB: SELECT COUNT(*) FROM users WHERE id = ? ⚠️ bug
    DB-->>R: count
    alt username exists
        U-->>FE: 409 Conflict "Username already taken"
    end
    U->>U: bcrypt.GenerateFromPassword(password)
    U->>U: utils.GenerateUserId(username) → "usr_{name}_{rand}"
    U->>R: Create(tx, user)
    R->>DB: INSERT INTO users
    U->>U: middleware.GenerateJWT(secret, id, username)
    U->>DB: COMMIT
    U-->>C: AuthResponse{Token, User}
    C-->>FE: 200 WebResponse{data: {token, user}, success: true}
```

> ⚠️ **Known Bug**: `CountByUsername` query menggunakan `WHERE id = ?` alih-alih `WHERE username = ?` — lihat Section 13.1.

#### 8.1.2 Login User

```mermaid
sequenceDiagram
    participant FE as Frontend
    participant C as AuthController
    participant U as AuthUseCase
    participant R as UserRepository
    participant DB as MySQL

    FE->>C: POST /api/users/_login {username, password}
    C->>U: Login(ctx, request)
    U->>U: Validate(request)
    U->>DB: BEGIN TRANSACTION
    U->>R: FindByUsername(tx, username)
    R->>DB: SELECT * FROM users WHERE username = ?
    DB-->>R: user record
    U->>U: bcrypt.CompareHashAndPassword(hash, password)
    alt password mismatch
        U-->>FE: 404 "Username atau password anda salah"
    end
    U->>U: middleware.GenerateJWT(secret, id, username)
    U->>DB: COMMIT
    U-->>C: AuthResponse{Token, User}
    C-->>FE: 200 WebResponse{data: {token, user}, success: true}
```

#### 8.1.3 Get Current User (Protected)

```mermaid
sequenceDiagram
    participant FE as Frontend
    participant MW as JWT Middleware
    participant C as AuthController
    participant U as AuthUseCase
    participant R as UserRepository
    participant DB as MySQL

    FE->>MW: GET /api/users/_current [Authorization: Bearer {token}]
    MW->>MW: Parse & validate JWT (HS256)
    MW->>MW: Extract claims.id → set ctx.Locals("auth")
    MW->>C: Current(ctx)
    C->>C: middleware.GetUser(ctx) → Auth{ID}
    C->>U: Current(ctx, auth.ID)
    U->>R: FindById(tx, user, id)
    R->>DB: SELECT * FROM users WHERE id = ?
    DB-->>R: user record
    U->>U: converter.UserToResponse(user)
    U-->>C: UserResponse
    C-->>FE: 200 WebResponse{data: user, success: true}
```

### 8.2 Flow yang DIRENCANAKAN (Belum Ada di Codebase)

#### 8.2.1 Pencatatan Transaksi via Suara (Planned)

```mermaid
sequenceDiagram
    participant FE as Frontend
    participant TXN as Transaction Module
    participant ML as ML Client ❌
    participant PC as Product Client ❌
    participant DB as MySQL

    Note over FE,DB: ❌ SELURUH FLOW INI BELUM DIIMPLEMENTASIKAN

    FE->>TXN: POST /api/transactions/extract/voice [audio file]
    TXN->>ML: TranscribeAndExtract(audio)
    ML-->>TXN: [{raw_text, qty, harga}, ...]
    loop Setiap item
        TXN->>PC: Search(storeID, raw_text)
        PC-->>TXN: matched product + confidence
    end
    TXN-->>FE: Preview [{raw_text, matched_product, qty, harga, confidence}]

    Note over FE: Pedagang edit & konfirmasi preview

    FE->>TXN: POST /api/transactions {store_id, items: [...]}
    TXN->>DB: BEGIN TRANSACTION
    loop Setiap item
        TXN->>PC: GetByID(product_id)
        TXN->>PC: DecrementStock(product_id, qty)
    end
    TXN->>DB: INSERT transactions + transaction_items (snapshot harga)
    TXN->>DB: COMMIT
    TXN-->>FE: 200 {transaction_id}
```

> Ref PRD: FR-06–FR-11, SDD §9.1–9.2

#### 8.2.2 Laporan Harian (Planned)

```mermaid
sequenceDiagram
    participant FE as Frontend
    participant RPT as Report Module ❌
    participant TC as Transaction Client ❌
    participant PC as Product Client ❌

    Note over FE,PC: ❌ BELUM DIIMPLEMENTASIKAN

    FE->>RPT: GET /api/reports/daily?date=2026-07-12
    RPT->>TC: ListItemsByStoreAndDate(storeID, date)
    TC-->>RPT: transaction_items[]
    RPT->>RPT: SUM(qty * selling_price_snapshot) → total_omset
    RPT->>RPT: SUM(qty * (selling - cost)) → total_untung
    RPT->>RPT: COUNT(DISTINCT transaction_id) → jumlah_transaksi
    RPT->>RPT: GROUP BY product_id, SUM(qty) DESC → produk_terlaris
    RPT->>PC: ListByStoreID(storeID) → sisa stok
    RPT-->>FE: 200 {total_omset, total_untung, jumlah_transaksi, produk_terlaris, sisa_stok}
```

> Ref PRD: FR-16–FR-20, SDD §9.3

#### 8.2.3 Cron Job — Prediksi Restock (Planned)

```mermaid
sequenceDiagram
    participant CRON as Scheduler ❌
    participant RST as Restock Module ❌
    participant TC as Transaction Client ❌
    participant PC as Product Client ❌
    participant DB as MySQL

    Note over CRON,DB: ❌ BELUM DIIMPLEMENTASIKAN

    CRON->>RST: Run (setiap malam, mis. 23:00)
    loop Setiap store
        RST->>PC: ListByStoreID(storeID)
        loop Setiap product
            RST->>TC: ListItemsByProduct(productID, 30 hari)
            RST->>RST: avg_daily = SUM(qty) / 30
            RST->>RST: days_until_stockout = stock / avg_daily
            alt days_until_stockout <= THRESHOLD (3 hari)
                RST->>DB: INSERT restock_predictions
            end
        end
    end
```

> Ref PRD: FR-21–FR-22, SDD §9.4

#### 8.2.4 Cron Job — Promosi Otomatis (Planned)

```mermaid
sequenceDiagram
    participant CRON as Scheduler ❌
    participant PROMO as Promotion Module ❌
    participant TC as Transaction Client ❌
    participant DB as MySQL

    Note over CRON,DB: ❌ BELUM DIIMPLEMENTASIKAN

    CRON->>PROMO: Run (setiap malam, terpisah dari restock)
    loop Setiap store → setiap product
        PROMO->>TC: SumQtyByProductInMonth(productID, month)
        alt qty >= RULE_THRESHOLD_QTY
            PROMO->>DB: CHECK existing (product_id, period)
            alt belum ada log
                PROMO->>DB: INSERT promotion_logs
            end
        end
    end
```

> Ref PRD: FR-23–FR-24, SDD §9.5

---

## 9. Shared Infrastructure

### 9.1 Konfigurasi (`config.json`)

```json
{
  "app":      { "name": "aic" },
  "web":      { "prefork": false, "port": 8080 },
  "log":      { "level": 6 },
  "database": { "username", "password", "host", "port", "name", "pool": { "idle", "max", "lifetime" } },
  "jwt":      { "secret": "..." },
  "group":    { "id": "aic" }
}
```

- Env vars override: `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_PORT`
- `group.id` — config legacy, tidak dipakai oleh kode manapun saat ini. Dibiarkan.

### 9.2 Module Interface

```go
// internal/shared/module/module.go
type Module interface {
    RegisterRoutes(router fiber.Router, authMiddleware fiber.Handler)
    Migrate() error
}
```

Setiap modul mengimplementasikan interface ini dan didaftarkan di `main.go`.

### 9.3 Generic Repository

```go
// internal/shared/repository/repository.go
type Repository[T any] struct{}

// Methods: Create, Update, Delete, CountById, Count, FindById
```

Dipakai sebagai embedded struct oleh repository setiap modul (contoh: `UserRepository` meng-embed `Repository[entity.User]`).

### 9.4 Response Format

```go
// WebResponse[T] — response standar untuk semua endpoint
type WebResponse[T any] struct {
    Data    T             `json:"data"`
    Message string        `json:"message,omitempty"`
    Success bool          `json:"success,omitempty"`
    Paging  *PageMetadata `json:"paging,omitempty"`
}

// PageMetadata — untuk paginated response
type PageMetadata struct {
    Page      int   `json:"page"`
    Size      int   `json:"size"`        // SDD: "page_size" → codebase: "size"
    TotalItem int64 `json:"total_item"`  // SDD: "total_data" → codebase: "total_item"
    TotalPage int64 `json:"total_page"`
}

// ApiErrorResponse — untuk error via Fiber ErrorHandler
type ApiErrorResponse struct {
    Message    string              `json:"message"`
    StatusCode int                 `json:"statusCode"`
    Errors     map[string][]string `json:"errors,omitempty"`
}
```

### 9.5 JWT & Auth Middleware

- **Signing**: HS256, secret dari `config.json → jwt.secret`
- **Expiry**: 72 jam dari waktu generate
- **Claims**: `id` (user ID), `username`, `exp`, `iat`
- **Middleware**: extract `Authorization: Bearer {token}` → validate → inject `Auth{ID}` ke `ctx.Locals("auth")`

### 9.6 Logging

- Logrus dengan `JSONFormatter`
- Level dikonfigurasi via `config.json → log.level` (default: 6 = Trace)

### 9.7 Swagger / OpenAPI

- Route: `GET /swagger/*` (fiber-swagger)
- Annotations di `main.go`: `@title`, `@version`, `@BasePath`, `@securityDefinitions.apikey`

### 9.8 Migration Strategy

Dua mekanisme berjalan bersamaan (OQ-7):

1. **GORM AutoMigrate** — dijalankan saat startup via `module.Migrate()`. Untuk development convenience.
2. **SQL migration files** — di `db/migrations/`. Sebagai dokumentasi / referensi skema resmi.

---

## 10. Deployment

### 10.1 Dockerfile

```dockerfile
# Build stage
FROM golang:1.25.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/web/main.go

# Runtime stage
FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /app/server .
COPY --from=builder /app/config.json .
EXPOSE 8080
CMD ["./server"]
```

> **Gap vs SDD**: SDD menyebut `golang:1.22-alpine`, codebase menggunakan `golang:1.25.3-alpine`. SDD build path `./cmd/api`, codebase `./cmd/web/main.go`.

### 10.2 docker-compose.yml

| Service | Image | Port | Network |
| --- | --- | --- | --- |
| `aic_be` | Build from Dockerfile | 8080:8080 | aic_network |
| `aic_mysql` | mysql:8.0 | ${DB_PORT_EXTERNAL}:3306 | aic_network |

Env vars via `.env`: `MYSQL_ROOT_PASSWORD`, `MYSQL_DATABASE`, `MYSQL_USER`, `MYSQL_PASSWORD`, `DB_PORT_EXTERNAL`.

Volume: `mysql_data` (persistent data MySQL).

> **Gap vs SDD**: SDD menyebut service names `api` dan `db` + inline env. Codebase: `aic_be` dan `aic_mysql` + `.env` file.

---

## 11. Gap Analysis — PRD vs SDD vs Codebase

### 11.1 Modul / Fitur Gap

| Item | PRD | SDD | Codebase | Status | Keputusan |
| --- | --- | --- | --- | --- | --- |
| Auth via OAuth | ✅ FR-01 | ✅ §8.1, §11 | ❌ Username/password | **Divergen** | Pakai username/password (OQ-1) |
| Modul `store` | ✅ FR-02 | ✅ §3, §6.2, §8.2 | ❌ Belum ada | Gap | To be built |
| Modul `product` | ✅ FR-03–05 | ✅ §3, §6.3, §8.3 | 🟡 Stub kosong | Gap | To be built |
| Modul `transaction` | ✅ FR-05–15 | ✅ §3, §6.4–6.5, §8.4 | ❌ Belum ada | Gap | To be built |
| Modul `report` | ✅ FR-16–20 | ✅ §3, §8.5 | ❌ Belum ada | Gap | To be built |
| Modul `restock` | ✅ FR-21–22 | ✅ §3, §6.6, §8.6 | ❌ Belum ada | Gap | To be built |
| Modul `promotion` | ✅ FR-23–25 | ✅ §3, §6.7, §8.6 | ❌ Belum ada | Gap | To be built |
| Cron job scheduler | ✅ implied | ✅ §13 (robfig/cron) | ❌ Belum ada | Gap | To be built |
| ML client interface | ✅ implied | ✅ §10 | ❌ Belum ada | Gap | To be built |

### 11.2 Skema Data Gap

| Item | SDD | Codebase | Keputusan |
| --- | --- | --- | --- |
| `users.password` | Tidak ada | ✅ Ada (bcrypt) | Codebase benar (auth bukan OAuth) |
| User ID format | UUID | Custom `usr_xxx_0000` | Custom format (OQ-2) |
| Timestamp type | SQL TIMESTAMP | BIGINT (milli-epoch) | BIGINT milli-epoch (OQ-3) |
| `username` length | — | GORM=100 vs migration=50 | Migration benar: 50 (OQ-9) |
| `email` length | — | GORM=255 vs migration=100 | Migration benar: 100 (OQ-9) |

### 11.3 Penamaan / Konvensi Gap

| Item | SDD | Codebase | Keputusan |
| --- | --- | --- | --- |
| Entrypoint | `cmd/api/` | `cmd/web/` | `cmd/web/` (OQ-4) |
| Module folder | `internal/modules/` | `internal/module/` | Rename ke `modules/` (OQ-5) |
| Layer naming | `handler.go`, `service.go` (flat) | `src/controller/`, `src/usecase/` (nested) | Codebase (OQ-6) |
| Page metadata field | `page_size`, `total_data` | `size`, `total_item` | Codebase (OQ-8) |

### 11.4 Infrastruktur Gap

| Item | SDD | Codebase | Keputusan |
| --- | --- | --- | --- |
| Database support | MySQL/PostgreSQL | MySQL only | Accepted |
| Swagger | Tidak disebut | Ada | Dokumentasikan resmi (OQ-11) |
| Go version | 1.22 | 1.25.3 | Codebase |
| Static files `/public` | Tidak disebut | Ada (tidak dipakai) | Legacy, abaikan (OQ-13) |
| Fiber BodyLimit 10MB | Tidak disebut | Ada | Accepted (untuk upload audio/foto) |
| `config.group.id` | Tidak disebut | Ada | Legacy, dibiarkan (OQ-10) |

---

## 12. FR Traceability Matrix

| FR-ID | Deskripsi | Modul | Endpoint | Codebase Status | Catatan |
| --- | --- | --- | --- | --- | --- |
| FR-01 | Registrasi/login pedagang | auth | `POST /api/users`, `POST /api/users/_login`, `GET /api/users/_current` | ✅ Implemented | Via username/password, bukan OAuth |
| FR-02 | Membuat toko (create store) | store | `POST /api/stores`, `GET /api/stores` | ❌ Not Started | Modul belum ada |
| FR-03 | Menambah produk | product | `POST /api/products` | ❌ Stub Only | Folder ada, kode kosong |
| FR-04 | Simpan harga modal per produk | product | — (internal) | ❌ Stub Only | — |
| FR-05 | Decrement stok otomatis | product + transaction | `POST /api/transactions` | ❌ Not Started | — |
| FR-06 | Rekam suara transaksi | transaction | `POST /api/transactions/extract/voice` | ❌ Not Started | — |
| FR-07 | Whisper speech-to-text | transaction + mlclient | — (internal) | ❌ Not Started | ML client belum ada |
| FR-08 | NLP extraction | transaction + mlclient | — (internal) | ❌ Not Started | — |
| FR-09 | Matching produk (Golang) | transaction + product-client | — (internal) | ❌ Not Started | — |
| FR-10 | Preview editable | transaction | Response dari extract endpoint | ❌ Not Started | — |
| FR-11 | Konfirmasi transaksi | transaction | `POST /api/transactions` | ❌ Not Started | — |
| FR-12 | Upload foto nota | transaction | `POST /api/transactions/extract/photo` | ❌ Not Started | — |
| FR-13 | OCR foto nota | transaction + mlclient | — (internal) | ❌ Not Started | — |
| FR-14 | NLP extraction dari OCR | transaction + mlclient | — (internal) | ❌ Not Started | Reuse model yg sama dgn suara |
| FR-15 | Preview editable (foto) | transaction | Response dari extract/photo | ❌ Not Started | — |
| FR-16 | Total omset, untung, jml transaksi | report | `GET /api/reports/daily` | ❌ Not Started | On-the-fly, bukan tabel ringkasan |
| FR-17 | Untung dari snapshot harga | report | — (internal) | ❌ Not Started | — |
| FR-18 | Produk terlaris | report | `GET /api/reports/daily` | ❌ Not Started | — |
| FR-19 | Sisa stok produk | report | `GET /api/reports/daily` | ❌ Not Started | — |
| FR-20 | Laporan read-only (tanpa closing) | report | — | ❌ Not Started | — |
| FR-21 | Prediksi restock (cron) | restock | Background job | ❌ Not Started | — |
| FR-22 | Simpan ke restock_predictions | restock | `GET /api/restock-predictions` | ❌ Not Started | — |
| FR-23 | Cek rule promo (cron) | promotion | Background job | ❌ Not Started | — |
| FR-24 | Simpan ke promotion_logs | promotion | `GET /api/promotions` | ❌ Not Started | — |
| FR-25 | Notifikasi WA | — | — | ❌ Out of scope MVP | Won't (MVP) |

---

## 13. Known Issues

### 13.1 Bug: `CountByUsername` Query Menggunakan `WHERE id = ?`

**File**: `internal/module/auth/src/repository/user_repository.go`, line 26

```go
// Method bernama CountByUsername tapi query memakai id
func (r *UserRepository) CountByUsername(db *gorm.DB, username any) (int64, error) {
    var total int64
    err := db.Model(new(entity.User)).Where("id = ?", username).Count(&total).Error
    return total, err
}
```

**Impact**: Pada flow Register, method ini dipanggil untuk mengecek apakah username sudah ada. Karena query menggunakan `WHERE id = ?` alih-alih `WHERE username = ?`, pengecekan duplikat username **tidak berfungsi dengan benar**. Register bisa saja lolos pengecekan ini, tapi akan gagal di `INSERT` karena constraint `UNIQUE` pada kolom `username` di database.

**Fix yang direkomendasikan**: Ubah query menjadi `WHERE username = ?`.

### 13.2 GORM Entity vs SQL Migration Length Mismatch

**File entity**: `internal/module/auth/src/entity/user_entity.go`
**File migration**: `db/migrations/20260709160912_create_table_users.up.sql`

| Field | GORM Entity | SQL Migration | Yang Benar |
| --- | --- | --- | --- |
| `username` | `varchar(100)` | `VARCHAR(50)` | `VARCHAR(50)` — migration |
| `email` | `varchar(255)` | `VARCHAR(100)` | `VARCHAR(100)` — migration |

**Impact**: Jika GORM `AutoMigrate` dijalankan, GORM bisa saja mengubah kolom ke ukuran yang lebih besar dari yang dimaksud. Jika SQL migration dijalankan duluan, `AutoMigrate` mungkin mendeteksi perbedaan dan mencoba alter tabel.

**Fix yang direkomendasikan**: Update GORM entity agar cocok dengan SQL migration (`varchar(50)` dan `varchar(100)`).

### 13.3 Login Menggunakan DB Transaction untuk Operasi Read-Only

**File**: `internal/module/auth/src/usecase/auth_usecase.go`, fungsi `Login()`

Login hanya melakukan `SELECT` (read-only), tapi kode membuka `tx.Begin()` dan `tx.Commit()`. Tidak ada side effect, tapi ini overhead yang tidak perlu dan bisa membingungkan developer lain.

**Rekomendasi**: Gunakan `db.WithContext(ctx)` langsung tanpa `Begin()`/`Commit()`.

### 13.4 Product Module Belum Memvalidasi Kepemilikan Store ID

**File**: `internal/module/product/src/usecase/product_usecase.go` dan `product_controller.go`

**Impact**: Saat ini endpoint Product menerima `store_id` (via body atau query) dan memercayainya tanpa melakukan pengecekan kepemilikan. Seharusnya (sesuai SDD §11), `product` modul perlu mengecek ke modul `store` via `store-client` untuk memastikan bahwa `store_id` tersebut benar-benar milik user yang saat ini sedang login. Karena modul `store` belum diimplementasikan, validasi ini hilang.

**Rekomendasi**: Setelah modul `store` diimplementasikan, tambahkan pemanggilan `store-client.GetStoreByOwnerID` atau `GetStoreByID` untuk memverifikasi hak akses pada `store_id` di request.

---

## 14. Pending Action Items (Ringkasan)

Berdasarkan gap analysis, berikut action items yang perlu dikerjakan:

| # | Action | Prioritas | Terkait |
| --- | --- | --- | --- |
| 1 | Rename `internal/module/` → `internal/modules/` | Medium | OQ-5 |
| 2 | Fix bug `CountByUsername` query | High | §13.1 |
| 3 | Fix GORM entity field length (username=50, email=100) | Medium | §13.2, OQ-9 |
| 4 | Build modul `store` | High | FR-02, dependency modul lain |
| 6 | Build modul `transaction` | High | FR-05–FR-15 |
| 7 | Build modul `report` | High | FR-16–FR-20 |
| 8 | Build modul `restock` + cron job | Medium | FR-21–FR-22 |
| 9 | Build modul `promotion` + cron job | Medium | FR-23–FR-24 |
| 10 | Setup `mlclient` interface + mock | Medium | SDD §10 |
| 11 | Update PRD/SDD untuk reflect keputusan OQ | Low | Semua OQ |

---

_Dokumen ini terakhir diperbarui pada 12 Juli 2026. Selalu rujuk ke codebase aktual sebagai source of truth utama._
