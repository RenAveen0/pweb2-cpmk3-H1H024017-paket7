package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

// MethodOverride membaca field tersembunyi "_method" pada form HTML.
// Form HTML hanya mendukung GET & POST, jadi untuk PUT/PATCH/DELETE kita
// kirim POST + <input type="hidden" name="_method" value="PUT">,
// persis seperti @method('PUT') di Laravel Blade.
func MethodOverride() fiber.Handler {
	return func(c fiber.Ctx) error {
		if c.Method() == fiber.MethodPost {
			method := strings.ToUpper(c.FormValue("_method"))
			switch method {
			case fiber.MethodPut, fiber.MethodPatch, fiber.MethodDelete:
				c.Method(method)          // ubah method request
				return c.RestartRouting() // cocokkan ulang ke route PUT/DELETE
			}
		}
		return c.Next()
	}
}
