package main

import (
	"context"
	"encoding/json"
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

	e.GET("/project/:id/karte", karteHandler)
	e.Logger.Fatal(e.Start(":8080"))
}

// ReceiveMessage represents a message sent to or from a client.
type ReceiveMessage struct {
	Key     string `json:"key"`
	Content string `json:"content"`
	SentBy  string `json:"sentBy"`
}

// SendMessage represents a message sent to or from a client.
type SendMessage struct {
	Key       string   `json:"key"`
	Content   string   `json:"content"`
	SentBy    string   `json:"sentBy"`
	ProjectId string   `json:"projectId"`
	UserIds   []string `json:"userIds"`
}

// joinRoom 入室
func joinRoom(roomName string, clientID string) error {
	roomKey := fmt.Sprintf("%s-room", roomName)
	_, err := rdb.SAdd(ctx, roomKey, clientID).Result()
	return err
}

// leaveRoom 退室
func leaveRoom(roomName string, clientID string) error {
	roomKey := fmt.Sprintf("%s-room", roomName)
	_, err := rdb.SRem(ctx, roomKey, clientID).Result()
	return err
}

// getUsersInRoom ルームにいるユーザー一覧を取得
func getUsersInRoom(roomName string) ([]string, error) {
	roomKey := fmt.Sprintf("%s-room", roomName)
	return rdb.SMembers(ctx, roomKey).Result()
}

// cleanRoom ルームにいるユーザーがいなくなったらデータを削除
func cleanRoom(roomName string) error {
	userIds, err := getUsersInRoom(roomName)
	if err != nil {
		return err
	}
	if len(userIds) == 0 {
		roomKey := fmt.Sprintf("%s-room", roomName)
		key := "message"
		dataKey := fmt.Sprintf("%s-data-%s", roomName, key)
		// ここでDBに保存する
		_, err = rdb.Del(ctx, roomKey, dataKey).Result()
		if err != nil {
			return err
		}
	}
	return nil
}

// readRoomData ルームのデータを読み込む
func recoverRoomData(roomName string) ([]byte, error) {
	// DBからRedisへのデータ移行
	key := "message"
	record := currentRecords[key]
	dataKey := fmt.Sprintf("%s-data-%s", roomName, key)

	err := rdb.Watch(ctx, func(tx *redis.Tx) error {
		// Check if the key exists
		n, err := tx.Exists(ctx, dataKey).Result()
		if err != nil {
			return err
		}

		// If the key doesn't exist, add the value
		if n == 0 {
			for _, o := range record {
				data, err := json.Marshal(o)
				if err != nil {
					return err
				}
				_, err = tx.Pipelined(ctx, func(pipe redis.Pipeliner) error {
					pipe.RPush(ctx, dataKey, string(data))
					return nil
				})
				return err
			}
		}
		return nil
	}, dataKey)

	if err != nil {
		return []byte{}, err
	}

	// 最新のデータを取得
	values, err := rdb.LRange(ctx, dataKey, 0, -1).Result()
	if err != nil {
		return []byte{}, err
	}

	userIDs, err := getUsersInRoom(roomName)
	if err != nil {
		return []byte{}, err
	}

	messages := make([]SendMessage, 0)
	for _, v := range values {
		var currentData O
		err = json.Unmarshal([]byte(v), &currentData)
		if err != nil {
			return []byte{}, err
		}

		messages = append(messages, SendMessage{
			Key:       key,
			Content:   currentData.Value,
			SentBy:    currentData.UpdatedBy,
			ProjectId: roomName,
			UserIds:   userIDs,
		})
	}

	return json.Marshal(messages)
}

// rPush データを追加
func rPush(roomName string, message ReceiveMessage) error {
	key := message.Key
	value := O{
		Value:     message.Content,
		UpdatedBy: message.SentBy,
		UpdateAt:  time.Now(),
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	dataKey := fmt.Sprintf("%s-data-%s", roomName, key)
	_, err = rdb.RPush(ctx, dataKey, data).Result()
	if err != nil {
		return err
	}
	return nil
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
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to upgrade connection: %v", err))
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
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to join room: %v", err))
	}

	data, err := recoverRoomData(id)
	err = ws.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to write message: %v", err))
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

			messages := []SendMessage{{
				Key:       receiveMsg.Key,
				Content:   receiveMsg.Content,
				SentBy:    receiveMsg.SentBy,
				ProjectId: id,
				UserIds:   userIDs,
			}}

			data, err := json.Marshal(messages)
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

		var receiveMsg ReceiveMessage
		err = json.Unmarshal(message, &receiveMsg)
		if err != nil {
			log.Println("Failed to encode data:", err)
			continue
		}

		err = rPush(id, receiveMsg)
		if err != nil {
			log.Println("Failed to encode data:", err)
			continue
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
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to leave room: %v", err))
	}

	// 部屋にだれもいなければ片付ける
	err = cleanRoom(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to clean room: %v", err))
	}
	return nil
}
