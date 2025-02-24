package client

import (
	"bufio"
	"fmt"
	"io"
	"os"

	models "vestantest/internal/models"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn     *websocket.Conn
	username string
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Connect(serverURL, username string) error {
	url := fmt.Sprintf("%s/ws?username=%s", serverURL, username)
	conn, resp, err := websocket.DefaultDialer.Dial(url, nil)

	if err != nil {
		if resp != nil {
			// Читаем сообщение об ошибке от сервера
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("%s", string(body))
		}
		return err
	}

	c.conn = conn
	c.username = username
	return nil
}

func (c *Client) Run() {
	go c.receiveMessages()
	c.sendMessages()
}

func (c *Client) receiveMessages() {
	for {
		var msg models.Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Lost connection to server")
			os.Exit(1)
		}
		fmt.Printf("[%s] %s: %s\n", msg.Time.Format("15:04:05"), msg.User, msg.Message)
	}
}

func (c *Client) sendMessages() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()

		if text == "/exit" {
			c.conn.Close()
			fmt.Println("Disconnected from server")
			os.Exit(0)
		}

		msg := models.Message{
			Message: text,
		}

		if err := c.conn.WriteJSON(msg); err != nil {
			fmt.Println("Error sending message:", err)
			return
		}
	}
}
