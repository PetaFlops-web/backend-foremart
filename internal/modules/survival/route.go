package survival

import "github.com/gofiber/fiber/v2"

func (m *Module) RegisterRoutes(router fiber.Router, authMiddleware fiber.Handler) {
	router.Post("/predict-survival", authMiddleware, m.Controller.Predict)
}
