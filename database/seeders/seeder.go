package seeders

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"pweb2-cpmk3-paket7/app/helpers"
	"pweb2-cpmk3-paket7/app/models"

	"gorm.io/gorm"
)

type seminarSeed struct {
	title    string
	speaker  string
	location string
	quota    int
	fee      int64
}

// Run mengisi data contoh (mirip "php artisan db:seed"): 20 seminar + pendaftarnya.
// Tanggal dibuat relatif terhadap hari ini: 6 seminar sudah lewat, 14 akan datang.
func Run(db *gorm.DB) {
	seminars := []seminarSeed{
		{"Pengenalan Go dan Fiber untuk Pemula", "Rizky Pratama, S.Kom.", "Aula Fakultas Teknik", 100, 0},
		{"Workshop REST API dengan GORM", "Dewi Anggraeni, M.T.", "Lab Komputer 2", 40, 50000},
		{"Dasar Keamanan Aplikasi Web", "Bagas Wicaksono", "Ruang Seminar Teknik", 6, 25000},
		{"Docker untuk Developer Pemula", "Nadia Putri", "Lab Jaringan", 35, 0},
		{"UI/UX Praktis dengan Bootstrap 5", "Salsabila Rahma", "Aula Fakultas Teknik", 80, 0},
		{"Version Control dengan Git dan GitHub", "Fajar Nugroho", "Lab Komputer 1", 50, 15000},
		{"Karier di Dunia Software Engineering", "Andika Saputra", "Auditorium Kampus", 200, 0},
		{"Mini Bootcamp Database MySQL", "Dr. Hendra Kusuma", "Lab Basis Data", 5, 75000},
		{"Internet of Things dengan ESP32", "Yoga Permana, S.T.", "Lab Elektronika", 30, 60000},
		{"Pengantar Machine Learning", "Dr. Laras Wulandari", "Auditorium Kampus", 120, 0},
		{"Membangun Portofolio Digital", "Intan Maharani", "Ruang Seminar Teknik", 60, 0},
		{"Cloud Computing untuk Mahasiswa", "Rendy Firmansyah", "Lab Jaringan", 45, 20000},
		{"Pemrograman Konkuren di Go", "Satria Adi, M.Kom.", "Lab Komputer 2", 25, 35000},
		{"Dasar Jaringan Komputer dan Mikrotik", "Eko Prasetyo", "Lab Jaringan", 40, 30000},
		{"Pitching Ide Startup Teknologi", "Maya Kartika", "Aula Fakultas Teknik", 90, 0},
		{"Testing Aplikasi Web Otomatis", "Bima Aryasatya", "Lab Komputer 1", 30, 40000},
		{"Desain Sistem Tertanam", "Dr. Wahyu Hidayat", "Lab Elektronika", 35, 45000},
		{"Teknik Presentasi untuk Mahasiswa Teknik", "Citra Lestari", "Ruang Seminar Teknik", 70, 0},
		{"Open Source dan Cara Berkontribusi", "Galih Ramadhan", "Lab Komputer 2", 50, 0},
		{"Persiapan Sertifikasi Profesi IT", "Ir. Budi Santoso", "Auditorium Kampus", 150, 100000},
	}

	firstNames := []string{"Aditya", "Bella", "Cahya", "Dimas", "Elsa", "Farhan", "Gita", "Hana", "Ilham", "Jihan",
		"Kevin", "Lina", "Miko", "Nesa", "Oki", "Putri", "Qori", "Raka", "Sinta", "Tegar"}
	lastNames := []string{"Pratama", "Lestari", "Wijaya", "Saputra", "Rahma", "Hidayat", "Utami", "Nugroho"}
	institutions := []string{"Universitas Jenderal Soedirman", "Universitas Negeri Semarang",
		"Politeknik Negeri Cilacap", "Universitas Muhammadiyah Purwokerto", "Umum"}

	r := rand.New(rand.NewSource(7)) // seed tetap => data selalu sama
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
	totalRegistrations := 0

	for i, s := range seminars {
		var eventDate time.Time
		if i < 6 {
			eventDate = today.AddDate(0, 0, -(i+1)*5) // sudah lewat -> status Selesai
		} else {
			eventDate = today.AddDate(0, 0, (i-5)*4) // akan datang
		}

		seminar := models.Seminar{
			Title: s.title, Speaker: s.speaker, EventDate: eventDate,
			Location: s.location, Quota: s.quota, Fee: s.fee,
		}
		// FirstOrCreate mencegah data dobel bila seeder dijalankan berkali-kali.
		db.Where(models.Seminar{Title: s.title}).FirstOrCreate(&seminar)

		// Jumlah pendaftar 3-8; seminar berkuota kecil diisi penuh agar badge "Penuh" terlihat.
		n := 3 + r.Intn(6)
		if s.quota <= 8 {
			n = s.quota
		}

		for j := 0; j < n; j++ {
			first := firstNames[r.Intn(len(firstNames))]
			last := lastNames[r.Intn(len(lastNames))]
			email := fmt.Sprintf("%s.%s.%d.%d@example.com", strings.ToLower(first), strings.ToLower(last), i+1, j+1)

			var nim *string
			if r.Intn(2) == 0 {
				v := fmt.Sprintf("H1H02%04d", 1000+r.Intn(8999))
				nim = &v
			}

			reg := models.Registration{
				SeminarID:   seminar.ID,
				Name:        first + " " + last,
				Email:       email,
				NIM:         nim,
				Institution: institutions[r.Intn(len(institutions))],
				TicketCode:  helpers.RandomTicketCode(),
			}
			// Peserta seminar yang sudah lewat: sebagian besar tercatat hadir.
			if eventDate.Before(now) && r.Intn(10) < 7 {
				at := eventDate.Add(time.Duration(r.Intn(40)) * time.Minute)
				reg.Attended = true
				reg.AttendedAt = &at
			}

			db.Where(models.Registration{SeminarID: seminar.ID, Email: email}).FirstOrCreate(&reg)
			totalRegistrations++
		}
	}

	log.Printf("Seeder selesai: %d seminar, %d pendaftar", len(seminars), totalRegistrations)
}
