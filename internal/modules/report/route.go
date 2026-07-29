package report

import "github.com/gofiber/fiber/v2"

func (m *Module) RegisterRoutes(router fiber.Router, authMiddleware fiber.Handler) {
	reports := router.Group("/reports", authMiddleware)
	reports.Get("/daily", m.Controller.Daily)
}
