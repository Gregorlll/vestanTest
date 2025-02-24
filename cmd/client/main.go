package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"vestantest/internal/client"
)

func main() {
	c := client.NewClient()
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Welcome to the chat!")
	fmt.Println("Use /connect username to connect")
	fmt.Println("Use /exit to quit")

	for scanner.Scan() {
		text := scanner.Text()
		args := strings.Split(text, " ")

		switch args[0] {
		case "/connect":
			if len(args) != 2 {
				fmt.Println("Usage: /connect username")
				continue
			}

			serverURL := "ws://localhost:8080"
			err := c.Connect(serverURL, args[1])
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Printf("Connected to server as %s\n", args[1])
			c.Run()

		case "/exit":
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Println("Unknown command. Use /connect username or /exit")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("Error reading input:", err)
	}
}
