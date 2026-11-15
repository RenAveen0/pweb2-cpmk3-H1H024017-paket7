package models

import "time"

// Registration = tabel "registrations". Setiap pendaftaran milik satu seminar (belongsTo).
type Registration struct {
	ID uint `gorm:"primaryKey"`

	// uniqueIndex dengan nama yang sama pada dua field = unique GABUNGAN
	// (seminar_id + email): satu email hanya boleh mendaftar sekali per seminar.
	SeminarID uint    `gorm:"not null;uniqueIndex:idx_seminar_email"`
	Seminar   Seminar // relasi belongsTo, diisi lewat Preload("Seminar")

	Name        string  `gorm:"size:100;not null"`
	Email       string  `gorm:"size:150;not null;uniqueIndex:idx_seminar_email"`
	NIM         *string `gorm:"size:20"` // nullable: peserta umum tidak punya NIM
	Institution string  `gorm:"size:100;not null"`
	TicketCode  string  `gorm:"size:8;not null;uniqueIndex"`
	Attended    bool    `gorm:"not null;default:false"`
	AttendedAt  *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}
