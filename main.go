package main

import (
	"jwt/db"
	"jwt/handler"
	"jwt/redis"
	"jwt/tokens"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// PostgreDB 연결
	db.ConnectDB()
	defer db.Sqldb.Close()

	// redis 연결
	if err := redis.ConnectRedis(); err != nil {
		panic(err)
	}

	// 컨트롤러
	r := fiber.New()

	api := r.Group("/api")
	v1 := api.Group("/v1")

	v1.Use(tokens.TokenAuthMiddleware()) // 토큰 검증 미들웨어

	v1.Post("/signup", handler.SignUp) // 회원가입 API
	v1.Post("/login", handler.Login)   // 로그인 API
	v1.Post("/logout", handler.LogOut) // 로그아웃 API
	v1.Get("/admin", handler.Admin)    // 로그인된 유저만 접속 가능

	r.Listen(":8080")
}
