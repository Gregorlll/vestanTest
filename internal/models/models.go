package models

import (
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	User    string    `json:"user"`
	Time    time.Time `json:"time"`
	Message string    `json:"message"`
}

type ConnectionEvent struct {
	User      string    `json:"user"`
	Time      time.Time `json:"time"`
	EventType string    `json:"event"` // "connected" or "disconnected"
}

type MessagesResponse struct {
	Total    int       `json:"total"`
	Messages []Message `json:"messages"`
}

type Client struct {
	Username string
	Conn     *websocket.Conn
}
