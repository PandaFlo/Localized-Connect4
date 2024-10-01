package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Game struct represents the game state
type Game struct {
	board       [][]int         // 2D slice to represent the board
	rows        int             // Number of rows
	columns     int             // Number of columns
	player      int             // Current player (1 or 2)
	mode        string          // Game mode
	connections []net.Conn      // Slice to hold client connections
	readers     []*bufio.Reader // Slice to hold buffered readers for each connection
}

func main() {
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator

	// Ask for board size
	columns, rows := selectBoardSize() // Reversed order: columns, rows

	// Ask for game mode
	mode := selectGameMode()

	// Create a new game with the selected settings
	game := NewGame(rows, columns, mode) // Pass rows and columns in correct order

	if mode == "ServerVsComputer" {
		// Run the game locally without networking
		game.PlayLocal()
	} else {
		// Start the TCP server
		game.startTCPServer()

		// Start the game loop
		game.Play()
	}
}

// Function to select the board size
func selectBoardSize() (int, int) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Select Board Size:")
		fmt.Println("1. 8x6")  // 8 columns x 6 rows
		fmt.Println("2. 6x7")  // 6 columns x 7 rows
		fmt.Println("3. 9x14") // 9 columns x 14 rows
		fmt.Println("4. Custom")
		fmt.Print("Enter your choice: ")

		choiceStr, _ := reader.ReadString('\n')
		choiceStr = strings.TrimSpace(choiceStr)
		choice, err := strconv.Atoi(choiceStr)
		if err != nil || choice < 1 || choice > 4 {
			fmt.Println("Invalid choice. Please try again.")
			continue
		}

		switch choice {
		case 1:
			return 8, 6 // 8 columns, 6 rows
		case 2:
			return 6, 7 // 6 columns, 7 rows
		case 3:
			return 9, 14 // 9 columns, 14 rows
		case 4:
			// Ask for custom size
			minSize := 4
			maxSize := 20
			fmt.Printf("Enter number of columns (%d-%d): ", minSize, maxSize)
			colsStr, _ := reader.ReadString('\n')
			colsStr = strings.TrimSpace(colsStr)
			cols, err1 := strconv.Atoi(colsStr)

			fmt.Printf("Enter number of rows (%d-%d): ", minSize, maxSize)
			rowsStr, _ := reader.ReadString('\n')
			rowsStr = strings.TrimSpace(rowsStr)
			rows, err2 := strconv.Atoi(rowsStr)

			if err1 != nil || err2 != nil || rows < minSize || cols < minSize || rows > maxSize || cols > maxSize {
				fmt.Printf("Invalid board size. Minimum size is %dx%d and maximum size is %dx%d.\n", minSize, minSize, maxSize, maxSize)
				continue
			}

			if cols > 9 {
				fmt.Println("Warning: The board may look wonky when columns exceed 9.")
			}

			return cols, rows // Return columns, rows
		}
	}
}

// Function to select the game mode
func selectGameMode() string {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Select Game Mode:")
		fmt.Println("1. Server vs Client")
		fmt.Println("2. Client vs Client")
		fmt.Println("3. Server vs Computer")
		fmt.Println("4. Client vs Computer")
		fmt.Print("Enter your choice: ")

		choiceStr, _ := reader.ReadString('\n')
		choiceStr = strings.TrimSpace(choiceStr)
		choice, err := strconv.Atoi(choiceStr)
		if err != nil || choice < 1 || choice > 4 {
			fmt.Println("Invalid choice. Please try again.")
			continue
		}

		switch choice {
		case 1:
			return "ServerVsClient"
		case 2:
			return "ClientVsClient"
		case 3:
			return "ServerVsComputer"
		case 4:
			return "ClientVsComputer"
		}
	}
}

// NewGame initializes the game with dynamic board size and selected mode
func NewGame(rows, columns int, mode string) *Game {
	// Initialize the board as a 2D slice
	board := make([][]int, rows)
	for i := range board {
		board[i] = make([]int, columns)
	}
	return &Game{
		board:       board,
		rows:        rows,
		columns:     columns,
		player:      1,
		mode:        mode,
		connections: make([]net.Conn, 0),
		readers:     make([]*bufio.Reader, 0),
	}
}

// startTCPServer starts the server and accepts client connections based on the game mode
func (g *Game) startTCPServer() {
	// Listen on port 8000
	ln, err := net.Listen("tcp", ":8000")
	if err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}
	fmt.Println("Server started on port 8000")

	expectedPlayers := 0
	switch g.mode {
	case "ServerVsClient", "ClientVsComputer":
		expectedPlayers = 1
	case "ClientVsClient":
		expectedPlayers = 2
	}

	// Accept connections
	for i := 1; i <= expectedPlayers; i++ {
		var playerNumber int
		switch g.mode {
		case "ServerVsClient":
			playerNumber = 2 // Waiting for Player 2 to connect
		case "ClientVsClient":
			playerNumber = i // Waiting for Player 1 and then Player 2
		case "ClientVsComputer":
			playerNumber = 1 // Waiting for Player 1 to connect
		}

		fmt.Printf("Waiting for Player %d to connect...\n", playerNumber)
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			os.Exit(1)
		}
		fmt.Printf("Player %d connected\n", playerNumber)
		g.connections = append(g.connections, conn)
		g.readers = append(g.readers, bufio.NewReader(conn))
	}

	fmt.Println("All players connected. Starting the game!")

	// Display the initial empty board to all players
	g.broadcastBoard()
}

// Play handles the main game loop, taking turns based on the game mode
func (g *Game) Play() {
	for {
		var col int
		var err error

		// Determine whose turn it is and get input accordingly
		switch g.mode {
		case "ServerVsClient":
			if g.player == 1 {
				// Server's turn
				g.PrintBoard()
				fmt.Printf("Your turn (Player %d). Enter column (1-%d): ", g.player, g.columns)
				// Notify client that the server is making a move
				g.sendTurnPrompt(g.connections[0], false)
				col, err = g.getPlayerInput()
				if err != nil {
					fmt.Println("Error getting input:", err)
					fmt.Println("Game ended due to error.")
					g.notifyClients("Game ended due to error.\n")
					return
				}
			} else {
				// Client's turn
				g.sendTurnPrompt(g.connections[0], true)
				fmt.Printf("Player %d (Client) is making a move...\n", g.player)
				col, err = g.getClientInput(0)
				if err != nil {
					fmt.Printf("Error reading from client: %v\n", err)
					fmt.Println("Game ended due to client disconnection.")
					return
				}
			}
		case "ClientVsClient":
			if g.player == 1 {
				// Player 1 (Client 1)
				g.sendTurnPrompt(g.connections[0], true)
				g.sendTurnPrompt(g.connections[1], false)
				col, err = g.getClientInput(0)
				if err != nil {
					fmt.Printf("Error reading from client 1: %v\n", err)
					fmt.Println("Game ended due to client disconnection.")
					return
				}
			} else {
				// Player 2 (Client 2)
				g.sendTurnPrompt(g.connections[1], true)
				g.sendTurnPrompt(g.connections[0], false)
				col, err = g.getClientInput(1)
				if err != nil {
					fmt.Printf("Error reading from client 2: %v\n", err)
					fmt.Println("Game ended due to client disconnection.")
					return
				}
			}
		case "ClientVsComputer":
			if g.player == 1 {
				// Client's turn
				g.sendTurnPrompt(g.connections[0], true)
				fmt.Printf("Player %d (Client) is making a move...\n", g.player)
				col, err = g.getClientInput(0)
				if err != nil {
					fmt.Printf("Error reading from client: %v\n", err)
					fmt.Println("Game ended due to client disconnection.")
					return
				}
			} else {
				// Computer's turn
				col = rand.Intn(g.columns) + 1
				fmt.Printf("Computer (Player %d) chooses column %d\n", g.player, col)
			}
		}

		// Validate input
		if err == nil {
			if col < 1 || col > g.columns {
				// Inform the current player about invalid input
				g.informInvalidInput()
				continue
			}

			if !g.MakeMove(col - 1) {
				// Inform the current player that the column is full
				g.informColumnFull()
				continue
			}
		} else {
			// Handle error from input (e.g., invalid data from client)
			fmt.Printf("Error: %v\n", err)
			g.informInvalidInput()
			continue
		}

		// Broadcast the updated board
		g.broadcastBoard()

		// Print the updated board to the server console
		g.PrintBoard()

		if g.CheckWin() {
			fmt.Printf("Player %d wins!\n", g.player)
			// Notify clients of the game over
			g.notifyClients(fmt.Sprintf("Player %d wins!\n", g.player))
			break
		}
		if g.IsDraw() {
			fmt.Println("It's a draw!")
			g.notifyClients("It's a draw!\n")
			break
		}
		g.SwitchPlayer()
	}

	// Close client connections
	for _, conn := range g.connections {
		conn.Close()
	}
}

// PlayLocal handles the game loop for Server vs Computer mode
func (g *Game) PlayLocal() {
	// Display the initial empty board
	g.PrintBoard()

	for {
		var col int
		var err error

		if g.player == 1 {
			// Server's turn
			fmt.Printf("Your turn (Player %d). Enter column (1-%d): ", g.player, g.columns)
			col, err = g.getPlayerInput()
			if err != nil {
				fmt.Println("Error getting input:", err)
				fmt.Println("Game ended due to error.")
				return
			}
		} else {
			// Computer's turn
			col = rand.Intn(g.columns) + 1
			fmt.Printf("Computer (Player %d) chooses column %d\n", g.player, col)
		}

		if col < 1 || col > g.columns {
			fmt.Println("Invalid input. Please enter a number between 1 and", g.columns)
			continue
		}

		if !g.MakeMove(col - 1) {
			fmt.Println("Column is full. Try another one.")
			continue
		}

		// Display the updated board
		g.PrintBoard()

		if g.CheckWin() {
			g.PrintBoard()
			fmt.Printf("Player %d wins!\n", g.player)
			break
		}
		if g.IsDraw() {
			g.PrintBoard()
			fmt.Println("It's a draw!")
			break
		}
		g.SwitchPlayer()
	}
}

// getPlayerInput reads input from the server console
func (g *Game) getPlayerInput() (int, error) {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	col, err := strconv.Atoi(input)
	if err != nil {
		return 0, err
	}
	return col, nil
}

// getClientInput reads input from a client over TCP
func (g *Game) getClientInput(index int) (int, error) {
	reader := g.readers[index]

	// Read input from the client
	input, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return 0, fmt.Errorf("empty input")
	}
	col, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("invalid input: %v", err)
	}
	return col, nil
}

// sendTurnPrompt sends a prompt to the client indicating if it's their turn or the other player's turn
func (g *Game) sendTurnPrompt(conn net.Conn, isYourTurn bool) {
	var message string
	if isYourTurn {
		message = fmt.Sprintf("Your turn (Player %d). Enter column (1-%d): \n", g.player, g.columns)
	} else {
		message = fmt.Sprintf("Player %d is making a move...\n", g.player)
	}
	conn.Write([]byte(message))
}

// broadcastBoard sends the current board state to all clients
func (g *Game) broadcastBoard() {
	boardStr := g.getBoardString()
	for _, conn := range g.connections {
		conn.Write([]byte(boardStr))
	}
}

// getBoardString returns the board as a string
func (g *Game) getBoardString() string {
	var sb strings.Builder
	sb.WriteString("\n")
	for i := 0; i < g.rows; i++ {
		for j := 0; j < g.columns; j++ {
			var c string
			switch g.board[i][j] {
			case 0:
				c = "."
			case 1:
				c = "X"
			case 2:
				c = "O"
			}
			sb.WriteString(c + " ")
		}
		sb.WriteString("\n")
	}
	// Append column numbers
	for i := 1; i <= g.columns; i++ {
		sb.WriteString(fmt.Sprintf("%d ", i))
	}
	sb.WriteString("\n\n")
	return sb.String()
}

// notifyClients sends a message to all connected clients
func (g *Game) notifyClients(message string) {
	for _, conn := range g.connections {
		conn.Write([]byte(message))
	}
}

// informInvalidInput informs the current player about invalid input
func (g *Game) informInvalidInput() {
	message := "Invalid input. Please enter a valid column number.\n"
	switch g.mode {
	case "ServerVsClient":
		if g.player == 1 {
			fmt.Print(message)
		} else {
			g.connections[0].Write([]byte(message))
		}
	case "ClientVsClient":
		currentConn := g.connections[g.player-1]
		currentConn.Write([]byte(message))
	case "ClientVsComputer":
		if g.player == 1 {
			g.connections[0].Write([]byte(message))
		} else {
			fmt.Print(message)
		}
	}
}

// informColumnFull informs the current player that the selected column is full
func (g *Game) informColumnFull() {
	message := "Column is full. Try another one.\n"
	switch g.mode {
	case "ServerVsClient":
		if g.player == 1 {
			fmt.Print(message)
		} else {
			g.connections[0].Write([]byte(message))
		}
	case "ClientVsClient":
		currentConn := g.connections[g.player-1]
		currentConn.Write([]byte(message))
	case "ClientVsComputer":
		if g.player == 1 {
			g.connections[0].Write([]byte(message))
		} else {
			fmt.Print(message)
		}
	}
}

// MakeMove attempts to place a player's piece in the given column
func (g *Game) MakeMove(col int) bool {
	for i := g.rows - 1; i >= 0; i-- {
		if g.board[i][col] == 0 {
			g.board[i][col] = g.player
			return true
		}
	}
	return false
}

// SwitchPlayer alternates the current player between 1 and 2
func (g *Game) SwitchPlayer() {
	if g.player == 1 {
		g.player = 2
	} else {
		g.player = 1
	}
}

// IsDraw checks if the board is full and the game is a draw
func (g *Game) IsDraw() bool {
	for i := 0; i < g.columns; i++ {
		if g.board[0][i] == 0 {
			return false
		}
	}
	return true
}

// CheckWin checks if the current player has 4 consecutive pieces
func (g *Game) CheckWin() bool {
	// Horizontal check
	for row := 0; row < g.rows; row++ {
		for col := 0; col <= g.columns-4; col++ {
			if g.board[row][col] == g.player &&
				g.board[row][col+1] == g.player &&
				g.board[row][col+2] == g.player &&
				g.board[row][col+3] == g.player {
				return true
			}
		}
	}
	// Vertical check
	for col := 0; col < g.columns; col++ {
		for row := 0; row <= g.rows-4; row++ {
			if g.board[row][col] == g.player &&
				g.board[row+1][col] == g.player &&
				g.board[row+2][col] == g.player &&
				g.board[row+3][col] == g.player {
				return true
			}
		}
	}
	// Diagonal (bottom-left to top-right)
	for row := 3; row < g.rows; row++ {
		for col := 0; col <= g.columns-4; col++ {
			if g.board[row][col] == g.player &&
				g.board[row-1][col+1] == g.player &&
				g.board[row-2][col+2] == g.player &&
				g.board[row-3][col+3] == g.player {
				return true
			}
		}
	}
	// Diagonal (top-left to bottom-right)
	for row := 0; row <= g.rows-4; row++ {
		for col := 0; col <= g.columns-4; col++ {
			if g.board[row][col] == g.player &&
				g.board[row+1][col+1] == g.player &&
				g.board[row+2][col+2] == g.player &&
				g.board[row+3][col+3] == g.player {
				return true
			}
		}
	}
	return false
}

// PrintBoard displays the current state of the game board
func (g *Game) PrintBoard() {
	fmt.Println()
	for i := 0; i < g.rows; i++ {
		for j := 0; j < g.columns; j++ {
			var c string
			switch g.board[i][j] {
			case 0:
				c = "."
			case 1:
				c = "X"
			case 2:
				c = "O"
			}
			fmt.Print(c, " ")
		}
		fmt.Println()
	}
	for i := 1; i <= g.columns; i++ {
		fmt.Print(i, " ")
	}
	fmt.Println("\n")
}
