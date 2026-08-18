package restock

import "github.com/gofiber/fiber/v2"

func (m *Module) RegisterRoutes(router fiber.Router, authMiddleware fiber.Handler) {
	restockPredictions := router.Group("/restock-predictions", authMiddleware)
	restockPredictions.Post("/_generate", m.Controller.Generate)
	restockPredictions.Get("/", m.Controller.List)
}
