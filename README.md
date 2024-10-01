# Localized-Connect 4

This project is a TCP-based multiplayer version of the classic **Connect 4** game, written in Go. It allows players to connect and play over a network in various game modes. The server handles communication between players, broadcasting moves, and determining game outcomes.

## Features

- **Multi-Player Support**: Supports multiple game modes including:
  - **Server vs. Client**
  - **Client vs. Client**
  - **Server vs. Computer**
  - **Client vs. Computer**
- **Customizable Board Size**: Choose between predefined sizes (8x6, 6x7, 9x14) or define a custom board size.
- **Graceful Networking**: The game utilizes TCP for smooth client-server communication and handles client disconnections gracefully.
- **Real-time Gameplay**: Players can make moves in real time, and the server broadcasts updates to all connected clients.

## How to Play

### Game Modes

1. **Server vs. Client**: The server plays against one connected client.
2. **Client vs. Client**: Two clients can connect and play against each other.
3. **Server vs. Computer**: The server plays against an AI opponent.
4. **Client vs. Computer**: A single client plays against an AI opponent.

### Board Size Selection

You can choose from the following board sizes:

- **8x6**: Standard 8 columns, 6 rows.
- **6x7**: Slightly smaller grid with 6 columns, 7 rows.
- **9x14**: A large grid with 9 columns, 14 rows.
- **Custom**: Set your own number of rows and columns.

### Running the Game

#### Server Setup:

1. Start the server on your machine:
   ```bash
   go run server.go
   ```

2. The server will prompt you to select a game mode and board size.

#### Client Setup:

1. Start the client to connect to the server:
   ```bash
   go run client.go
   ```

2. Enter the server IP and port (`localhost:8000` if running locally).

3. The game will prompt the player to enter moves by selecting columns.

### Example Gameplay

```
. . . . . . . 
. . . . . . . 
. . . . . . . 
. . . . . . . 
X O X O X O X 
1 2 3 4 5 6 7 

Player 1's turn. Enter column (1-7): 3
```

## Project Structure

- **server.go**: Manages client connections, gameplay logic, and broadcasts moves to clients.
- **client.go**: Connects to the server, handles user input, and displays game updates.


## Future Enhancements

- Add a GUI interface to replace the command-line display.
- Implement private messaging between players.
- Improve AI difficulty for the computer opponent.

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests with improvements.

