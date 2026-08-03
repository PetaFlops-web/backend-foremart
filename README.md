# Smart Commerce Backend

Backend untuk aplikasi Smart Commerce, dikembangkan menggunakan arsitektur Modular Monolith dengan Golang (Fiber + GORM) dan berjalan di dalam container Docker.

## Sekilas Arsitektur (Modular Monolith)
Proyek ini mengadopsi pola **Modular Monolith** untuk mempermudah pemisahan tanggung jawab sekaligus menjaga kemudahan operasional pada fase awal pengembangan (tanpa perlu kompleksitas *microservices*). 

Prinsip utama yang diterapkan:
1. **Data Isolation (Isolasi Data)**: Setiap modul (seperti `auth`, `product`) mengelola tabel database-nya sendiri. Referensi relasi antar-modul menggunakan *plain ID* (string/UUID), **bukan** *Foreign Key* (FK) pada database.
2. **Client Interfaces**: Modul tidak boleh me-query langsung tabel milik modul lain (tidak ada operasi `JOIN` lintas modul). Komunikasi dan permintaan data lintas modul hanya dilakukan melalui *interface* publik (`<nama_modul>-client`).
3. **Standar Respons Global**: Seluruh endpoint API selalu menggunakan struktur JSON terpusat (`WebResponse[T]`) agar mudah di-parse dan seragam di sisi *client/frontend*.

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
   *Catatan: Konfigurasi default di `config.example.json` menggunakan `localhost`. Ubah menjadi `aic_mysql` agar container backend bisa berkomunikasi dengan container database.*

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
*Informasi teknis dan arsitektur lebih lanjut mengenai backend dapat dilihat pada dokumen di dalam folder `docs/`.*
