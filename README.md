# Go Chat App 🚀

A real-time chat application built with Go and WebSockets that supports multiple clients chatting simultaneously.

## Features

✨ **Real-time Communication** - Instant message delivery using WebSocket technology
👥 **Multi-User Support** - Handle multiple concurrent clients seamlessly
🔄 **Automatic Reconnection** - Clients automatically reconnect if disconnected
📝 **Clean UI** - Modern, responsive interface with smooth animations
🎨 **Username Display** - Each message shows the sender's name and timestamp
🔔 **Join/Leave Notifications** - System messages when users enter or exit the chat
🛡️ **Graceful Disconnect Handling** - Properly cleans up resources when clients disconnect

## Architecture

The application follows a clean, idiomatic Go design:

- **Hub**: Central manager that coordinates all client connections and message broadcasting
- **Client**: Represents individual WebSocket connections with read/write goroutines
- **Message**: JSON-based message structure with type, username, text, and timestamp
- **Channels & Goroutines**: Proper concurrency handling to prevent data races

## Prerequisites

- Go 1.16 or higher
- Modern web browser with WebSocket support

## Installation

1. Clone the repository:
```bash
git clone https://github.com/cloneyjay/go-chatapp.git
cd go-chatapp
```

2. Install dependencies:
```bash
go mod download
```

## Running the Application

1. Start the server:
```bash
go run server.go
```

2. Open your browser and navigate to:
```
http://localhost:8080
```

3. Enter a username and start chatting!

## Building the Application

To build a standalone executable:
```bash
go build -o chatserver server.go
```

Then run it:
```bash
./chatserver
```

## How It Works

### Server Components

**Hub (`Hub` struct)**:
- Maintains a registry of all connected clients
- Manages three channels:
  - `register`: for new client connections
  - `unregister`: for client disconnections
  - `broadcast`: for distributing messages to all clients
- Runs in a goroutine, continuously processing events

**Client (`Client` struct)**:
- Represents a single WebSocket connection
- Has two goroutines:
  - `readPump`: reads messages from the WebSocket connection
  - `writePump`: writes messages to the WebSocket connection
- Uses a buffered send channel for non-blocking message delivery

**Message Format**:
```json
{
  "type": "chat|join|leave",
  "username": "string",
  "text": "string",
  "time": "ISO8601 timestamp"
}
```

### WebSocket Endpoint

- **URL**: `ws://localhost:8080/ws?username=YourName`
- **Protocol**: WebSocket (binary messages for pings, text for chat)
- **Origin**: Accepts all origins (configurable via `CheckOrigin` in upgrader)

## Project Structure

```
go-chatapp/
├── server.go           # Main server with Hub, Client, and WebSocket logic
├── public/
│   └── index.html      # Frontend HTML/CSS/JavaScript
├── go.mod              # Go module file
├── go.sum              # Go dependencies checksum
└── README.md           # This file
```

## Testing Multiple Clients

To test the multi-user functionality:

1. Start the server
2. Open multiple browser windows/tabs to `http://localhost:8080`
3. Use different usernames in each window
4. Send messages from any window and watch them appear in all windows instantly

## Key Design Decisions

- **Gorilla WebSocket**: Industry-standard, well-maintained WebSocket library
- **JSON Messages**: Human-readable, easy to debug, widely supported
- **Ping/Pong Mechanism**: Keeps connections alive and detects dead connections
- **Non-blocking Sends**: Buffered channels prevent slow clients from blocking others
- **Graceful Cleanup**: Proper resource cleanup on client disconnect

## Security Notes

⚠️ **Development Mode**: The current configuration uses `CheckOrigin: true` which accepts WebSocket connections from any origin. For production, implement proper origin checking:

```go
CheckOrigin: func(r *http.Request) bool {
    origin := r.Header.Get("Origin")
    return origin == "https://yourdomain.com"
}
```

## License

See LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
