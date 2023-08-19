package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/shiguredo/websocket"
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

	e.GET("/project/karte/:id", karteHandler)
	e.Logger.Fatal(e.Start(":8080"))
}

// 仮データ
var currentRecords = KV{
	"message": []O{{Value: "おはようございます。", UpdatedBy: "matsubara", UpdateAt: time.Now()}},
}

// KV 仮データの型
type KV map[string][]O
type O struct {
	Value     string    `json:"value"`
	UpdatedBy string    `json:"updatedBy"`
	UpdateAt  time.Time `json:"updatedAt"`
}

func karteHandler(c echo.Context) error {
	id := c.Param("id") // RoomNameとして利用

	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room id")
	}

	// WebSocketの接続を確立
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to upgrade connection: %v", err))
	}
	defer func() {
		err = ws.Close()
		if err != nil {
			log.Printf("Error while closing websocket: %v", err)
		}
	}()

	// サブスクライブの開始
	pubsub := rdb.Subscribe(ctx, id)
	defer func() {
		err = pubsub.Close()
		if err != nil {
			log.Printf("Error while closing pubsub: %v", err)
		}
	}()

	// RedisからのメッセージをWebSocketに転送
	go func() {
		ch := pubsub.Channel()
		for sub := range ch {
			err = ws.WriteMessage(websocket.BinaryMessage, []byte(sub.Payload))
			if err != nil {
				log.Printf("Error while writing message: %v", err)
			}
		}
	}()

	// WebSocketからのメッセージをRedisのPub/Subチャネルに発行
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Error while reading message: %v", err)
			break
		}

		_, err = rdb.Publish(ctx, id, message).Result()
		if err != nil {
			log.Println("Failed to decode:", err)
			break
		}
	}

	return nil
}
