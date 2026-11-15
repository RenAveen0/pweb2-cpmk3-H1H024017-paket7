package models

import "time"

// Seminar = tabel "seminars". Satu seminar memiliki banyak pendaftaran (hasMany).
type Seminar struct {
	ID        uint      `gorm:"primaryKey"`
	Title     string    `gorm:"size:150;not null"`
	Speaker   string    `gorm:"size:100;not null"`
	EventDate time.Time `gorm:"not null;index"`
	Location  string    `gorm:"size:150;not null"`
	Quota     int       `gorm:"not null;default:0"`
	Fee       int64     `gorm:"not null;default:0"` // 0 = gratis
	Poster    *string   `gorm:"size:255"`           // nullable

	// Hapus seminar => pendaftarnya ikut terhapus (CASCADE), karena data
	// pendaftaran tidak berarti tanpa seminarnya.
	Registrations []Registration `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	CreatedAt time.Time
	UpdatedAt time.Time

	// Kolom virtual hasil subquery COUNT (mirip withCount('registrations') di Laravel).
	// "->" = hanya dibaca, "-:migration" = tidak dibuat sebagai kolom tabel.
	RegistrationsCount int64 `gorm:"->;-:migration"`
}

// Status menghasilkan label Selesai / Penuh / Tersedia.
// Membutuhkan RegistrationsCount (diisi lewat subquery COUNT).
func (s Seminar) Status() string {
	if s.EventDate.Before(time.Now()) {
		return "Selesai"
	}
	if s.RegistrationsCount >= int64(s.Quota) {
		return "Penuh"
	}
	return "Tersedia"
}

// Percent = persentase kuota terisi (0-100) untuk progress bar.
func (s Seminar) Percent() int {
	if s.Quota <= 0 {
		return 0
	}
	p := int(s.RegistrationsCount * 100 / int64(s.Quota))
	if p > 100 {
		p = 100
	}
	return p
}

// Remaining = sisa kursi yang masih tersedia.
func (s Seminar) Remaining() int {
	left := s.Quota - int(s.RegistrationsCount)
	if left < 0 {
		return 0
	}
	return left
}
