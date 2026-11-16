# Seminar & Pendaftaran Peserta — Ujian CPMK 3 Pemrograman Web II

| | |
|---|---|
| **Nama** | Hana Nur Fathiyyah |
| **NIM** | H1H024017 |
| **Paket** | 7 — Seminar dan Pendaftaran Peserta |
| **Video YouTube** | `<LINK VIDEO>` |

Aplikasi web fullstack MVC (struktur ala Laravel) untuk himpunan mahasiswa yang menyelenggarakan seminar dan mengelola pendaftaran pesertanya.

**Stack:** Go ≥ 1.26 · Go Fiber v3 · `gofiber/template/html/v2` · GORM + MySQL · Bootstrap 5.3 (CDN) · JavaScript murni (hanya modal hapus).

## Cara instalasi

Prasyarat: Go ≥ 1.26, MySQL 8 / MariaDB yang sudah berjalan, Git.

```bash
# 1. Clone repository
git clone https://github.com/<username>/pweb2-cpmk3-H1H024017-paket7.git
cd pweb2-cpmk3-H1H024017-paket7

# 2. Buat database (jalankan di phpMyAdmin / MySQL client)
#    CREATE DATABASE seminar_paket7 CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 3. Salin konfigurasi lalu sesuaikan DB_USERNAME / DB_PASSWORD
cp .env.example .env          # Windows: copy .env.example .env

# 4. Unduh dependency, buat tabel, isi data contoh, jalankan
go mod tidy
go run . migrate:fresh
go run . seed
go run .
```

Buka <http://localhost:3000>.

## Struktur folder

```
main.go                  entry point + perintah CLI (migrate, migrate:fresh, seed)
bootstrap/app.go         setup Fiber, view engine, middleware (method override, recover, logger, CSRF), 404
config/                  config.go (.env), database.go (GORM MySQL)
app/models/              Seminar, Registration (+ relasi)
app/controllers/         home, seminar, registration, checkin
app/requests/            struct request + tag validate (pesan error bahasa Indonesia)
app/middleware/          method_override.go
app/helpers/             view, pagination, format, ticket
database/migrations/     AutoMigrate & Fresh
database/seeders/        20 seminar + pendaftar
routes/web.go            daftar route
resources/views/         layouts/, partials/, home/, seminars/, registrations/, checkin/, errors/
public/                  css/, js/
```

## Skema data

- `seminars` — judul, pembicara, tanggal, lokasi, kuota, biaya (0 = gratis), poster (nullable) — **hasMany** `registrations`.
- `registrations` — seminar_id, nama, email, nim (nullable), instansi, kode_tiket (unik), hadir (boolean) — **belongsTo** `seminar`.
- Foreign key `registrations.seminar_id` → `seminars.id` dengan `ON DELETE CASCADE` (pendaftar tidak bermakna tanpa seminarnya) dan `ON UPDATE CASCADE`.
- Unique gabungan `seminar_id + email`; unique `kode_tiket`.

## Daftar fitur

**Standar minimal CRUD**
- CRUD seminar (+ pencarian judul/pembicara) dan CRUD pendaftar (+ dropdown seminar, pencarian nama/email/kode tiket, filter seminar).
- Resource route + method override (`_method`) untuk PUT/DELETE.
- Satu partial form (`_form.html`) dipakai bersama oleh create dan edit.
- Validasi bahasa Indonesia di bawah input (`is-invalid` / `invalid-feedback`), old input tidak hilang.
- Pagination 10 data, parameter pencarian/filter ikut terbawa saat pindah halaman.
- `Preload` / subquery `COUNT` (tanpa N+1).
- Modal konfirmasi hapus Bootstrap + flash message setelah tambah/ubah/hapus.
- Layout (`{{embed}}`) + partial navbar/flash/pagination/delete-modal.
- CSRF aktif di semua form, output template ter-escape, halaman 404 kustom.
- `go run . migrate:fresh` dan `go run . seed` (20 seminar + ±100 pendaftar).

## Status tantangan

| Tantangan | Status | Ringkasan |
|---|---|---|
| **T1** (8) | Selesai | Unique `seminar_id + email` (cek di aplikasi + unique index DB); pendaftaran ditolak bila kuota penuh atau tanggal seminar sudah lewat; kode tiket acak 8 karakter unik (`crypto/rand`). Cek kuota dilakukan dalam transaksi dengan `SELECT ... FOR UPDATE`. |
| **T2** (10) | Selesai | Tabel seminar menampilkan `pendaftar / kuota` dengan progress bar Bootstrap dari subquery `COUNT` (1 query), plus badge Penuh / Tersedia / Selesai. Halaman detail seminar menampilkan daftar peserta. |
| **T3** (12) | Selesai | Halaman `/checkin`: input kode tiket → tandai hadir (tidak bisa dua kali, pesan peringatan bila sudah hadir), rekap kehadiran per seminar (satu query agregat), dan export CSV daftar hadir per seminar. |

## Screenshot

Disimpan di folder [`screenshots/`](screenshots/):

1. `01-dashboard.png` — Dashboard
2. `02-seminar-index.png` — Daftar seminar + progress bar + badge status
3. `03-validasi-gagal.png` — Validasi gagal + old input
4. `04-modal-hapus.png` — Modal konfirmasi hapus
5. `05-checkin.png` — Halaman check-in + rekap kehadiran

![Dashboard](screenshots/01-dashboard.png)
![Daftar seminar](screenshots/02-seminar-index.png)
![Validasi gagal](screenshots/03-validasi-gagal.png)
![Modal hapus](screenshots/04-modal-hapus.png)
![Check-in](screenshots/05-checkin.png)

## Video penjelasan workflow coding

`<LINK VIDEO YOUTUBE>`
