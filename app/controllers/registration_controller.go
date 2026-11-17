package controllers

import (
	"errors"
	"strconv"
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

// RegistrationController = resource controller pendaftar
// (index, create, store, edit, update, destroy).
type RegistrationController struct{}

// Index GET /registrations — daftar pendaftar + cari + filter seminar + pagination.
func (ctl *RegistrationController) Index(c fiber.Ctx) error {
	search := strings.TrimSpace(c.Query("q"))
	seminarID := fiber.Query[int](c, "seminar", 0)
	page := fiber.Query[int](c, "page", 1)

	base := config.DB.Model(&models.Registration{})
	if search != "" {
		like := "%" + search + "%"
		base = base.Where("(name LIKE ? OR email LIKE ? OR ticket_code LIKE ?)", like, like, like)
	}
	if seminarID > 0 {
		base = base.Where("seminar_id = ?", seminarID)
	}

	var total int64
	base.Session(&gorm.Session{}).Count(&total)

	pagination := helpers.Paginate(page, 10, total).WithQuery(map[string]string{
		"q":       search,
		"seminar": strconv.Itoa(seminarID),
	})

	var registrations []models.Registration
	base.Session(&gorm.Session{}).
		Preload("Seminar"). // eager loading, mirip with('seminar') -> tanpa N+1
		Order("created_at desc, id desc").
		Limit(pagination.PerPage).
		Offset(pagination.Offset()).
		Find(&registrations)

	return helpers.View(c, "registrations/index", fiber.Map{
		"Title":         "Pendaftar",
		"Registrations": registrations,
		"Seminars":      ctl.seminarOptions(),
		"Pagination":    pagination,
		"Search":        search,
		"SeminarID":     uint(seminarID),
	})
}

// Create GET /registrations/create — form tambah pendaftar (?seminar=ID untuk pilihan awal).
func (ctl *RegistrationController) Create(c fiber.Ctx) error {
	preselected := fiber.Query[int](c, "seminar", 0)
	return helpers.View(c, "registrations/create", fiber.Map{
		"Title":    "Tambah Pendaftar",
		"Form":     requests.RegistrationRequest{SeminarID: uint(preselected)},
		"Seminars": ctl.seminarOptions(),
	})
}

// Store POST /registrations — validasi lalu simpan pendaftar baru.
func (ctl *RegistrationController) Store(c fiber.Ctx) error {
	req, errs := ctl.bindAndValidate(c)

	var registration models.Registration
	if len(errs) == 0 {
		// Cek kuota + unik email + simpan harus satu transaksi. Baris seminar
		// dikunci (FOR UPDATE) agar dua pendaftar bersamaan tidak melewati kuota.
		err := config.DB.Transaction(func(tx *gorm.DB) error {
			errs = ctl.applyRules(tx, req, 0, true)
			if len(errs) > 0 {
				return nil // tidak ada yang disimpan
			}
			code, err := ctl.uniqueTicketCode(tx)
			if err != nil {
				return err
			}
			registration = models.Registration{
				SeminarID:   req.SeminarID,
				Name:        req.Name,
				Email:       req.Email,
				NIM:         nimPointer(req.NIM),
				Institution: req.Institution,
				TicketCode:  code,
			}
			return tx.Create(&registration).Error
		})
		if err != nil {
			return c.Redirect().With("error", "Gagal menyimpan pendaftar: "+err.Error()).To("/registrations")
		}
	}

	if len(errs) > 0 {
		c.Status(fiber.StatusUnprocessableEntity)
		return helpers.View(c, "registrations/create", fiber.Map{
			"Title": "Tambah Pendaftar", "Form": req, "Errors": errs,
			"Seminars": ctl.seminarOptions(),
		})
	}
	return c.Redirect().
		With("success", "Pendaftaran \""+registration.Name+"\" berhasil. Kode tiket: "+registration.TicketCode).
		To("/registrations")
}

// Edit GET /registrations/:id/edit — form ubah pendaftar.
func (ctl *RegistrationController) Edit(c fiber.Ctx) error {
	registration, err := ctl.findRegistration(c)
	if err != nil {
		return err
	}
	nim := ""
	if registration.NIM != nil {
		nim = *registration.NIM
	}
	return helpers.View(c, "registrations/edit", fiber.Map{
		"Title":        "Ubah Pendaftar",
		"Registration": registration,
		"Form": requests.RegistrationRequest{
			SeminarID:   registration.SeminarID,
			Name:        registration.Name,
			Email:       registration.Email,
			NIM:         nim,
			Institution: registration.Institution,
		},
		"Seminars": ctl.seminarOptions(),
	})
}

// Update PUT /registrations/:id — validasi lalu simpan perubahan.
func (ctl *RegistrationController) Update(c fiber.Ctx) error {
	registration, err := ctl.findRegistration(c)
	if err != nil {
		return err
	}

	req, errs := ctl.bindAndValidate(c)
	if len(errs) == 0 {
		// Kuota & tanggal hanya dicek bila pendaftar dipindah ke seminar lain.
		seminarChanged := req.SeminarID != registration.SeminarID
		err := config.DB.Transaction(func(tx *gorm.DB) error {
			errs = ctl.applyRules(tx, req, registration.ID, seminarChanged)
			if len(errs) > 0 {
				return nil
			}
			registration.SeminarID = req.SeminarID
			registration.Name = req.Name
			registration.Email = req.Email
			registration.NIM = nimPointer(req.NIM)
			registration.Institution = req.Institution
			return tx.Omit(clause.Associations).Save(&registration).Error
		})
		if err != nil {
			return c.Redirect().With("error", "Gagal memperbarui pendaftar: "+err.Error()).To("/registrations")
		}
	}

	if len(errs) > 0 {
		c.Status(fiber.StatusUnprocessableEntity)
		return helpers.View(c, "registrations/edit", fiber.Map{
			"Title": "Ubah Pendaftar", "Registration": registration, "Form": req, "Errors": errs,
			"Seminars": ctl.seminarOptions(),
		})
	}
	return c.Redirect().With("success", "Pendaftar \""+registration.Name+"\" berhasil diperbarui.").To("/registrations")
}

// Destroy DELETE /registrations/:id — hapus pendaftar.
func (ctl *RegistrationController) Destroy(c fiber.Ctx) error {
	registration, err := ctl.findRegistration(c)
	if err != nil {
		return err
	}
	config.DB.Delete(&registration)
	return c.Redirect().With("success", "Pendaftar \""+registration.Name+"\" berhasil dihapus.").To("/registrations")
}

// applyRules memeriksa aturan bisnis di dalam transaksi tx:
//  1. seminar harus ada (baris dikunci FOR UPDATE),
//  2. bila checkSeminar: seminar belum lewat dan kuota belum penuh,
//  3. email unik per seminar (abaikan baris ignoreID saat update).
func (ctl *RegistrationController) applyRules(tx *gorm.DB, req requests.RegistrationRequest, ignoreID uint, checkSeminar bool) map[string]string {
	errs := map[string]string{}

	var seminar models.Seminar
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&seminar, req.SeminarID).Error; err != nil {
		errs["seminar_id"] = "Seminar yang dipilih tidak ditemukan."
		return errs
	}

	if checkSeminar {
		if seminar.EventDate.Before(time.Now()) {
			errs["seminar_id"] = "Seminar sudah selesai, pendaftaran ditutup."
		} else {
			var taken int64
			tx.Model(&models.Registration{}).
				Where("seminar_id = ? AND id <> ?", seminar.ID, ignoreID).
				Count(&taken)
			if taken >= int64(seminar.Quota) {
				errs["seminar_id"] = "Kuota seminar sudah penuh."
			}
		}
	}

	var duplicate int64
	tx.Model(&models.Registration{}).
		Where("seminar_id = ? AND email = ? AND id <> ?", req.SeminarID, req.Email, ignoreID).
		Count(&duplicate)
	if duplicate > 0 {
		errs["email"] = "Email ini sudah terdaftar pada seminar tersebut."
	}
	return errs
}

// uniqueTicketCode membuat kode tiket acak 8 karakter yang belum dipakai.
func (ctl *RegistrationController) uniqueTicketCode(tx *gorm.DB) (string, error) {
	for i := 0; i < 10; i++ {
		code := helpers.RandomTicketCode()
		var used int64
		tx.Model(&models.Registration{}).Where("ticket_code = ?", code).Count(&used)
		if used == 0 {
			return code, nil
		}
	}
	return "", errors.New("gagal membuat kode tiket unik, coba lagi")
}

// bindAndValidate mengikat input form lalu memvalidasi format (belum aturan DB).
func (ctl *RegistrationController) bindAndValidate(c fiber.Ctx) (requests.RegistrationRequest, map[string]string) {
	var req requests.RegistrationRequest
	if err := c.Bind().Form(&req); err != nil {
		return req, map[string]string{"_form": "Format input tidak valid, periksa kembali isian form."}
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.NIM = strings.ToUpper(strings.TrimSpace(req.NIM))
	req.Institution = strings.TrimSpace(req.Institution)
	return req, requests.Validate(req)
}

func (ctl *RegistrationController) findRegistration(c fiber.Ctx) (models.Registration, error) {
	var registration models.Registration
	id, ok := helpers.ParseID(c)
	if !ok {
		return registration, fiber.ErrNotFound
	}
	if err := config.DB.First(&registration, id).Error; err != nil {
		return registration, fiber.ErrNotFound
	}
	return registration, nil
}

// seminarOptions = daftar seminar untuk dropdown (lengkap dengan jumlah pendaftar).
func (ctl *RegistrationController) seminarOptions() []models.Seminar {
	var seminars []models.Seminar
	config.DB.Select(seminarSelect).Order("event_date desc").Find(&seminars)
	return seminars
}

// nimPointer: NIM kosong disimpan NULL, bukan string kosong.
func nimPointer(nim string) *string {
	if nim == "" {
		return nil
	}
	return &nim
}
