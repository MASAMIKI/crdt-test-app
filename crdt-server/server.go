package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// 参考: https://github.com/googollee/go-socket.io/blob/master/_examples/redis-adapter/main.go
func main() {
	//　socket ioのサーバーを作成
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	//　socket ioのサーバーを作成
	server := createSocketIOServer()

	//　socket ioのサーバーをechoでラップ
	e.GET("/socket.io/*", echo.WrapHandler(server))
	e.POST("/socket.io/*", echo.WrapHandler(server))
	e.Logger.Fatal(e.Start(":8080"))
}
