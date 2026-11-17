package controllers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"pweb2-cpmk3-H1H024017-paket7/app/helpers"
	"pweb2-cpmk3-H1H024017-paket7/app/models"
	"pweb2-cpmk3-H1H024017-paket7/app/requests"
	"pweb2-cpmk3-H1H024017-paket7/config"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// seminarSelect = mirip withCount('registrations') di Laravel: jumlah pendaftar
// dihitung database lewat subquery, jadi 1 query saja (tanpa N+1).
const seminarSelect = "seminars.*, (SELECT COUNT(*) FROM registrations WHERE registrations.seminar_id = seminars.id) AS registrations_count"

// Format nilai <input type="datetime-local">.
const dateTimeLayout = "2006-01-02T15:04"

// SeminarController = resource controller untuk seminar
// (index, create, store, show, edit, update, destroy) + export CSV daftar hadir.
type SeminarController struct{}

// Index GET /seminars — daftar seminar + pencarian + pagination + progress bar kuota.
func (ctl *SeminarController) Index(c fiber.Ctx) error {
	search := strings.TrimSpace(c.Query("q"))
	page := fiber.Query[int](c, "page", 1)

	base := config.DB.Model(&models.Seminar{})
	if search != "" {
		like := "%" + search + "%"
		base = base.Where("(title LIKE ? OR speaker LIKE ?)", like, like)
	}

	// Session(&gorm.Session{}) membuat salinan query, sehingga Count dan Find
	// tidak saling mengganggu.
	var total int64
	base.Session(&gorm.Session{}).Count(&total)

	pagination := helpers.Paginate(page, 10, total).WithQuery(map[string]string{"q": search})

	var seminars []models.Seminar
	base.Session(&gorm.Session{}).
		Select(seminarSelect).
		Order("event_date desc").
		Limit(pagination.PerPage).
		Offset(pagination.Offset()).
		Find(&seminars)

	return helpers.View(c, "seminars/index", fiber.Map{
		"Title":      "Seminar",
		"Seminars":   seminars,
		"Pagination": pagination,
		"Search":     search,
	})
}

// Show GET /seminars/:id — detail seminar + daftar peserta.
func (ctl *SeminarController) Show(c fiber.Ctx) error {
	seminar, err := ctl.findSeminar(c)
	if err != nil {
		return err
	}

	var registrations []models.Registration
	config.DB.Where("seminar_id = ?", seminar.ID).Order("created_at desc").Find(&registrations)

	var attended int64
	config.DB.Model(&models.Registration{}).
		Where("seminar_id = ? AND attended = ?", seminar.ID, true).
		Count(&attended)

	return helpers.View(c, "seminars/show", fiber.Map{
		"Title":         "Detail Seminar",
		"Seminar":       seminar,
		"Registrations": registrations,
		"Attended":      attended,
	})
}

// Create GET /seminars/create — form tambah seminar.
func (ctl *SeminarController) Create(c fiber.Ctx) error {
	return helpers.View(c, "seminars/create", fiber.Map{
		"Title": "Tambah Seminar",
		"Form":  requests.SeminarRequest{},
	})
}

// Store POST /seminars — validasi lalu simpan seminar baru.
func (ctl *SeminarController) Store(c fiber.Ctx) error {
	req, eventDate, errs := ctl.validateRequest(c)
	if len(errs) > 0 {
		c.Status(fiber.StatusUnprocessableEntity)
		return helpers.View(c, "seminars/create", fiber.Map{
			"Title": "Tambah Seminar", "Form": req, "Errors": errs,
		})
	}

	seminar := models.Seminar{
		Title:     req.Title,
		Speaker:   req.Speaker,
		EventDate: eventDate,
		Location:  req.Location,
		Quota:     req.Quota,
		Fee:       req.Fee,
	}
	if err := config.DB.Create(&seminar).Error; err != nil {
		return c.Redirect().With("error", "Gagal menyimpan seminar: "+err.Error()).To("/seminars")
	}
	return c.Redirect().With("success", "Seminar \""+seminar.Title+"\" berhasil ditambahkan.").To("/seminars")
}

// Edit GET /seminars/:id/edit — form ubah seminar.
func (ctl *SeminarController) Edit(c fiber.Ctx) error {
	seminar, err := ctl.findSeminar(c)
	if err != nil {
		return err
	}
	return helpers.View(c, "seminars/edit", fiber.Map{
		"Title":   "Ubah Seminar",
		"Seminar": seminar,
		"Form": requests.SeminarRequest{
			Title:     seminar.Title,
			Speaker:   seminar.Speaker,
			EventDate: helpers.DateTimeInput(seminar.EventDate),
			Location:  seminar.Location,
			Quota:     seminar.Quota,
			Fee:       seminar.Fee,
		},
	})
}

// Update PUT /seminars/:id — validasi lalu simpan perubahan.
func (ctl *SeminarController) Update(c fiber.Ctx) error {
	seminar, err := ctl.findSeminar(c)
	if err != nil {
		return err
	}

	req, eventDate, errs := ctl.validateRequest(c)
	// Kuota tidak boleh lebih kecil dari jumlah peserta yang sudah mendaftar.
	if _, bad := errs["quota"]; !bad && int64(req.Quota) < seminar.RegistrationsCount {
		errs["quota"] = fmt.Sprintf("Kuota tidak boleh kurang dari jumlah pendaftar saat ini (%d).", seminar.RegistrationsCount)
	}
	if len(errs) > 0 {
		c.Status(fiber.StatusUnprocessableEntity)
		return helpers.View(c, "seminars/edit", fiber.Map{
			"Title": "Ubah Seminar", "Seminar": seminar, "Form": req, "Errors": errs,
		})
	}

	seminar.Title = req.Title
	seminar.Speaker = req.Speaker
	seminar.EventDate = eventDate
	seminar.Location = req.Location
	seminar.Quota = req.Quota
	seminar.Fee = req.Fee
	config.DB.Omit(clause.Associations).Save(&seminar)

	return c.Redirect().With("success", "Seminar \""+seminar.Title+"\" berhasil diperbarui.").To("/seminars")
}

// Destroy DELETE /seminars/:id — hapus seminar (pendaftarnya ikut terhapus oleh CASCADE).
func (ctl *SeminarController) Destroy(c fiber.Ctx) error {
	seminar, err := ctl.findSeminar(c)
	if err != nil {
		return err
	}
	if err := config.DB.Delete(&seminar).Error; err != nil {
		return c.Redirect().With("error", "Gagal menghapus seminar: "+err.Error()).To("/seminars")
	}
	return c.Redirect().
		With("success", fmt.Sprintf("Seminar \"%s\" berhasil dihapus beserta %d pendaftarnya.", seminar.Title, seminar.RegistrationsCount)).
		To("/seminars")
}

// ExportAttendance GET /seminars/:id/export — unduh daftar hadir seminar (CSV).
func (ctl *SeminarController) ExportAttendance(c fiber.Ctx) error {
	seminar, err := ctl.findSeminar(c)
	if err != nil {
		return err
	}

	var registrations []models.Registration
	config.DB.Where("seminar_id = ? AND attended = ?", seminar.ID, true).
		Order("attended_at asc").
		Find(&registrations)

	var buf bytes.Buffer
	buf.WriteString("\xEF\xBB\xBF") // BOM UTF-8 agar Excel membaca huruf dengan benar

	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"No", "Kode Tiket", "Nama", "Email", "NIM", "Instansi", "Waktu Check-in"})
	for i, r := range registrations {
		nim := ""
		if r.NIM != nil {
			nim = *r.NIM
		}
		checkIn := ""
		if r.AttendedAt != nil {
			checkIn = helpers.TanggalWaktu(*r.AttendedAt)
		}
		_ = w.Write([]string{
			fmt.Sprint(i + 1),
			r.TicketCode,
			helpers.CSVSafe(r.Name),
			helpers.CSVSafe(r.Email),
			helpers.CSVSafe(nim),
			helpers.CSVSafe(r.Institution),
			checkIn,
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return fiber.ErrInternalServerError
	}

	filename := "daftar-hadir-" + helpers.Slug(seminar.Title) + ".csv"
	c.Set(fiber.HeaderContentType, "text/csv; charset=utf-8")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+filename+`"`)
	return c.Send(buf.Bytes())
}

// findSeminar mencari seminar (beserta jumlah pendaftar) berdasarkan :id.
// Id tidak valid / data tidak ada -> fiber.ErrNotFound (halaman 404 kustom).
func (ctl *SeminarController) findSeminar(c fiber.Ctx) (models.Seminar, error) {
	var seminar models.Seminar
	id, ok := helpers.ParseID(c)
	if !ok {
		return seminar, fiber.ErrNotFound
	}
	if err := config.DB.Select(seminarSelect).First(&seminar, id).Error; err != nil {
		return seminar, fiber.ErrNotFound
	}
	return seminar, nil
}

// validateRequest mengikat (bind) input form, memvalidasi, lalu mengubah
// teks tanggal menjadi time.Time. Dipakai bersama oleh Store dan Update.
func (ctl *SeminarController) validateRequest(c fiber.Ctx) (requests.SeminarRequest, time.Time, map[string]string) {
	var req requests.SeminarRequest
	if err := c.Bind().Form(&req); err != nil {
		return req, time.Time{}, map[string]string{"_form": "Format input tidak valid, periksa kembali angka kuota dan biaya."}
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Speaker = strings.TrimSpace(req.Speaker)
	req.Location = strings.TrimSpace(req.Location)
	req.EventDate = strings.TrimSpace(req.EventDate)

	errs := requests.Validate(req)

	var eventDate time.Time
	if _, bad := errs["event_date"]; !bad {
		t, err := time.ParseInLocation(dateTimeLayout, req.EventDate, time.Local)
		if err != nil {
			errs["event_date"] = "Format tanggal pelaksanaan tidak valid."
		} else {
			eventDate = t
		}
	}
	return req, eventDate, errs
}
