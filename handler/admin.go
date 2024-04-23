package handler

import "github.com/gofiber/fiber/v2"

func Admin(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Admin 진입 성공",
	})
}
