package migrations

import (
	"log"

	"pweb2-cpmk3-H1H024017-paket7/app/models"

	"gorm.io/gorm"
)

// Migrate membuat / memperbarui tabel sesuai struct model
// (mirip "php artisan migrate"). Urutan penting: tabel induk dulu.
func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.Seminar{},
		&models.Registration{},
	)
	if err != nil {
		log.Fatalf("Migrasi gagal: %v", err)
	}
	log.Println("Migrasi selesai")
}

// Fresh menghapus semua tabel lalu migrasi ulang (mirip "migrate:fresh").
// Urutan drop: tabel anak (registrations) dulu agar foreign key tidak menghalangi.
func Fresh(db *gorm.DB) {
	if err := db.Migrator().DropTable(&models.Registration{}, &models.Seminar{}); err != nil {
		log.Fatalf("Gagal menghapus tabel: %v", err)
	}
	log.Println("Semua tabel dihapus")
	Migrate(db)
}
