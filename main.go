package main

import (
	"log"
	"os"

	"pweb2-cpmk3-H1H024017-paket7/bootstrap"
	"pweb2-cpmk3-H1H024017-paket7/config"
	"pweb2-cpmk3-H1H024017-paket7/database/migrations"
	"pweb2-cpmk3-H1H024017-paket7/database/seeders"
	"pweb2-cpmk3-H1H024017-paket7/routes"

	"github.com/gofiber/fiber/v3"
)

func main() {
	config.LoadEnv()
	config.ConnectDatabase()

	// Perintah CLI ala "php artisan":
	//   go run . migrate        -> buat tabel
	//   go run . migrate:fresh  -> hapus semua tabel lalu buat ulang
	//   go run . seed           -> isi data contoh
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			migrations.Migrate(config.DB)
		case "migrate:fresh":
			migrations.Fresh(config.DB)
		case "seed":
			seeders.Run(config.DB)
		default:
			log.Fatalf("Perintah tidak dikenal: %s (gunakan: migrate | migrate:fresh | seed)", os.Args[1])
		}
		return
	}

	// Tanpa argumen: jalankan web server (mirip "php artisan serve")
	app := bootstrap.NewApp()
	routes.Web(app)

	// 404 untuk route yang tidak terdaftar (harus paling akhir)
	app.Use(func(c fiber.Ctx) error { return fiber.ErrNotFound })

	log.Fatal(app.Listen(":" + config.Env("APP_PORT", "3000")))
}
