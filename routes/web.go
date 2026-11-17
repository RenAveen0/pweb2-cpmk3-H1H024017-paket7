package routes

import (
	"pweb2-cpmk3-H1H024017-paket7/app/controllers"

	"github.com/gofiber/fiber/v3"
)

// Web mendaftarkan semua route aplikasi (mirip routes/web.php).
func Web(app *fiber.App) {
	home := &controllers.HomeController{}
	seminar := &controllers.SeminarController{}
	registration := &controllers.RegistrationController{}
	checkin := &controllers.CheckinController{}

	app.Get("/", home.Index).Name("home")

	// Route::resource('seminars', SeminarController::class)
	// PENTING: /seminars/create harus SEBELUM /seminars/:id.
	app.Get("/seminars", seminar.Index).Name("seminars.index")
	app.Get("/seminars/create", seminar.Create).Name("seminars.create")
	app.Post("/seminars", seminar.Store).Name("seminars.store")
	app.Get("/seminars/:id", seminar.Show).Name("seminars.show")
	app.Get("/seminars/:id/edit", seminar.Edit).Name("seminars.edit")
	app.Put("/seminars/:id", seminar.Update).Name("seminars.update")
	app.Delete("/seminars/:id", seminar.Destroy).Name("seminars.destroy")
	app.Get("/seminars/:id/export", seminar.ExportAttendance).Name("seminars.export")

	// Route::resource('registrations', RegistrationController::class)->except('show')
	app.Get("/registrations", registration.Index).Name("registrations.index")
	app.Get("/registrations/create", registration.Create).Name("registrations.create")
	app.Post("/registrations", registration.Store).Name("registrations.store")
	app.Get("/registrations/:id/edit", registration.Edit).Name("registrations.edit")
	app.Put("/registrations/:id", registration.Update).Name("registrations.update")
	app.Delete("/registrations/:id", registration.Destroy).Name("registrations.destroy")

	// Check-in peserta (T3)
	app.Get("/checkin", checkin.Index).Name("checkin.index")
	app.Post("/checkin", checkin.Store).Name("checkin.store")
}
