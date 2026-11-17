package bootstrap

import (
	"errors"
	"strings"

	"pweb2-cpmk3-H1H024017-paket7/app/helpers"
	"pweb2-cpmk3-H1H024017-paket7/app/middleware"
	"pweb2-cpmk3-H1H024017-paket7/config"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/gofiber/template/html/v2"
)

// NewApp menyiapkan aplikasi Fiber: view engine, middleware, static file.
// Perannya mirip bootstrap/app.php di Laravel.
func NewApp() *fiber.App {
	// 1. View engine: folder resources/views, ekstensi .html
	engine := html.New("./resources/views", ".html")
	engine.Reload(config.Env("APP_ENV", "local") == "local") // auto reload template saat development
	engine.AddFunc("rupiah", helpers.Rupiah)
	engine.AddFunc("feeLabel", helpers.FeeLabel)
	engine.AddFunc("tanggal", helpers.Tanggal)
	engine.AddFunc("tanggalWaktu", helpers.TanggalWaktu)
	engine.AddFunc("statusClass", helpers.StatusClass)
	engine.AddFunc("barClass", helpers.BarClass)
	engine.AddFunc("hasPrefix", strings.HasPrefix)
	engine.AddFunc("add", func(a, b int) int { return a + b })

	// 2. Membuat instance Fiber
	app := fiber.New(fiber.Config{
		AppName:      config.Env("APP_NAME", "Seminar HMTK"),
		Views:        engine,
		ViewsLayout:  "layouts/app", // layout default untuk semua c.Render()
		ErrorHandler: errorHandler,
	})

	// 3. Middleware global (urutan penting!)
	app.Use(middleware.MethodOverride()) // _method=PUT/DELETE harus PALING AWAL
	app.Use(recover.New())               // cegah server mati karena panic
	app.Use(logger.New())                // log setiap request
	app.Use(csrf.New(csrf.Config{        // proteksi CSRF, mirip @csrf
		Extractor:      extractors.FromForm("_csrf"),
		CookieName:     "csrf_token",
		CookieSameSite: "Lax",
		ErrorHandler: func(c fiber.Ctx, err error) error {
			return c.Status(419).SendString("419 | Halaman kedaluwarsa (token CSRF tidak valid). Silakan muat ulang halaman.")
		},
	}))

	// 4. File statis: /public/css/app.css -> ./public/css/app.css
	app.Use("/public", static.New("./public"))

	return app
}

// errorHandler menampilkan halaman error (404, 500, dst) dengan Bootstrap.
// Dipanggil otomatis saat controller mengembalikan error, misal fiber.ErrNotFound.
func errorHandler(c fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}

	message := "Terjadi kesalahan pada server."
	if code == fiber.StatusNotFound {
		message = "Halaman atau data yang Anda cari tidak ditemukan."
	}

	c.Status(code)
	return helpers.View(c, "errors/error", fiber.Map{
		"Title":   "Error",
		"Code":    code,
		"Message": message,
	})
}
