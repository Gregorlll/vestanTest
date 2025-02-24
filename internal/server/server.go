package server

import (
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"

	"vestantest/internal/database"
	models "vestantest/internal/models"
	"vestantest/internal/server/config"

	"github.com/gorilla/websocket"
)

type Server struct {
	clients    map[*models.Client]bool
	broadcast  chan models.Message
	register   chan *models.Client
	unregister chan *models.Client
	db         *database.DB
	config     *config.Config
	mu         sync.Mutex
	logger     *log.Logger
	done       chan struct{}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { //TODO add check origin for production
		return true
	},
}

func NewServer(db *database.DB, cfg *config.Config) *Server {
	logger := log.New(os.Stdout, "[SERVER] ", log.LstdFlags)
	return &Server{
		clients:    make(map[*models.Client]bool),
		broadcast:  make(chan models.Message),
		register:   make(chan *models.Client),
		unregister: make(chan *models.Client),
		db:         db,
		config:     cfg,
		logger:     logger,
		done:       make(chan struct{}),
	}
}

func (s *Server) ValidateUsername(username string) bool {
	if len(username) < s.config.MinUsernameLen || len(username) > s.config.MaxUsernameLen {
		return false
	}
	match, _ := regexp.MatchString("^[a-zA-Z0-9._-]+$", username)
	return match
}

func (s *Server) Shutdown() {
	s.logger.Println("Starting server shutdown...")

	close(s.done)

	s.mu.Lock()
	defer s.mu.Unlock()

	for client := range s.clients {
		s.logger.Printf("Disconnecting client: %s", client.Username)
		s.db.LogConnection(client.Username, "disconnected")
		client.Conn.Close()
		delete(s.clients, client)
	}

	s.logger.Println("Server shutdown completed")
}

func (s *Server) Run() {
	s.logger.Println("Starting server message handling")
	for {
		select {
		case <-s.done:
			s.logger.Println("Stopping message handling")
			return

		case client := <-s.register:
			s.mu.Lock()
			s.clients[client] = true
			s.mu.Unlock()
			s.logger.Printf("New client connected: %s", client.Username)
			s.db.LogConnection(client.Username, "connected")

		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				client.Conn.Close()
				s.logger.Printf("Client disconnected: %s", client.Username)
			}
			s.mu.Unlock()
			s.db.LogConnection(client.Username, "disconnected")

		case message := <-s.broadcast:
			s.logger.Printf("Broadcasting message from %s", message.User)
			s.db.SaveMessage(message.User, message.Message)
			s.mu.Lock()
			for client := range s.clients {
				err := client.Conn.WriteJSON(message)
				if err != nil {
					s.logger.Printf("Error sending message to %s: %v", client.Username, err)
					client.Conn.Close()
					delete(s.clients, client)
				}
			}
			s.mu.Unlock()
		}
	}
}
