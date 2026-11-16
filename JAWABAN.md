# Jawaban Ujian CPMK 3 – Pemrograman Web II
Nama  : Hana Nur Fathiyyah
NIM   : H1H024017
Paket : 7 – Seminar dan Pendaftaran Peserta
Video : `<LINK VIDEO YOUTUBE>`

## Pertanyaan 1
**Bagaimana go-playground/validator membaca tag `validate:"required,email"` pada struct (reflection) dan bagaimana mengubah pesan error menjadi bahasa Indonesia?**

Validator memakai paket `reflect`. Saat `validate.Struct(req)` dipanggil, ia mengambil tipe struct, mengulang setiap field, membaca *struct tag* `validate` (string `"required,email,max=150"`), memecahnya berdasarkan koma, lalu menjalankan satu per satu fungsi aturan yang cocok (`required`, `email`, `max`, ...) terhadap nilai field. Setiap aturan yang gagal menjadi satu `FieldError` yang berisi nama field (`StructField()`), nama aturan (`Tag()`), dan parameternya (`Param()`, misal `150`); semuanya dikumpulkan sebagai `validator.ValidationErrors`.

Pesan bawaan validator berbahasa Inggris (`Key: 'X.Email' Error:Field validation ... failed on the 'email' tag`), jadi saya mengubahnya sendiri di dua langkah:

1. `Validate()` melakukan reflection kedua: dari `FieldError.StructField()` saya cari field-nya dengan `reflect.TypeOf(req).FieldByName(...)`, lalu membaca tag `form` (nama input HTML, jadi kunci map error) dan tag `label` (nama tampilan, misal "Email") — `app/requests/validator.go` baris 32–44.
2. Fungsi `message(label, fe)` memetakan `fe.Tag()` ke kalimat Indonesia lewat `switch` — `app/requests/validator.go` baris 50–75. Contoh: `required` → "Email wajib diisi.", `email` → "Email harus berupa email yang valid.".

Tag yang dipakai ada di `app/requests/registration_request.go` baris 7–11, misalnya `Email ... validate:"required,email,max=150"` (baris 9). Hasilnya berupa `map[string]string` yang di template ditampilkan di bawah input lewat `{{with index .Errors "email"}}` di `resources/views/registrations/_form.html`.

## Pertanyaan 2
**Jelaskan konsep layout dan partial di gofiber/template/html: peran `{{embed}}`, `{{template "partials/navbar" .}}`, serta perbedaan `.` dan `$` di dalam `{{range}}`.**

- **Layout & `{{embed}}`**: `ViewsLayout: "layouts/app"` (`bootstrap/app.go` baris 39) membuat setiap `c.Render("seminars/index", data)` dirender dua tahap: halaman dirender dulu, lalu hasilnya disisipkan ke layout pada tempat `{{embed}}` (`resources/views/layouts/app.html` baris 16). Jadi kerangka HTML (head, navbar, footer, script) hanya ditulis sekali — setara `@extends` + `@yield('content')` di Blade.
- **Partial `{{template "partials/navbar" .}}`** (`layouts/app.html` baris 12, juga baris 15 dan 23): memanggil template lain berdasarkan nama = path relatif tanpa `.html`. Tanda titik di akhir adalah data yang dikirim ke partial; tanpa itu partial menerima `nil` dan `.AppName`, `.CurrentPath` tidak tersedia. Setara `@include`. Di `seminars/create.html` partial `seminars/_form` dipanggil dengan `.` supaya bisa membaca `.Form`, `.Errors`, `.CsrfToken`.
- **`.` vs `$`**: `.` adalah *konteks saat ini* dan berubah di dalam `range`/`with` (di `{{range .Pages}}` titik menjadi satu angka halaman). `$` adalah data akar yang diberikan saat template dijalankan dan selalu bisa diakses. Contoh di proyek ini:
  - `resources/views/partials/pagination.html` baris 11–12: `{{if eq . $.Page}}` (`.` = nomor halaman, `$.Page` = halaman aktif) dan `{{$.Link .}}` (memanggil method `Link` milik Pagination dengan nomor halaman).
  - `resources/views/registrations/_form.html` baris 8–9: di dalam `{{range .Seminars}}`, `.ID` adalah id seminar yang sedang diulang, sedangkan `$.Form.SeminarID` adalah pilihan form di luar `range`.
  - `resources/views/registrations/index.html` baris 15–16: `{{if eq .ID $.SeminarID}}selected{{end}}`.
  - Catatan: pada partial yang dipanggil dengan `.Pagination`, `$` berarti struct Pagination itu sendiri (data yang dikirim ke pemanggilan template), bukan data halaman.

## Pertanyaan 3
**Mengapa nilai dari `c.Params()` / `c.Query()` tidak boleh disimpan dan dipakai setelah handler selesai (zero-allocation fasthttp)? Bagaimana cara aman menyimpannya?**

Fiber berjalan di atas fasthttp yang menghindari alokasi memori: string yang dikembalikan `c.Params()`, `c.Query()`, `c.FormValue()` tidak disalin, melainkan menunjuk langsung ke *buffer* request yang dimiliki fasthttp. Begitu handler selesai, buffer itu dikembalikan ke pool dan dipakai ulang untuk request berikutnya. Kalau string tadi masih disimpan (variabel global, map/cache, dikirim ke goroutine yang berjalan lebih lama dari handler), isinya bisa berubah menjadi data request lain atau terjadi *data race* — bug yang sulit dilacak.

Cara aman menyimpannya:
1. Salin string: `strings.Clone(c.Query("q"))` (atau `utils.CopyString`).
2. Ubah ke tipe nilai baru: angka, bool, struct — hasil konversi adalah data baru. Di proyek ini `helpers.ParseID` memakai `strconv.ParseUint(c.Params("id"), ...)` sehingga yang dipakai selanjutnya adalah angka `uint`, bukan string asli (`app/helpers/view.go` baris 43–44).
3. Set `Immutable: true` di `fiber.Config` agar semua nilai disalin otomatis (konsekuensi: sedikit lebih boros memori).

Pada kodeku nilai seperti `search := strings.TrimSpace(c.Query("q"))` (`app/controllers/seminar_controller.go` baris 33 dan `registration_controller.go` baris 25) hanya dipakai di dalam handler yang sama (dipakai untuk query dan dirender sebelum `return`), sehingga aman. Pesan flash seperti `"Kode tiket \""+code+"\" tidak ditemukan."` (`app/controllers/checkin_controller.go` baris 54–84) dibentuk dengan penggabungan string yang membuat string baru. Bila nilai itu harus hidup lebih lama dari handler, saya akan `strings.Clone` dahulu.

## Status tantangan
- T1: selesai — unique `seminar_id + email` (`app/controllers/registration_controller.go` baris 194–225, unique index di `app/models/registration.go` baris 9–15); tolak bila kuota penuh / tanggal lewat (baris 202–214); kode tiket acak 8 karakter unik (`app/helpers/ticket.go`, `registration_controller.go` baris 227–238).
- T2: selesai — subquery COUNT (`seminar_controller.go` baris 20–22 dan 51), progress bar + badge di `resources/views/seminars/index.html`, status di `app/models/seminar.go` baris 30; halaman detail seminar memuat daftar peserta (`seminars/show.html`).
- T3: selesai — check-in (`app/controllers/checkin_controller.go` baris 54–84), rekap kehadiran satu query agregat (baris 35–51), export CSV (`app/controllers/seminar_controller.go` baris 185–229).
