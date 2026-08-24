# API Endpoints — Smart Commerce Backend

Dokumen ini merepresentasikan seluruh endpoint API beserta **payload (request)** dan **response** yang didapat dari operasi CRUD. Seluruh contoh di bawah ini adalah hasil pemanggilan **aktual** terhadap backend yang berjalan via Docker Compose.

## Informasi Umum

| Item             | Nilai                                                             |
| ---------------- | ----------------------------------------------------------------- |
| **Base URL**     | `http://127.0.0.1:8080`                                           |
| **API Prefix**   | `/api`                                                            |
| **Swagger UI**   | `http://127.0.0.1:8080/swagger/index.html`                       |
| **Content-Type** | `application/json`                                               |
| **Auth**         | JWT Bearer Token — header `Authorization: Bearer <token>`         |

### Standar Response

Semua endpoint memakai struktur `WebResponse[T]`.

**Sukses:**

```json
{
  "data": { "...": "..." },
  "message": "Pesan sukses",
  "success": true
}
```

**Sukses dengan pagination:**

```json
{
  "data": [ "..." ],
  "message": "Pesan sukses",
  "success": true,
  "paging": {
    "page": 1,
    "size": 10,
    "total_item": 2,
    "total_page": 1
  }
}
```

**Error (validasi / bisnis):**

```json
{
  "message": "Pesan error yang jelas",
  "statusCode": 400
}
```

**Error (unauthorized / middleware):**

```json
{
  "errors": "Unauthorized"
}
```

### Ringkasan Endpoint

| Modul       | Method   | Path                                        | Auth | Keterangan                          |
| ----------- | -------- | ------------------------------------------- | :--: | ----------------------------------- |
| Auth        | `POST`   | `/api/users`                                |  ❌  | Register user baru                  |
| Auth        | `POST`   | `/api/users/_login`                         |  ❌  | Login & dapatkan JWT                |
| Auth        | `GET`    | `/api/users/_current`                       |  ✅  | Profil user yang sedang login       |
| Store       | `POST`   | `/api/stores`                               |  ✅  | Buat toko                           |
| Store       | `GET`    | `/api/stores/:id`                           |  ✅  | Detail toko                         |
| Store       | `PUT`    | `/api/stores/:id`                           |  ✅  | Update toko                         |
| Product     | `POST`   | `/api/products`                             |  ✅  | Buat produk                         |
| Product     | `GET`    | `/api/products`                             |  ✅  | Daftar produk (pagination)          |
| Product     | `GET`    | `/api/products/:id`                         |  ✅  | Detail produk                       |
| Product     | `PUT`    | `/api/products/:id`                         |  ✅  | Update produk                       |
| Product     | `DELETE` | `/api/products/:id`                         |  ✅  | Hapus produk                        |
| Transaction | `POST`   | `/api/transactions`                         |  ✅  | Simpan transaksi (kurangi stok)     |
| Transaction | `GET`    | `/api/transactions`                         |  ✅  | Riwayat transaksi (pagination)      |
| Transaction | `GET`    | `/api/transactions/:id`                     |  ✅  | Detail transaksi                    |
| Transaction | `DELETE` | `/api/transactions/:id`                     |  ✅  | Hapus transaksi (kembalikan stok)   |
| Report      | `GET`    | `/api/reports/daily`                        |  ✅  | Laporan harian toko                 |
| Restock     | `POST`   | `/api/restock-predictions/_generate`        |  ✅  | Generate prediksi restock (ML)      |
| Restock     | `GET`    | `/api/restock-predictions`                  |  ✅  | Daftar prediksi restock tersimpan   |

---

## 1. Auth

### 1.1 Register — `POST /api/users`

Membuat user (merchant) baru. Mengembalikan JWT token langsung.

**Request Body**

| Field      | Tipe   | Wajib | Aturan                |
| ---------- | ------ | :---: | --------------------- |
| `username` | string |   ✅  | max 100 karakter      |
| `email`    | string |   ✅  | format email, max 255 |
| `password` | string |   ✅  | min 6, max 100        |

```json
{
  "username": "toko_baru",
  "email": "[EMAIL_ADDRESS]",
  "password": "rahasia123"
}
```

**Response `200 OK`**

```json
{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "usr_toko_32448_4756",
      "username": "toko_32448",
      "email": "[EMAIL_ADDRESS]",
      "created_at": 1787213071994,
      "updated_at": 1787213071994
    }
  },
  "message": "Register successful",
  "success": true
}
```

**Kemungkinan Error**

| Status | Kondisi                                          |
| :----: | ------------------------------------------------ |
| `400`  | Body tidak valid / gagal validasi field          |
| `409`  | `Username already taken` (username sudah dipakai) |

### 1.2 Login — `POST /api/users/_login`

**Request Body**

| Field      | Tipe   | Wajib |
| ---------- | ------ | :---: |
| `username` | string |   ✅  |
| `password` | string |   ✅  |

```json
{
  "username": "tokosaya",
  "password": "rahasia123"
}
```

**Response `200 OK`**

```json
{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "usr_tokosaya_8896",
      "username": "tokosaya",
      "email": "[EMAIL_ADDRESS]",
      "created_at": 1787212945335,
      "updated_at": 1787212945335
    }
  },
  "message": "Login successful",
  "success": true
}
```

**Kemungkinan Error**

| Status | Kondisi                                          |
| :----: | ------------------------------------------------ |
| `400`  | Body tidak valid                                 |
| `404`  | `Username atau password anda salah`              |

### 1.3 Current User — `GET /api/users/_current`

Header: `Authorization: Bearer <token>`

**Response `200 OK`**

```json
{
  "data": {
    "id": "usr_tokosaya_8896",
    "username": "tokosaya",
    "email": "[EMAIL_ADDRESS]",
    "created_at": 1787212945335,
    "updated_at": 1787212945335
  },
  "message": "Get current user successful",
  "success": true
}
```

**Tanpa token** → `401`:

```json
{ "errors": "Unauthorized" }
```

---

## 2. Store

> Semua endpoint store memerlukan header `Authorization`. Satu user hanya boleh memiliki **satu** toko.

### 2.1 Create Store — `POST /api/stores`

**Request Body**

| Field        | Tipe   | Wajib | Aturan          |
| ------------ | ------ | :---: | --------------- |
| `store_name` | string |   ✅  | min 3 karakter  |

> `user_id` diambil otomatis dari JWT, tidak perlu dikirim di body.

```json
{
  "store_name": "Toko Sembako Jaya"
}
```

**Response `200 OK`**

```json
{
  "data": {
    "id": "1146d375-6219-46b8-bfda-e772a8260fd5",
    "user_id": "usr_tokosaya_8896",
    "store_name": "Toko Sembako Jaya",
    "created_at": 1787213049709,
    "updated_at": 1787213049709
  },
  "message": "Berhasil menambahkan store",
  "success": true
}
```

**Kemungkinan Error**

| Status | Kondisi                                     |
| :----: | ------------------------------------------- |
| `400`  | Body tidak valid                            |
| `401`  | Tidak ada / token invalid                   |
| `409`  | `User already has a store`                  |

### 2.2 Get Store — `GET /api/stores/:id`

**Response `200 OK`**

```json
{
  "data": {
    "id": "1146d375-6219-46b8-bfda-e772a8260fd5",
    "user_id": "usr_tokosaya_8896",
    "store_name": "Toko Sembako Jaya",
    "created_at": 1787213049709,
    "updated_at": 1787213049709
  },
  "message": "Berhasil mendapatkan store",
  "success": true
}
```

### 2.3 Update Store — `PUT /api/stores/:id`

**Request Body**

```json
{
  "store_name": "Toko Sembako Jaya Abadi"
}
```

**Response `200 OK`**

```json
{
  "data": {
    "id": "1146d375-6219-46b8-bfda-e772a8260fd5",
    "user_id": "usr_tokosaya_8896",
    "store_name": "Toko Sembako Jaya Abadi",
    "created_at": 1787213049709,
    "updated_at": 1787213049735
  },
  "message": "Berhasil mengubah data store",
  "success": true
}
```

---

## 3. Product

> Semua endpoint product memerlukan header `Authorization`.

### 3.1 Create Product — `POST /api/products`

**Request Body**

| Field           | Tipe   | Wajib | Aturan          |
| --------------- | ------ | :---: | --------------- |
| `store_id`      | string |   ✅  | —               |
| `product_name`  | string |   ✅  | —               |
| `cost_price`    | int64  |   ➖  | min 0 (default 0)|
| `selling_price` | int64  |   ✅  | min 0           |
| `stock`         | int    |   ➖  | min 0 (default 0)|
| `unit`          | string |   ✅  | mis. kg, pcs    |

```json
{
  "store_id": "1146d375-6219-46b8-bfda-e772a8260fd5",
  "product_name": "Gula Pasir 1 kg",
  "cost_price": 12000,
  "selling_price": 15000,
  "stock": 50,
  "unit": "kg"
}
```

**Response `200 OK`**

```json
{
  "data": {
    "id": "prod_9938",
    "store_id": "1146d375-6219-46b8-bfda-e772a8260fd5",
    "product_name": "Gula Pasir 1 kg",
    "cost_price": 12000,
    "selling_price": 15000,
    "stock": 50,
    "unit": "kg",
    "created_at": 1787213049742,
    "updated_at": 1787213049742
  },
  "message": "Berhasil menambahkan produk",
  "success": true
}
```

### 3.2 List Products — `GET /api/products`

**Query Params**

| Field      | Tipe   | Wajib | Default | Keterangan            |
| ---------- | ------ | :---: | ------- | --------------------- |
| `store_id` | string |   ✅  | —       | ID toko               |
| `name`     | string |   ➖  | —       | Filter keyword nama   |
| `page`     | int    |   ➖  | 1       | Nomor halaman         |
| `size`     | int    |   ➖  | 10      | Ukuran halaman        |

Contoh: `GET /api/products?store_id=1146d375-...&page=1&size=10`

**Response `200 OK`**

```json
{
  "data": [
    {
      "id": "prod_8829",
      "store_id": "1146d375-6219-46b8-bfda-e772a8260fd5",
      "product_name": "Minyak Goreng 2L",
      "cost_price": 28000,
      "selling_price": 34000,
      "stock": 30,
      "unit": "pcs",
      "created_at": 1787213049767,
      "updated_at": 1787213049767
    },
    {
      "id": "prod_9938",
      "store_id": "1146d375-6219-46b8-bfda-e772a8260fd5",
      "product_name": "Gula Pasir Premium 1 kg",
      "cost_price": 12500,
      "selling_price": 16000,
      "stock": 45,
      "unit": "kg",
      "created_at": 1787213049742,
      "updated_at": 1787213049801
    }
  ],
  "message": "Berhasil menampilkan daftar produk",
  "success": true,
  "paging": {
    "page": 1,
    "size": 10,
    "total_item": 2,
    "total_page": 1
  }
}
```

### 3.3 Get Product — `GET /api/products/:id`

**Query Params:** `store_id` (wajib).

Contoh: `GET /api/products/prod_9938?store_id=1146d375-...`

**Response `200 OK`**

```json
{
  "data": {
    "id": "prod_9938",
    "store_id": "1146d375-6219-46b8-bfda-e772a8260fd5",
    "product_name": "Gula Pasir 1 kg",
    "cost_price": 12000,
    "selling_price": 15000,
    "stock": 50,
    "unit": "kg",
    "created_at": 1787213049742,
    "updated_at": 1787213049742
  },
  "message": "Berhasil mengambil data produk",
  "success": true
}
```

Jika `store_id` kosong → `400 Store ID tidak boleh kosong`.

### 3.4 Update Product — `PUT /api/products/:id`

**Request Body** (sama seperti create; `id` diambil dari path).

```json
{
  "store_id": "1146d375-6219-46b8-bfda-e772a8260fd5",
  "product_name": "Gula Pasir Premium 1 kg",
  "cost_price": 12500,
  "selling_price": 16000,
  "stock": 45,
  "unit": "kg"
}
```

**Response `200 OK`**

```json
{
  "data": {
    "id": "prod_9938",
    "store_id": "1146d375-6219-46b8-bfda-e772a8260fd5",
    "product_name": "Gula Pasir Premium 1 kg",
    "cost_price": 12500,
    "selling_price": 16000,
    "stock": 45,
    "unit": "kg",
    "created_at": 1787213049742,
    "updated_at": 1787213049801
  },
  "message": "Berhasil mengubah data produk",
  "success": true
}
```

### 3.5 Delete Product — `DELETE /api/products/:id`

**Query Params:** `store_id` (wajib).

Contoh: `DELETE /api/products/prod_8829?store_id=1146d375-...`

**Response `200 OK`**

```json
{
  "data": null,
  "message": "Berhasil menghapus produk",
  "success": true
}
```

---

## 4. Transaction

> Semua endpoint transaction memerlukan header `Authorization`.
> Membuat transaksi akan **mengurangi stok** produk; menghapus transaksi akan **mengembalikan stok**.

### 4.1 Create Transaction — `POST /api/transactions`

**Request Body**

| Field      | Tipe                    | Wajib | Keterangan                       |
| ---------- | ----------------------- | :---: | -------------------------------- |
| `store_id` | string                  |   ✅  | ID toko                          |
| `source`   | string                  |   ✅  | Sumber, mis. `manual` / `voice`  |
| `items`    | array                   |   ✅  | Minimal 1 item                   |

**Item (`items[]`)**

| Field                 | Tipe   | Wajib | Aturan   |
| --------------------- | ------ | :---: | -------- |
| `product_id`          | string |   ✅  | —        |
| `qty`                 | int    |   ✅  | min 1    |
| `selling_price_final` | int64  |   ✅  | min 0    |

```json
{
  "store_id": "1146d375-6219-46b8-bfda-e772a8260fd5",
  "source": "manual",
  "items": [
    { "product_id": "prod_9938", "qty": 2, "selling_price_final": 16000 },
    { "product_id": "prod_8829", "qty": 1, "selling_price_final": 34000 }
  ]
}
```

**Response `200 OK`**

```json
{
  "data": {
    "id": "txn_8557",
    "store_id": "1146d375-6219-46b8-bfda-e772a8260fd5",
    "transaction_date": "2026-08-20",
    "source": "manual",
    "created_at": 1787213049816,
    "items": [
      {
        "id": "txni_7967",
        "product_id": "prod_9938",
        "product_name_snapshot": "Gula Pasir Premium 1 kg",
        "qty": 2,
        "cost_price_snapshot": 12500,
        "selling_price_snapshot": 16000
      },
      {
        "id": "txni_1575",
        "product_id": "prod_8829",
        "product_name_snapshot": "Minyak Goreng 2L",
        "qty": 1,
        "cost_price_snapshot": 28000,
        "selling_price_snapshot": 34000
      }
    ]
  },
  "message": "Berhasil menyimpan transaksi",
  "success": true
}
```

> Catatan: `cost_price_snapshot` & `selling_price_snapshot` disimpan sebagai **snapshot** harga pada saat transaksi dibuat, sehingga perubahan harga produk di kemudian hari tidak mengubah histori transaksi.

### 4.2 List Transactions — `GET /api/transactions`

**Query Params**

| Field      | Tipe   | Wajib | Default |
| ---------- | ------ | :---: | ------- |
| `store_id` | string |   ✅  | —       |
| `page`     | int    |   ➖  | 1       |
| `size`     | int    |   ➖  | 10      |

**Response `200 OK`**

```json
{
  "data": [
    {
      "id": "txn_8557",
      "store_id": "1146d375-6219-46b8-bfda-e772a8260fd5",
      "transaction_date": "2026-08-20",
      "source": "manual",
      "created_at": 1787213049816,
      "items": [
        {
          "id": "txni_1575",
          "product_id": "prod_8829",
          "product_name_snapshot": "Minyak Goreng 2L",
          "qty": 1,
          "cost_price_snapshot": 28000,
          "selling_price_snapshot": 34000
        },
        {
          "id": "txni_7967",
          "product_id": "prod_9938",
          "product_name_snapshot": "Gula Pasir Premium 1 kg",
          "qty": 2,
          "cost_price_snapshot": 12500,
          "selling_price_snapshot": 16000
        }
      ]
    }
  ],
  "message": "Berhasil menampilkan riwayat transaksi",
  "success": true,
  "paging": {
    "page": 1,
    "size": 10,
    "total_item": 1,
    "total_page": 1
  }
}
```

### 4.3 Get Transaction — `GET /api/transactions/:id`

**Query Params:** `store_id` (wajib).

Contoh: `GET /api/transactions/txn_8557?store_id=1146d375-...`

**Response `200 OK`**

```json
{
  "data": {
    "id": "txn_8557",
    "store_id": "1146d375-6219-46b8-bfda-e772a8260fd5",
    "transaction_date": "2026-08-20",
    "source": "manual",
    "created_at": 1787213049816,
    "items": [
      {
        "id": "txni_1575",
        "product_id": "prod_8829",
        "product_name_snapshot": "Minyak Goreng 2L",
        "qty": 1,
        "cost_price_snapshot": 28000,
        "selling_price_snapshot": 34000
      },
      {
        "id": "txni_7967",
        "product_id": "prod_9938",
        "product_name_snapshot": "Gula Pasir Premium 1 kg",
        "qty": 2,
        "cost_price_snapshot": 12500,
        "selling_price_snapshot": 16000
      }
    ]
  },
  "message": "Berhasil mengambil data transaksi",
  "success": true
}
```

### 4.4 Delete Transaction — `DELETE /api/transactions/:id`

**Query Params:** `store_id` (wajib). Menghapus transaksi akan mengembalikan (increment) stok produk.

Contoh: `DELETE /api/transactions/txn_8557?store_id=1146d375-...`

**Response `200 OK`**

```json
{
  "data": null,
  "message": "Berhasil menghapus transaksi",
  "success": true
}
```

---

## 5. Report

### 5.1 Daily Report — `GET /api/reports/daily`

Menghitung laporan harian toko secara _on-the-fly_ berdasarkan transaksi dan stok terkini. Memerlukan header `Authorization`.

**Query Params**

| Field      | Tipe   | Wajib | Keterangan                                       |
| ---------- | ------ | :---: | ------------------------------------------------ |
| `store_id` | string |   ✅  | ID toko                                          |
| `date`     | string |   ➖  | Format `YYYY-MM-DD` (default: hari ini)          |

Contoh: `GET /api/reports/daily?store_id=1146d375-...`

**Response `200 OK`**

```json
{
  "data": {
    "store": {
      "id": "1146d375-6219-46b8-bfda-e772a8260fd5",
      "user_id": "usr_tokosaya_8896",
      "store_name": "Toko Sembako Jaya Abadi"
    },
    "date": "2026-08-20",
    "total_omset": 66000,
    "total_untung": 13000,
    "jumlah_transaksi": 1,
    "produk_terlaris": [
      {
        "product_id": "prod_9938",
        "product_name": "Gula Pasir Premium 1 kg",
        "qty_sold": 2,
        "current_stock": 43,
        "unit": "kg"
      },
      {
        "product_id": "prod_8829",
        "product_name": "Minyak Goreng 2L",
        "qty_sold": 1,
        "current_stock": 29,
        "unit": "pcs"
      }
    ],
    "sisa_stok": [
      {
        "product_id": "prod_9938",
        "product_name": "Gula Pasir Premium 1 kg",
        "stock": 43,
        "unit": "kg"
      },
      {
        "product_id": "prod_8829",
        "product_name": "Minyak Goreng 2L",
        "stock": 29,
        "unit": "pcs"
      }
    ]
  },
  "message": "Berhasil menampilkan laporan harian",
  "success": true
}
```

**Keterangan field**

| Field              | Tipe   | Keterangan                                          |
| ------------------ | ------ | --------------------------------------------------- |
| `total_omset`      | int64  | Total penjualan (harga jual × qty)                  |
| `total_untung`     | int64  | Total profit ((harga jual − harga modal) × qty)     |
| `jumlah_transaksi` | int    | Banyaknya transaksi pada tanggal tersebut           |
| `produk_terlaris`  | array  | Produk terurut berdasarkan qty terjual              |
| `sisa_stok`        | array  | Stok terkini seluruh produk toko                    |

---

## 6. Restock Prediction

> Memerlukan header `Authorization`. Menghitung rekomendasi restock berdasarkan prediksi penjualan harian dari layanan ML (`/predict-inventory`).

### 6.1 Generate Prediction — `POST /api/restock-predictions/_generate`

**Request Body**

| Field           | Tipe             | Wajib | Default | Keterangan                                                       |
| --------------- | ---------------- | :---: | ------- | ---------------------------------------------------------------- |
| `store_id`      | string           |   ✅  | —       | ID toko                                                          |
| `product_id`    | string           |   ➖  | —       | Kosong = prediksi seluruh produk toko                            |
| `forecast_date` | string (RFC3339) |   ➖  | besok   | Tanggal target (`YYYY-MM-DDT00:00:00Z`), harus besok/setelahnya  |
| `history_days`  | int              |   ➖  | 30      | Jumlah hari histori penjualan (minimal 30)                       |

> **Penting:** `forecast_date` **wajib** format RFC3339, contoh `"2030-01-01T00:00:00Z"`, agar unmarshalling `*time.Time` di Go tidak error.

```json
{
  "store_id": "1146d375-6219-46b8-bfda-e772a8260fd5",
  "product_id": "prod_9938",
  "forecast_date": "2030-01-01T00:00:00Z",
  "history_days": 30
}
```

**Response `200 OK`** — contoh saat histori belum cukup (produk masuk `skipped`):

```json
{
  "data": {
    "generated_count": 0,
    "skipped_count": 1,
    "items": [],
    "skipped": [
      {
        "product_id": "prod_9938",
        "product_name": "Gula Pasir Premium 1 kg",
        "reason": "riwayat_tidak_cukup"
      }
    ]
  },
  "message": "Berhasil membuat prediksi restock",
  "success": true
}
```

**Response `200 OK`** — contoh saat prediksi berhasil dibuat (`items` terisi):

```json
{
  "data": {
    "generated_count": 1,
    "skipped_count": 0,
    "items": [
      {
        "id": "restock_4241",
        "store_id": "1146d375-6219-46b8-bfda-e772a8260fd5",
        "product_id": "prod_9938",
        "product_name": "Gula Pasir Premium 1 kg",
        "unit": "kg",
        "forecast_date": "2030-01-01",
        "daily_sales": 10,
        "current_stock": 43,
        "recommended_restock_qty": 212,
        "created_at": 1787067747182
      }
    ],
    "skipped": []
  },
  "message": "Berhasil membuat prediksi restock",
  "success": true
}
```

**Keterangan field `items[]`**

| Field                     | Tipe   | Keterangan                                                        |
| ------------------------- | ------ | ----------------------------------------------------------------- |
| `id`                      | string | ID record prediksi                                                |
| `product_name`            | string | Nama produk                                                       |
| `unit`                    | string | Satuan produk                                                     |
| `forecast_date`           | string | Tanggal target prediksi (`YYYY-MM-DD`)                            |
| `daily_sales`             | int    | Prediksi penjualan harian (unit/hari) dari ML                     |
| `current_stock`           | int    | Stok produk saat ini                                             |
| `recommended_restock_qty` | int    | Jumlah unit yang direkomendasikan untuk di-restock                |
| `created_at`              | int64  | Timestamp prediksi dibuat (milli-epoch)                           |

**Keterangan field `skipped[]`**

| Field          | Tipe   | Keterangan                                                                              |
| -------------- | ------ | --------------------------------------------------------------------------------------- |
| `reason`       | string | `riwayat_tidak_cukup`, `kesalahan_ml`, atau `restock_tidak_diperlukan`                  |

### 6.2 List Predictions — `GET /api/restock-predictions`

**Query Params:** `store_id` (wajib).

Contoh: `GET /api/restock-predictions?store_id=1146d375-...`

**Response `200 OK`** (array `RestockPredictionResponse`; kosong bila belum ada data tersimpan):

```json
{
  "data": [],
  "message": "Berhasil menampilkan prediksi restock",
  "success": true
}
```

---

## 7. Referensi Error

| Status | Bentuk Response                                          | Kapan Terjadi                                        |
| :----: | ------------------------------------------------------- | ---------------------------------------------------- |
| `400`  | `{ "message": "...", "statusCode": 400 }`               | Body tidak valid / gagal validasi / query param wajib kosong |
| `401`  | `{ "errors": "Unauthorized" }`                          | Token tidak ada / invalid / kedaluwarsa              |
| `404`  | `{ "message": "...", "statusCode": 404 }`               | Data tidak ditemukan / kredensial login salah        |
| `409`  | `{ "message": "...", "statusCode": 409 }`               | Konflik (username sudah dipakai / user sudah punya toko) |
| `500`  | `{ "message": "...", "statusCode": 500 }`               | Kesalahan internal server                            |

Contoh error validasi (`POST /api/users` dengan body kurang lengkap):

```json
{
  "message": "Invalid request body",
  "statusCode": 400
}
```

Contoh error tanpa token (`GET /api/users/_current`):

```json
{
  "errors": "Unauthorized"
}
```
