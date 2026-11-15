package helpers

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Rupiah memformat angka menjadi "Rp 1.250.000".
func Rupiah(n int64) string {
	s := fmt.Sprintf("%d", n)
	var out []byte
	for i, ch := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, '.')
		}
		out = append(out, ch)
	}
	return "Rp " + string(out)
}

// FeeLabel: biaya 0 ditampilkan "Gratis".
func FeeLabel(n int64) string {
	if n == 0 {
		return "Gratis"
	}
	return Rupiah(n)
}

// Tanggal memformat waktu menjadi "12-11-2026".
func Tanggal(t time.Time) string {
	return t.Format("02-01-2006")
}

// TanggalWaktu memformat waktu menjadi "12-11-2026 14:05".
func TanggalWaktu(t time.Time) string {
	return t.Format("02-01-2006 15:04")
}

// DateTimeInput memformat waktu untuk <input type="datetime-local">.
func DateTimeInput(t time.Time) string {
	return t.Format("2006-01-02T15:04")
}

// StatusClass memberi kelas badge Bootstrap sesuai status seminar.
func StatusClass(status string) string {
	switch status {
	case "Penuh":
		return "text-bg-danger"
	case "Tersedia":
		return "text-bg-success"
	default:
		return "text-bg-secondary"
	}
}

// BarClass memberi warna progress bar: hijau, kuning (>=80%), merah (penuh).
func BarClass(percent int) string {
	switch {
	case percent >= 100:
		return "bg-danger"
	case percent >= 80:
		return "bg-warning"
	default:
		return "bg-success"
	}
}

// CSVSafe mencegah CSV injection: sel yang diawali = + - @ dianggap rumus oleh Excel,
// jadi diberi tanda petik di depan agar dibaca sebagai teks biasa.
func CSVSafe(s string) string {
	if s != "" && strings.ContainsRune("=+-@\t\r", rune(s[0])) {
		return "'" + s
	}
	return s
}

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

// Slug mengubah "Seminar Go & Fiber" menjadi "seminar-go-fiber" (maks. 50 karakter).
func Slug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonAlphaNum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 50 {
		s = strings.Trim(s[:50], "-")
	}
	if s == "" {
		return "seminar"
	}
	return s
}
