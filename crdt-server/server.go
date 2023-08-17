package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

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

	e.GET("/project/:id/karte", karteHandler)
	e.Logger.Fatal(e.Start(":8080"))
}

// ReceiveMessage represents a message sent to or from a client.
type ReceiveMessage struct {
	Content string `json:"content"`
	SentBy  string `json:"sentBy"`
}

// SendMessage represents a message sent to or from a client.
type SendMessage struct {
	Content   string   `json:"content"`
	SentBy    string   `json:"sentBy"`
	ProjectId string   `json:"projectId"`
	UserIds   []string `json:"userIds"`
}

// joinRoom 入室
func joinRoom(roomName string, clientID string) error {
	_, err := rdb.SAdd(ctx, roomName, clientID).Result()
	return err
}

// leaveRoom 退室
func leaveRoom(roomName string, clientID string) error {
	_, err := rdb.SRem(ctx, roomName, clientID).Result()
	return err
}

// getUsersInRoom ルームにいるユーザー一覧を取得
func getUsersInRoom(roomName string) ([]string, error) {
	return rdb.SMembers(ctx, roomName).Result()
}

func karteHandler(c echo.Context) error {
	id := c.Param("id") // RoomNameとして利用
	userID := c.QueryParam("userId")

	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid room id")
	}

	if userID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}

	// WebSocketの接続を確立
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer func() {
		err = ws.Close()
		if err != nil {
			log.Printf("Error while closing websocket: %v", err)
		}
	}()

	// 入室
	err = joinRoom(id, userID)
	if err != nil {
		log.Printf("Could not set the clientID-room mapping: %v", err)
	}

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
			userIDs, err := getUsersInRoom(id)
			if err != nil {
				log.Println("Failed to encode data:", err)
				continue
			}

			var receiveMsg ReceiveMessage
			err = json.Unmarshal([]byte(sub.Payload), &receiveMsg)
			if err != nil {
				log.Println("Failed to encode data:", err)
				continue
			}

			msg := SendMessage{
				Content:   receiveMsg.Content,
				SentBy:    receiveMsg.SentBy,
				ProjectId: id,
				UserIds:   userIDs,
			}

			data, err := json.Marshal(msg)
			if err != nil {
				log.Println("Failed to encode data:", err)
				continue
			}
			err = ws.WriteMessage(websocket.BinaryMessage, data)
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

	// 退室
	err = leaveRoom(id, userID)
	if err != nil {
		log.Printf("Could not delete the clientID-room mapping: %v", err)
	}

	return nil
}
