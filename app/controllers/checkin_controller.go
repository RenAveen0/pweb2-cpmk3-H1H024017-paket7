package controllers

import (
	"strings"
	"time"

	"pweb2-cpmk3-H1H024017-paket7/app/helpers"
	"pweb2-cpmk3-H1H024017-paket7/app/models"
	"pweb2-cpmk3-H1H024017-paket7/config"

	"github.com/gofiber/fiber/v3"
)

// AttendanceRecap = satu baris rekap kehadiran per seminar.
type AttendanceRecap struct {
	ID         uint
	Title      string
	EventDate  time.Time
	Registered int64
	Attended   int64
}

// Percent = persentase kehadiran (hadir / terdaftar).
func (r AttendanceRecap) Percent() int {
	if r.Registered == 0 {
		return 0
	}
	return int(r.Attended * 100 / r.Registered)
}

// CheckinController mengurus halaman check-in peserta di hari acara.
type CheckinController struct{}

// Index GET /checkin — form input kode tiket + rekap kehadiran per seminar.
func (ctl *CheckinController) Index(c fiber.Ctx) error {
	// Satu query agregat (LEFT JOIN + GROUP BY), bukan hitung per seminar di loop.
	var recap []AttendanceRecap
	config.DB.Raw(`
		SELECT s.id, s.title, s.event_date,
		       COUNT(r.id) AS registered,
		       COUNT(CASE WHEN r.attended = 1 THEN 1 END) AS attended
		FROM seminars s
		LEFT JOIN registrations r ON r.seminar_id = s.id
		GROUP BY s.id, s.title, s.event_date
		ORDER BY s.event_date DESC`).Scan(&recap)

	return helpers.View(c, "checkin/index", fiber.Map{
		"Title": "Check-in",
		"Recap": recap,
	})
}

// Store POST /checkin — tandai peserta hadir berdasarkan kode tiket.
func (ctl *CheckinController) Store(c fiber.Ctx) error {
	code := strings.ToUpper(strings.TrimSpace(c.FormValue("ticket_code")))
	if code == "" || len(code) > 20 {
		return c.Redirect().With("error", "Masukkan kode tiket yang valid.").To("/checkin")
	}

	var registration models.Registration
	if err := config.DB.Preload("Seminar").Where("ticket_code = ?", code).First(&registration).Error; err != nil {
		return c.Redirect().With("error", "Kode tiket \""+code+"\" tidak ditemukan.").To("/checkin")
	}

	if registration.Attended {
		return c.Redirect().With("warning", alreadyCheckedIn(registration)).To("/checkin")
	}

	// UPDATE ... WHERE attended = false bersifat atomik: bila dua petugas memindai
	// kode yang sama bersamaan, hanya satu yang berhasil (RowsAffected = 1).
	result := config.DB.Model(&models.Registration{}).
		Where("id = ? AND attended = ?", registration.ID, false).
		Updates(map[string]any{"attended": true, "attended_at": time.Now()})
	if result.Error != nil {
		return c.Redirect().With("error", "Gagal mencatat kehadiran: "+result.Error.Error()).To("/checkin")
	}
	if result.RowsAffected == 0 {
		return c.Redirect().With("warning", alreadyCheckedIn(registration)).To("/checkin")
	}

	return c.Redirect().
		With("success", "Check-in berhasil: "+registration.Name+" — seminar \""+registration.Seminar.Title+"\".").
		To("/checkin")
}

func alreadyCheckedIn(r models.Registration) string {
	when := ""
	if r.AttendedAt != nil {
		when = " pada " + helpers.TanggalWaktu(*r.AttendedAt)
	}
	return r.Name + " sudah check-in sebelumnya" + when + " (seminar \"" + r.Seminar.Title + "\")."
}
