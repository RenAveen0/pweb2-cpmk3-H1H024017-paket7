package helpers

import "crypto/rand"

// Alfabet tanpa karakter yang mudah tertukar (0/O, 1/I). Panjangnya 32,
// pembagi 256, sehingga pemetaan byte -> karakter tidak bias.
const ticketAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// RandomTicketCode menghasilkan kode tiket acak 8 karakter memakai crypto/rand
// (aman, tidak bisa ditebak). Keunikan dicek oleh controller + unique index DB.
func RandomTicketCode() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		panic(err) // kegagalan sumber acak OS = kondisi fatal
	}
	out := make([]byte, 8)
	for i, b := range buf {
		out[i] = ticketAlphabet[int(b)%len(ticketAlphabet)]
	}
	return string(out)
}
