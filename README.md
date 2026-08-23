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

| Field           | Tipe             | Wajib    | Default | Keterangan                                                                                  |
| --------------- | ---------------- | -------- | ------- | ------------------------------------------------------------------------------------------- |
| `store_id`      | string           | Required | —       | ID toko yang ingin diprediksi                                                               |
| `product_id`    | string           | Nullable | —       | ID produk. Kosong = prediksi **seluruh** produk toko                                        |
| `forecast_date` | string (RFC3339) | Required | besok   | Tanggal target yang ingin di prediksi (`YYYY-MM-DDT00:00:00Z`). Harus besok atau setelahnya |
| `history_days`  | int              | Required | 30      | Jumlah hari histori penjualan (minimal 30)                                                  |

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

## Endpoint: Prediksi Pembelian Ulang (Survival Prediction)

Endpoint ini memprediksi kapan seorang customer akan membeli ulang sebuah produk, berdasarkan fitur survival yang dihitung dari riwayat transaksi, lalu memanggil layanan ML (`/predict-survival`).

### 1. Prediksi Pembelian Ulang

Menghitung fitur survival dari riwayat pembelian customer terhadap suatu produk, lalu mengembalikan prediksi tanggal pembelian ulang beserta probabilitasnya.

```
POST /api/predict-survival
Authorization: Bearer <token>
Content-Type: application/json
```

#### Request Body

| Field         | Tipe   | Wajib    | Keterangan                                                                                          |
| ------------- | ------ | -------- | --------------------------------------------------------------------------------------------------- |
| `store_id`    | string | Required | ID toko (UUID) pemilik produk                                                                       |
| `customer_id` | int    | Required | ID customer (minimal 1) yang akan diprediksi pembelian ulangnya                                     |
| `product_id`  | string | Required | ID produk (format `prod_XXXX`) yang diprediksi                                                      |

Contoh:

```json
{
  "store_id": "2ce7f883-4666-432f-aa65-ccd97e5c4686",
  "customer_id": 42,
  "product_id": "prod_1003"
}
```

#### Response Sukses

```json
{
  "data": {
    "customer_id": 42,
    "stock_code": "Gula Pasir 1 kg",
    "predicted_restock_date": "2026-08-25",
    "pred_days_left": 3,
    "pred_median_survival_days": 12.5,
    "days_since_last_buy": 9,
    "prob_buy_within_7d": 0.32,
    "prob_buy_within_14d": 0.61,
    "prob_buy_within_30d": 0.88,
    "partial_hazard": 1.24
  },
  "message": "Berhasil membuat prediksi pembelian ulang",
  "success": true
}
```

**Keterangan field `data`:**

| Field                       | Tipe   | Keterangan                                                                                       |
| --------------------------- | ------ | ------------------------------------------------------------------------------------------------ |
| `customer_id`               | int    | ID customer                                                                                      |
| `stock_code`                | string | Nama produk (diisi dari `product_name`, bukan kode produk)                                        |
| `predicted_restock_date`    | string | Estimasi tanggal pembelian ulang (`YYYY-MM-DD`) dari layanan ML                                  |
| `pred_days_left`            | int    | Sisa hari menuju tanggal prediksi pembelian ulang                                                |
| `pred_median_survival_days` | float  | Median hari survival (jarak antar pembelian)                                                     |
| `days_since_last_buy`       | int    | Hari sejak pembelian terakhir                                                                     |
| `prob_buy_within_7d`        | float  | Probabilitas membeli ulang dalam 7 hari (0.0 – 1.0)                                              |
| `prob_buy_within_14d`       | float  | Probabilitas membeli ulang dalam 14 hari (0.0 – 1.0)                                             |
| `prob_buy_within_30d`       | float  | Probabilitas membeli ulang dalam 30 hari (0.0 – 1.0)                                             |
| `partial_hazard`            | float  | Nilai partial hazard dari model survival                                                          |

> [!IMPORTANT]
> **Penting:** Prediksi hanya berhasil jika customer tersebut **sudah pernah** membeli produk tersebut di toko terkait (ada minimal satu record transaksi). Jika belum ada riwayat, endpoint mengembalikan `404` dengan pesan `Tidak ada riwayat pembelian produk ini`.

#### Response Error

| Kode | `message`                            | Trigger                                                          |
| ---- | ------------------------------------ | ---------------------------------------------------------------- |
| 400  | `Format data request tidak valid`    | Body JSON tidak valid                                            |
| 400  | `Data request tidak valid`           | Gagal validasi (field wajib kosong / `customer_id` < 1)          |
| 401  | `Unauthorized`                       | Token JWT hilang atau tidak valid                                |
| 403  | `Produk tidak termasuk dalam toko Anda` | `store_id` tidak sesuai dengan pemilik produk                   |
| 404  | `Produk tidak ditemukan`             | `product_id` tidak ada                                           |
| 404  | `Tidak ada riwayat pembelian produk ini` | Customer belum pernah membeli produk tsb di toko tsb          |
| 500  | `Gagal mengambil riwayat pembelian`  | Error saat query riwayat pembelian                               |
| 500  | `Gagal memanggil layanan prediksi`   | Layanan ML gagal dipanggil / tidak reachable                     |

## Endpoint: Notifikasi Pengingat Pembelian Ulang (Reorder Reminder)

Backend mengirim notifikasi WhatsApp ke customer ketika prediksi pembelian ulang (`predict-survival`) menunjukkan mereka akan segera membeli kembali sebuah produk. Notifikasi dijalankan otomatis oleh scheduler (cron) setiap hari, dan bisa disisipi promo untuk pelanggan yang sudah sering membeli.

### Setup Gateway WhatsApp (Fonnte)

Notifikasi memakai gateway **Fonnte** (unofficial WhatsApp API gateway Indonesia). Langkah-langkah setup:

#### 1. Daftar & Login

1. Buka [md.fonnte.com](https://md.fonnte.com/new/login.php) (atau daftar di [md.fonnte.com/new/register.php](https://md.fonnte.com/new/register.php)).
2. Buat akun (gratis, dapat kuota awal untuk testing).

#### 2. Tambah Device (hubungkan nomor WhatsApp)

1. Di dashboard, buka menu **Device**.
2. Klik **Tambah Device / Add Device**.
3. Akan muncul **QR code** — scan pakai WhatsApp di HP (mirip WhatsApp Web).
4. Setelah scan, device (nomor WhatsApp) terhubung ke akun Fonnte.

> Satu paket kuota terikat ke satu device. Nomor WhatsApp yang dipakai kirim notifikasi adalah nomor device yang terhubung ini.

#### 3. Ambil Token (API Key)

1. Buka menu **Device**.
2. Di daftar device, cari token (API key) milik device kamu.
3. **Klik token** → otomatis tersalin ke clipboard.

> **Penting:** Token Fonnte bersifat rahasia. Siapa pun yang memegang token ini bisa mengirim pesan atas nama WhatsApp kamu. Jangan commit token ke git.

#### 4. Isi Konfigurasi

Tambahkan blok berikut di `config.json`:

```json
{
  "fonnte": {
    "token": "TOKEN_FONNTE_ANDA",
    "target": ""
  },
  "notification": {
    "schedule": "0 8 * * *"
  }
}
```

| Field                        | Keterangan                                                                                          |
| ---------------------------- | --------------------------------------------------------------------------------------------------- |
| `fonnte.token`               | Token API Fonnte dari menu **Device**. Kosong = fallback ke log-only (pesan tidak benar-benar terkirim). |
| `fonnte.target`              | Nomor default (opsional). Biarkan `""` — nomor tujuan diambil otomatis dari `customers.phone`.       |
| `notification.schedule`      | Jadwal cron (format cron standar). Default `0 8 * * *` = setiap hari pukul 08:00 waktu lokal.       |

> Nomor WhatsApp tujuan harus berformat **E.164** (awalan `62`, tanpa `+`, tanpa `0` di depan). Contoh: `6281234567890`. Pastikan `customers.phone` di database memakai format ini.

### Cara Kerja

1. Scheduler berjalan tiap hari sesuai `notification.schedule`.
2. Untuk setiap toko, diiterasi semua pasangan (customer × produk).
3. Pasangan yang `pred_days_left ≤ 3` hari (menjelang prediksi beli ulang) memicu pengiriman.
4. Rule promo berdasarkan jumlah pembelian (`purchase_number`):
   - `≥ 5x` beli → diskon **30%**
   - `≥ 3x` beli → diskon **20%**
   - di bawah itu → pesan pengingat biasa (tanpa diskon)
5. Setiap kirim dicatat ke `notification_logs` dengan kunci dedup `(customer_id, product_id, period)` agar tidak terkirim dua kali di hari yang sama.

Contoh pesan terkirim:

> Halo Budi, persediaan Gula Pasir 1 kg Anda mungkin sudah mulai habis. Anda dapat membeli kembali Gula Pasir 1 kg dengan potongan harga 20% karena Anda sudah berbelanja beberapa kali. Prediksi waktu pembelian ulang Anda: 2026-08-25. Sampai jumpa!

### 2. Trigger Pengiriman (Manual)

Menjalankan proses pengiriman notifikasi secara manual untuk seluruh toko — memicu alur yang sama persis dengan cron job. Berguna untuk testing.

```
POST /api/notifications/_send
Authorization: Bearer <token>
```

**Tanpa request body.** Response berisi jumlah notifikasi yang terkirim:

```json
{
  "data": 1,
  "message": "Berhasil menjalankan pengiriman notifikasi",
  "success": true
}
```

| Field | Tipe | Keterangan |
| ----- | ---- | ---------- |
| `data` | int  | Jumlah notifikasi yang terkirim pada run ini |

### 3. Daftar Log Notifikasi

Mengembalikan log notifikasi yang sudah terkirim untuk suatu toko.

```
GET /api/notifications?store_id=<store_id>
Authorization: Bearer <token>
```

**Query Param:**

| Field      | Tipe   | Wajib    | Keterangan |
| ---------- | ------ | -------- | ---------- |
| `store_id` | string | Required | ID toko    |

**Response Sukses:**

```json
{
  "data": [
    {
      "id": "notif_4241",
      "store_id": "2ce7f883-4666-432f-aa65-ccd97e5c4686",
      "customer_id": 42,
      "product_id": "prod_1003",
      "channel": "whatsapp",
      "message": "Halo Budi, persediaan Gula Pasir 1 kg Anda ...",
      "predicted_restock_date": "2026-08-25",
      "rule_triggered": "REPEAT_3X",
      "status": "sent",
      "period": "2026-08-22",
      "created_at": 1787067747182
    }
  ],
  "message": "Berhasil mengambil log notifikasi",
  "success": true
}
```

**Keterangan field `data`:**

| Field                    | Tipe   | Keterangan                                                        |
| ------------------------ | ------ | ----------------------------------------------------------------- |
| `id`                     | string | ID log notifikasi                                                 |
| `store_id`               | string | ID toko                                                           |
| `customer_id`            | int    | ID customer                                                       |
| `product_id`             | string | ID produk                                                         |
| `channel`                | string | Channel notifikasi (`whatsapp`)                                   |
| `message`                | string | Isi pesan yang dikirim                                            |
| `predicted_restock_date` | string | Snapshot tanggal prediksi (`YYYY-MM-DD`)                          |
| `rule_triggered`         | string | `REPEAT_3X` (ada promo) atau `REMINDER` (tanpa promo)             |
| `status`                 | string | `sent` (terkirim) atau `failed` (gagal kirim)                     |
| `period`                 | string | Periode dedup (`YYYY-MM-DD`)                                      |
| `created_at`             | int64  | Timestamp log dibuat (milli-epoch)                                |

---

_Informasi teknis dan arsitektur lebih lanjut mengenai backend dapat dilihat pada dokumen di dalam folder `docs/`._
