package customer

import "github.com/gofiber/fiber/v2"

func (m *Module) RegisterRoutes(router fiber.Router, authMiddleware fiber.Handler) {
	customers := router.Group("/customers", authMiddleware)
	customers.Post("/", m.Controller.Create)
	customers.Get("/", m.Controller.Search)
	customers.Get("/:id", m.Controller.Get)
}
