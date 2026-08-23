package notification

import "github.com/gofiber/fiber/v2"

func (m *Module) RegisterRoutes(router fiber.Router, authMiddleware fiber.Handler) {
	notifications := router.Group("/notifications", authMiddleware)
	notifications.Get("/", m.Controller.List)
	notifications.Post("/_send", m.Controller.Send)
}
