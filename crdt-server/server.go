package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
)

var (
	rdb      *redis.Client
	ctx      context.Context
	upgrader websocket.Upgrader
)

func init() {
	ctx = context.Background()
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	allowedOrigins := map[string]bool{
		"http://localhost:3000": true,
	}
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return allowedOrigins[r.Header.Get("Origin")]
		},
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/ws", handleConnection)
	e.Logger.Fatal(e.Start(":8080"))
}

func handleConnection(c echo.Context) error {
	// WebSocketの接続を確立
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	pubsub := rdb.Subscribe(ctx, "chat")
	defer pubsub.Close()

	// RedisからのメッセージをWebSocketに転送
	go func() {
		ch := pubsub.Channel()
		for msg := range ch {
			ws.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
		}
	}()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			break
		}

		// 受信したメッセージをRedisのPub/Subチャネルに発行
		log.Printf("recv: %s", msg)
		rdb.Publish(ctx, "chat", string(msg))
	}

	return nil
}
