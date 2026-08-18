# Smart Commerce Backend

Backend untuk aplikasi Smart Commerce, dikembangkan menggunakan arsitektur Modular Monolith dengan Golang (Fiber + GORM) dan berjalan di dalam container Docker.

## Sekilas Arsitektur (Modular Monolith)

Proyek ini mengadopsi pola **Modular Monolith** untuk mempermudah pemisahan tanggung jawab sekaligus menjaga kemudahan operasional pada fase awal pengembangan (tanpa perlu kompleksitas _microservices_).

Prinsip utama yang diterapkan:

1. **Data Isolation (Isolasi Data)**: Setiap modul (seperti `auth`, `product`) mengelola tabel database-nya sendiri. Referensi relasi antar-modul menggunakan _plain ID_ (string/UUID), **bukan** _Foreign Key_ (FK) pada database.
2. **Client Interfaces**: Modul tidak boleh me-query langsung tabel milik modul lain (tidak ada operasi `JOIN` lintas modul). Komunikasi dan permintaan data lintas modul hanya dilakukan melalui _interface_ publik (`<nama_modul>-client`).
3. **Standar Respons Global**: Seluruh endpoint API selalu menggunakan struktur JSON terpusat (`WebResponse[T]`) agar mudah di-parse dan seragam di sisi _client/frontend_.

## Persyaratan

- Docker & Docker Compose
- Git

## Cara Setup Lokal

Untuk menjalankan backend ini secara lokal di mesin Anda, ikuti langkah-langkah berikut:

1. **Clone repositori** (jika belum)

   ```bash
   git clone https://github.com/PetaFlops-web/backend-shop-smbk.git
   cd backend-shop-smbk
   ```

2. **Siapkan konfigurasi `.env`**
   Salin template `.env.example` menjadi `.env` dan sesuaikan nilainya jika perlu.

   ```bash
   cp .env.example .env
   ```

3. **Siapkan konfigurasi `config.json`**
   Salin template `config.example.json` menjadi `config.json`. Anda bisa menggunakan nilai default atau menyesuaikannya (terutama bagian `database` jika menjalankan MySQL secara terpisah, namun default `config.example.json` disiapkan untuk digunakan dengan environment yang sama bila dijalankan di luar docker, jika menggunakan docker-compose pastikan host db mengarah ke `aic_mysql`).

   Untuk integrasi dengan Docker Compose, atur `config.json` pada bagian database host ke `aic_mysql`:

   ```json
   "database": {
     "username": "your_user",       // sesuaikan dengan MYSQL_USER di .env
     "password": "your_password",   // sesuaikan dengan MYSQL_PASSWORD di .env
     "host": "aic_mysql",
     "port": 3306,
     "name": "database_name",       // sesuaikan dengan MYSQL_DATABASE di .env
     // ...
   }
   ```

   _Catatan: Konfigurasi default di `config.example.json` menggunakan `localhost`. Ubah menjadi `aic_mysql` agar container backend bisa berkomunikasi dengan container database._

4. **Jalankan dengan Docker Compose**
   Gunakan perintah berikut untuk melakukan build dan menyalakan container:
   ```bash
   docker compose up --build -d
   ```
   Tunggu beberapa saat hingga container database MySQL siap (ready for connections) dan backend berjalan.

## Informasi Endpoint

### Base URL

Bila dijalankan secara lokal dengan konfigurasi default, Base URL API adalah:
`http://127.0.0.1:8080`

### Dokumentasi API (Swagger)

Seluruh daftar endpoint, parameter yang dibutuhkan (termasuk Auth header), serta struktur data request/response dapat dilihat dan diuji coba secara interaktif melalui Swagger UI.

Akses Swagger UI melalui browser di:
👉 **[http://127.0.0.1:8080/swagger/](http://127.0.0.1:8080/swagger/)**

### Standar Response API

Setiap endpoint API selalu mengembalikan format JSON standar berikut:

**Response Sukses:**

```json
{
  "data": { ... },
  "message": "Pesan sukses opsional",
  "success": true
}
```

**Response dengan Pagination:**

```json
{
  "data": [ ... ],
  "message": "Pesan sukses opsional",
  "success": true,
  "paging": {
    "page": 1,
    "size": 10,
    "total_item": 25,
    "total_page": 3
  }
}
```

**Response Error:**

```json
{
  "data": null,
  "message": "Pesan error yang jelas",
  "success": false
}
```

### Autentikasi

Endpoint yang diproteksi memerlukan token JWT.
Kirimkan token pada header request:

```http
Authorization: Bearer <token_anda_dari_login>
```

---

## Endpoint: Prediksi Restock (Restock Prediction)

Endpoint ini menghitung rekomendasi jumlah restock untuk produk di suatu toko, berdasarkan prediksi penjualan harian dari layanan ML (`/predict-inventory`).

### 1. Generate Prediksi Restock

Menghitung dan menyimpan rekomendasi restock. Bisa untuk satu produk atau seluruh produk dalam satu toko.

```
POST /api/restock-predictions/_generate
Authorization: Bearer <token>
Content-Type: application/json
```

#### Request Body

| Field           | Tipe             | Wajib    | Default | Keterangan                                                                    |
| --------------- | ---------------- | -------- | ------- | ----------------------------------------------------------------------------- |
| `store_id`      | string           | Required | —       | ID toko yang ingin diprediksi                                                 |
| `product_id`    | string           | Nullable | —       | ID produk. Kosong = prediksi **seluruh** produk toko                          |
| `forecast_date` | string (RFC3339) | Required | besok   | Tanggal target prediksi (`YYYY-MM-DDT00:00:00Z`). Harus besok atau setelahnya |
| `history_days`  | int              | Required | 30      | Jumlah hari histori penjualan (minimal 30)                                    |

> [!IMPORTANT]
> **Penting:** Pastikan format date menggunakan RFC3339.
> **Contoh:** `"2026-09-10T00:00:00Z"`.
> Karena backend menggunakan tipe data `*time.Time` dalam Go, penggunaan format **RFC3339** sangat krusial untuk memastikan proses _unmarshalling_ JSON berjalan tanpa error dan menghindari inkonsistensi zona waktu.

Contoh (prediksi satu produk):

```json
{
  "store_id": "2ce7f883-4666-432f-aa65-ccd97e5c4686",
  "product_id": "prod_1003",
  "forecast_date": "2026-09-10T00:00:00Z",
  "history_days": 30
}
```

Contoh (prediksi seluruh produk toko):

```json
{
  "store_id": "2ce7f883-4666-432f-aa65-ccd97e5c4686",
  "forecast_date": "2026-09-10T00:00:00Z"
}
```

#### Response Sukses

```json
{
  "data": {
    "generated_count": 1,
    "skipped_count": 0,
    "items": [
      {
        "id": "restock_4241",
        "store_id": "2ce7f883-4666-432f-aa65-ccd97e5c4686",
        "product_id": "prod_1003",
        "product_name": "Gula Pasir 1 kg",
        "unit": "kg",
        "forecast_date": "2026-09-10",
        "daily_sales": 10,
        "current_stock": 18,
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

**Keterangan field `items`:**

| Field                     | Tipe   | Keterangan                                                                                                                |
| ------------------------- | ------ | ------------------------------------------------------------------------------------------------------------------------- |
| `id`                      | string | ID record prediksi                                                                                                        |
| `store_id`                | string | ID toko                                                                                                                   |
| `product_id`              | string | ID produk                                                                                                                 |
| `product_name`            | string | Nama produk                                                                                                               |
| `unit`                    | string | Satuan produk (kg, pcs, box, dst.)                                                                                        |
| `forecast_date`           | string | Tanggal target prediksi (`YYYY-MM-DD`)                                                                                    |
| `daily_sales`             | int    | Prediksi penjualan harian (unit/hari) dari ML                                                                             |
| `current_stock`           | int    | Stok produk saat ini                                                                                                      |
| `recommended_restock_qty` | int    | **Jumlah unit yang direkomendasikan untuk di-restock agar stok mencukupi hingga target tanggal prediksi yang ditentukan** |
| `created_at`              | int64  | Timestamp prediksi dibuat (milli-epoch)                                                                                   |

**Keterangan field `skipped`:**

| Field          | Tipe   | Keterangan                                                                              |
| -------------- | ------ | --------------------------------------------------------------------------------------- |
| `product_id`   | string | ID produk yang dilewati                                                                 |
| `product_name` | string | Nama produk yang dilewati                                                               |
| `reason`       | string | Alasan dilewati: `riwayat_tidak_cukup`, `kesalahan_ml`, atau `restock_tidak_diperlukan` |

Jika hasilnya `0` atau negatif (stok masih cukup), produk masuk ke `skipped` dengan reason `restock_tidak_diperlukan`.

### 2. Daftar Prediksi Restock Tersimpan

Mengembalikan seluruh prediksi restock yang sudah tersimpan untuk suatu toko.

```
GET /api/restock-predictions?store_id=<store_id>
Authorization: Bearer <token>
```

**Query Param:**

| Field      | Tipe   | Wajib    | Keterangan |
| ---------- | ------ | -------- | ---------- |
| `store_id` | string | Required | ID toko    |

**Response:** array berisi objek `RestockPredictionResponse` yang sama dengan field `items` di atas.

---

_Informasi teknis dan arsitektur lebih lanjut mengenai backend dapat dilihat pada dokumen di dalam folder `docs/`._
