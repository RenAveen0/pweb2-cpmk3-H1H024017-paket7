package helpers

import (
	"strconv"

	"pweb2-cpmk3-H1H024017-paket7/config"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"
)

// View merender template dengan data bersama (shared data) yang selalu
// tersedia di semua halaman, mirip helper view() + View::share() di Laravel.
func View(c fiber.Ctx, name string, data fiber.Map) error {
	if data == nil {
		data = fiber.Map{}
	}
	data["AppName"] = config.Env("APP_NAME", "Seminar HMTK")
	data["CsrfToken"] = csrf.TokenFromContext(c)
	data["CurrentPath"] = c.Path()
	if _, ok := data["Errors"]; !ok {
		data["Errors"] = map[string]string{}
	}

	// Flash message dari redirect sebelumnya (mirip session('success') di Laravel)
	if msg := c.Redirect().Message("success"); msg.Value != "" {
		data["FlashSuccess"] = msg.Value
	}
	if msg := c.Redirect().Message("error"); msg.Value != "" {
		data["FlashError"] = msg.Value
	}
	if msg := c.Redirect().Message("warning"); msg.Value != "" {
		data["FlashWarning"] = msg.Value
	}

	return c.Render(name, data)
}

// ParseID membaca parameter :id lalu mengubahnya menjadi angka positif.
// Nilai non-angka (misal "/seminars/abc") ditolak sehingga controller membalas 404,
// dan string mentah dari URL tidak pernah masuk ke query SQL.
// Hasil konversi berupa angka baru, jadi aman dipakai (tidak menunjuk memori fasthttp).
func ParseID(c fiber.Ctx) (uint, bool) {
	n, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || n == 0 {
		return 0, false
	}
	return uint(n), true
}
