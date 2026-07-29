package transaction

import (
	"github.com/gofiber/fiber/v2"
)

func (m *Module) RegisterRoutes(router fiber.Router, authMiddleware fiber.Handler) {
	transactions := router.Group("/transactions", authMiddleware)

	// Voice extraction (Preview)
	transactions.Post("/extract/voice", m.Controller.ExtractVoice)

	// CRUD
	transactions.Post("/", m.Controller.Create)
	transactions.Get("/", m.Controller.List)
	transactions.Get("/:id", m.Controller.Get)
	transactions.Delete("/:id", m.Controller.Delete)
}