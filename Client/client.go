package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// Connect to the server on localhost port 8000
	conn, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	serverReader := bufio.NewReader(conn)

	for {
		// Read message from server
		serverMessage, err := serverReader.ReadString('\n')
		if err != nil {
			fmt.Println("Connection closed by server.")
			return
		}

		serverMessage = strings.TrimSpace(serverMessage)

		if serverMessage == "" {
			continue
		}

		fmt.Println(serverMessage)

		// Continue reading any additional messages
		for serverReader.Buffered() > 0 {
			extraMessage, err := serverReader.ReadString('\n')
			if err != nil {
				fmt.Println("Connection closed by server.")
				return
			}
			extraMessage = strings.TrimSpace(extraMessage)
			if extraMessage != "" {
				fmt.Println(extraMessage)
			}
		}

		// If prompted for input
		if strings.Contains(serverMessage, "Enter column") {
			// Read user input
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			// Send input to server
			_, err := conn.Write([]byte(input + "\n"))
			if err != nil {
				fmt.Println("Error sending input to server:", err)
				return
			}
		}
	}
}
