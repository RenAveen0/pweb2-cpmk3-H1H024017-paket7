package requests

// SeminarRequest menampung input form seminar (mirip FormRequest di Laravel).
// EventDate berupa string karena <input type="datetime-local"> mengirim teks
// "2026-11-20T09:00"; diubah menjadi time.Time di controller.
type SeminarRequest struct {
	Title     string `form:"title" label:"Judul seminar" validate:"required,min=5,max=150"`
	Speaker   string `form:"speaker" label:"Pembicara" validate:"required,min=3,max=100"`
	EventDate string `form:"event_date" label:"Tanggal pelaksanaan" validate:"required"`
	Location  string `form:"location" label:"Lokasi" validate:"required,min=3,max=150"`
	Quota     int    `form:"quota" label:"Kuota" validate:"required,gt=0,max=10000"`
	Fee       int64  `form:"fee" label:"Biaya" validate:"gte=0"`
}
