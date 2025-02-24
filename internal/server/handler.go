package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	models "vestantest/internal/models"
)

func (s *Server) HandleMessages(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	messages, total, err := s.db.GetMessages(page, pageSize)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.MessagesResponse{
		Total:    total,
		Messages: messages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) HandleConnectionHistory(w http.ResponseWriter, r *http.Request) {
	events, err := s.db.GetConnectionHistory()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if !s.ValidateUsername(username) {
		http.Error(w, "Error: Username must be 3-10 characters long and contain only letters, digits, '-', '_', or '.'", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error: WebSocket upgrade: %v", err)
		return
	}

	client := &models.Client{
		Username: username,
		Conn:     conn,
	}

	// Send successful connection message
	conn.WriteJSON(models.Message{
		User:    "System",
		Message: fmt.Sprintf("Connected as %s.", username),
		Time:    time.Now(),
	})

	s.register <- client
	go s.handleClientMessages(client)
}

func (s *Server) handleClientMessages(client *models.Client) {
	defer func() {
		s.unregister <- client
		// Send disconnection message to all
		s.broadcast <- models.Message{
			User:    "System",
			Message: fmt.Sprintf("%s has disconnected.", client.Username),
			Time:    time.Now(),
		}
	}()

	for {
		var msg models.Message
		err := client.Conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		msg.User = client.Username
		msg.Time = time.Now()
		s.broadcast <- msg
	}
}
