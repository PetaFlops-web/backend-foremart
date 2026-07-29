# Progress Tracking — Smart Commerce Backend

> **Referensi**: [prd.md](prd.md) | [sdd.md](sdd.md) | [SYSTEM_MAP.md](SYSTEM_MAP.md) | [workflow.md](../.agents/workflow.md)
>
> File ini adalah **satu-satunya tempat** tracking task teknis yang bisa dicentang.
>
> - `[x]` = done
> - `[ ]` = belum dimulai / belum selesai
> - Setiap task mencantumkan rujukan FR-ID (prd.md) dan/atau section SDD.
> - Urutan modul mengikuti dependency graph — modul yang dibutuhkan modul lain diletakkan lebih dulu.

---

## Ringkasan Status

| #   | Modul                   | Done   | Total  | Status                    |
| --- | ----------------------- | ------ | ------ | ------------------------- |
| 1   | Shared / Infrastructure | 8      | 9      | 🟢 Hampir selesai         |
| 2   | Auth                    | 8      | 11     | 🟡 Bug fix pending        |
| 3   | Store                   | 0      | 7      | ✅ Selesai dimulai        |
| 4   | Product                 | 10     | 10     | ✅ Selesai diimplementasi |
| 5   | Transaction             | 9      | 10     | 🟢 Hampir selesai         |
| 6   | ML Client / Integration | 2      | 2      | ✅ Selesai diimplementasi |
| 7   | Report                  | 0      | 3      | ❌ Belum dimulai          |
| 8   | Restock                 | 0      | 6      | ❌ Belum dimulai          |
| 9   | Promotion               | 0      | 6      | ❌ Belum dimulai          |
| 10  | Cron / Scheduler        | 0      | 2      | ❌ Belum dimulai          |
|     | **TOTAL**               | **26** | **66** | **~39%**                  |

---

## Dependency Graph (Urutan Pengerjaan)

```
1. Shared/Infra    ── ✅ sudah selesai (kecuali rename — owner handle)
2. Auth            ── ✅ sudah selesai (bug fix pending)
3. Store           ── dibutuhkan oleh: product, transaction, report, restock, promotion
4. Product         ── dibutuhkan oleh: transaction, report, restock, promotion
5. ML Client       ── dibutuhkan oleh: transaction (extract endpoints)
6. Transaction     ── dibutuhkan oleh: report, restock, promotion
7. Report          ── depends on: transaction-client + product-client
8. Cron/Scheduler  ── infrastruktur untuk restock + promotion
9. Restock         ── depends on: transaction-client + product-client + scheduler
10. Promotion      ── depends on: transaction-client + product-client + scheduler
```

---

## Task Breakdown

### 1. Shared / Infrastructure

- [x] Config setup: Viper, GORM (MySQL), Fiber, Logrus, Validator — SDD §4
- [x] Module interface (`module.Module`: Migrate + RegisterRoutes) — SDD §2.4
- [x] Generic `Repository[T]` base struct — SDD §2.4
- [x] `WebResponse[T]`, `PageMetadata`, `ApiErrorResponse` — SDD §5
- [x] JWT & Auth middleware (HS256, 72h, Bearer token) — SDD §11
- [x] ID generation utility (`generate_user_id.go`) — codebase convention
- [x] Swagger setup (fiber-swagger, route `/swagger/*`) — codebase
- [x] Dockerfile (multi-stage) + docker-compose.yml — SDD §16
- [x] Rename `internal/module/` → `internal/modules/` _(owner handle di branch `chore/setup-project`)_ — SYSTEM_MAP §14 #1

---

### 2. Auth Module

- [x] Entity `User` + SQL migration `create_table_users` — SDD §6.1, FR-01
- [x] Repository `UserRepository` (FindByUsername, CountByUsername) — SDD §6.1
- [x] Usecase `AuthUseCase` (Register, Login, Current) — SDD §8.1, FR-01
- [x] Endpoint `POST /api/users` — Register — SDD §8.1, FR-01
- [x] Endpoint `POST /api/users/_login` — Login — SDD §8.1, FR-01
- [x] Endpoint `GET /api/users/_current` — Get current user (protected) — SDD §8.1, FR-01
- [x] auth-client interface + client_impl (`GetUserByID`) — SDD §7
- [x] Wiring: module.go, route.go, register di main.go — SDD §2.5
- [ ] Fix bug: `CountByUsername` query `WHERE id = ?` → `WHERE username = ?` — SYSTEM_MAP §13.1
- [ ] Fix: GORM entity field length mismatch (username→varchar(50), email→varchar(100)) — SYSTEM_MAP §13.2
- [ ] Refactor: Hapus DB transaction di Login (read-only, tidak perlu Begin/Commit) — SYSTEM_MAP §13.3

---

### 3. Store Module

- [x] Entity `Store` + SQL migration `create_table_stores` — SDD §6.2, FR-02
- [x] Repository `StoreRepository` — SDD §6.2
- [x] Usecase `StoreUseCase` — SDD §8.2, FR-02
- [x] Endpoint `POST /api/stores` — Create store (FR-02) — SDD §8.2
- [x] Endpoint `GET /api/stores` — Get store by owner (FR-02) — SDD §8.2
- [x] Endpoint `PUT /api/stores` — Update store — SDD §8.2
- [x] store-client interface + client_impl (`GetStoreByUserID`, `GetStoreByID`) — SDD §7
- [x] Wiring: module.go, route.go, register di main.go — SDD §2.5

---

### 4. Product Module

- [x] Entity `Product` + SQL migration `create_table_products` — SDD §6.3, FR-03
- [x] Repository `ProductRepository` — SDD §6.3
- [x] Usecase `ProductUseCase` — SDD §8.3, FR-03–FR-05
- [x] Endpoint `POST /api/products` — Create product (FR-03) — SDD §8.3
- [x] Endpoint `GET /api/products` — List products, paginated — SDD §8.3
- [x] Endpoint `GET /api/products/:id` — Get product detail — SDD §8.3
- [x] Endpoint `PUT /api/products/:id` — Update product — SDD §8.3
- [x] Endpoint `DELETE /api/products/:id` — Delete product — SDD §8.3
- [x] product-client interface + client_impl (`GetByID`, `ListByStoreID`, `DecrementStock`, `Search`) — SDD §7
- [x] Wiring: module.go, route.go, register di main.go — SDD §2.5

---

### 5. Transaction Module

- [x] Entity `Transaction` + SQL migration `create_table_transactions` — SDD §6.4
- [x] Entity `TransactionItem` + SQL migration `create_table_transaction_items` — SDD §6.5
- [x] Repository `TransactionRepository` + `TransactionItemRepository` — SDD §6.4–6.5
- [x] Usecase `TransactionUseCase` — konfirmasi & preview logic — SDD §8.4, §9.1–9.2
- [x] Endpoint `POST /api/transactions` — Confirm & save transaction + decrement stock (FR-05, FR-11) — SDD §8.4, §9.2
- [x] Endpoint `POST /api/transactions/extract/voice` — Voice → ML → preview (FR-06–FR-10) — SDD §8.4, §9.1
- [ ] Endpoint `POST /api/transactions/extract/photo` — Photo → ML → preview (FR-12–FR-15) — SDD §8.4, §9.1
- [x] Endpoint `GET /api/transactions` — List transaction history — SDD §8.4
- [x] transaction-client interface + client_impl (`ListByStoreAndDate`, `ListItemsByStoreAndDateRange`, `ListItemsByProduct`, `SumQtyByProductInMonth`) — SDD §7, §9.4, §9.5
- [x] Wiring: module.go, route.go, register di main.go — SDD §2.5

---

### 6. ML Client / Integration

- [x] `MLClient` interface (`TranscribeAndExtract`, `OcrAndExtract`) — SDD §10.1
- [x] Mock implementation of `MLClient` (dummy response data) — SDD §10 (Note: Built real HTTP client instead)

---

### 7. Report Module

> Report module tidak punya tabel sendiri — query via transaction-client dan product-client.

- [ ] Usecase `ReportUseCase` — on-the-fly calculation (omset, untung, produk terlaris, sisa stok) — SDD §9.3, FR-16–FR-20
- [ ] Endpoint `GET /api/reports/daily?date=YYYY-MM-DD` — Daily report (FR-16–FR-20) — SDD §8.5, §9.3
- [ ] Wiring: module.go, route.go, register di main.go — SDD §2.5

---

### 8. Restock Module

- [ ] Entity `RestockPrediction` + SQL migration `create_table_restock_predictions` — SDD §6.6, FR-21–FR-22
- [ ] Repository `RestockRepository` — SDD §6.6
- [ ] Usecase `RestockUseCase` — prediction logic — SDD §9.4, FR-21–FR-22
- [ ] Endpoint `GET /api/restock-predictions` — List predictions (FR-21–FR-22) — SDD §8.6
- [ ] Cron job `restock_job.go` — nightly prediction — SDD §9.4, §13, FR-21
- [ ] Wiring: module.go, route.go, register di main.go — SDD §2.5

---

### 9. Promotion Module

- [ ] Entity `PromotionLog` + SQL migration `create_table_promotion_logs` — SDD §6.7, FR-23–FR-24
- [ ] Repository `PromotionRepository` — SDD §6.7
- [ ] Usecase `PromotionUseCase` — rule evaluation logic — SDD §9.5, FR-23–FR-24
- [ ] Endpoint `GET /api/promotions` — List promotion logs (FR-23–FR-24) — SDD §8.6
- [ ] Cron job `promotion_job.go` — nightly rule check — SDD §9.5, §13, FR-23
- [ ] Wiring: module.go, route.go, register di main.go — SDD §2.5

---

### 10. Cron / Scheduler Infrastructure

- [ ] Setup scheduler (`robfig/cron`) di main.go — SDD §13
- [ ] Wire `restock_job` + `promotion_job` ke scheduler — SDD §13

---

_Terakhir diperbarui: 13 Juli 2026._
