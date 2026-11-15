package requests

// RegistrationRequest menampung input form pendaftaran peserta.
// Aturan yang butuh database (seminar ada, email unik per seminar, kuota)
// dicek di controller, di dalam transaksi.
type RegistrationRequest struct {
	SeminarID   uint   `form:"seminar_id" label:"Seminar" validate:"required"`
	Name        string `form:"name" label:"Nama" validate:"required,min=3,max=100"`
	Email       string `form:"email" label:"Email" validate:"required,email,max=150"`
	NIM         string `form:"nim" label:"NIM" validate:"omitempty,alphanum,min=5,max=20"`
	Institution string `form:"institution" label:"Instansi" validate:"required,min=2,max=100"`
}
