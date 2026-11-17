package controllers

import (
	"time"

	"pweb2-cpmk3-H1H024017-paket7/app/helpers"
	"pweb2-cpmk3-H1H024017-paket7/app/models"
	"pweb2-cpmk3-H1H024017-paket7/config"

	"github.com/gofiber/fiber/v3"
)

type HomeController struct{}

// Index menampilkan dashboard ringkasan data.
func (h *HomeController) Index(c fiber.Ctx) error {
	var totalSeminars, upcomingCount, totalRegistrations, totalAttended int64
	now := time.Now()

	config.DB.Model(&models.Seminar{}).Count(&totalSeminars)
	config.DB.Model(&models.Seminar{}).Where("event_date >= ?", now).Count(&upcomingCount)
	config.DB.Model(&models.Registration{}).Count(&totalRegistrations)
	config.DB.Model(&models.Registration{}).Where("attended = ?", true).Count(&totalAttended)

	// Seminar mendatang beserta jumlah pendaftar (subquery COUNT, tanpa N+1).
	var upcoming []models.Seminar
	config.DB.Select(seminarSelect).
		Where("event_date >= ?", now).
		Order("event_date asc").
		Limit(5).
		Find(&upcoming)

	return helpers.View(c, "home/index", fiber.Map{
		"Title":              "Dashboard",
		"TotalSeminars":      totalSeminars,
		"UpcomingCount":      upcomingCount,
		"TotalRegistrations": totalRegistrations,
		"TotalAttended":      totalAttended,
		"Upcoming":           upcoming,
	})
}
