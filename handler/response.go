package handler

import "github.com/gofiber/fiber/v2"

type customResponse struct {
	Status int
	Data   string
}

type tokenResponse struct {
	Status      int
	Data        string
	AccessToken string
}

func HttpResponse(c *fiber.Ctx, status int, msg string) error {
	return c.JSON(&customResponse{
		Status: status,
		Data:   msg,
	})
}

// refresh 추가
func TokenResponse(c *fiber.Ctx, status int, msg, accessToken string) error {
	return c.JSON(&tokenResponse{
		Status:      status,
		Data:        msg,
		AccessToken: accessToken,
	})
}
