package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/manishchauhan/dugguGo/util/auth/jwtAuth"
	"github.com/rs/cors"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ErrorMessage struct {
	Error string `json:"error"`
}

func sendErrorMessage(conn *websocket.Conn, errorMsg string) error {
	message := ErrorMessage{Error: errorMsg}
	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, jsonData)
}

type IFMessage struct {
	Text string `json:"text"`
}

type WebSocketServer struct {
	upgrader      websocket.Upgrader
	addr          string
	connections   map[string]*websocket.Conn
	connectionsMu sync.Mutex
}

func NewWebSocketServer(addr string) *WebSocketServer {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	return &WebSocketServer{
		upgrader:    upgrader,
		addr:        addr,
		connections: make(map[string]*websocket.Conn),
	}
}

func (s *WebSocketServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.HandleWebSocket)
	fmt.Printf("WebSocket server listening on %s\n", s.addr)
	corsHandler := cors.Default().Handler(mux)
	return http.ListenAndServe(s.addr, corsHandler)
}

func (s *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading connection:", err)
		sendErrorMessage(conn, "Error upgrading connection")
		return
	}
	//remove connection when user disconnect it
	defer func() {
		// Remove the connection when a user disconnects
		s.connectionsMu.Lock()
		defer s.connectionsMu.Unlock()

		// Find the username associated with the connection
		var usernameToDelete string
		for username, userConn := range s.connections {
			if userConn == conn {
				usernameToDelete = username
				break
			}
		}
		println(usernameToDelete + "hello")
		// If a username is found, remove it from the map
		if usernameToDelete != "" {
			fmt.Printf("%s left the room\n", usernameToDelete)
			delete(s.connections, usernameToDelete)
			conn.Close()
		}
	}()

	//use refresh token to get user data starts___________________________________
	refreshTokenCookieName := "refresh_token"
	refreshToken, refreshTokenErr := r.Cookie(refreshTokenCookieName)
	if refreshTokenErr != nil {
		err = sendMessageToClient(conn, "Access token is invalid. Please log in again.")
		if err != nil {
			fmt.Println("Error sending welcome message:", err)
			return
		}
		return
	}
	claims, parseError := jwtAuth.ParseAndValidateToken(refreshToken.Value)
	if parseError != nil {
		err = sendMessageToClient(conn, "Access token is invalid. Please log in again.")
		if err != nil {
			fmt.Println("Error sending welcome message:", err)
			return
		}
		return
	}
	chatUserName := claims.UserName
	err = sendMessageToClient(conn, "Welcome to chat room "+chatUserName)
	if err != nil {
		fmt.Println("Error sending welcome message:", err)
		return
	}
	s.connectionsMu.Lock()
	s.connections[chatUserName] = conn
	s.connectionsMu.Unlock()
	//use refresh token to get user data starts___________________________________

	incomingMessages := make(chan []byte)
	outgoingMessages := make(chan []byte)

	var wgIncoming sync.WaitGroup
	var wgOutgoing sync.WaitGroup

	wgIncoming.Add(1)
	wgOutgoing.Add(1)

	go s.handleIncoming(&wgIncoming, conn, incomingMessages, outgoingMessages)
	go s.handleOutgoing(&wgOutgoing, outgoingMessages)

	wgIncoming.Wait()
	wgOutgoing.Wait()

	close(incomingMessages)
	close(outgoingMessages)
}

func (s *WebSocketServer) handleIncoming(wg *sync.WaitGroup, conn *websocket.Conn, incomingMessages chan<- []byte, outgoingMessages chan<- []byte) {
	defer wg.Done()
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			sendErrorMessage(conn, "Error reading message")
			return
		}
		fmt.Println("Received message from client:", string(p))

		// Send incoming message to outgoingMessages channel
		select {
		case outgoingMessages <- p:
			// Message sent to outgoingMessages channel
		default:
			fmt.Println("Warning: outgoingMessages channel is full, skipping message")
		}
	}
}

func (s *WebSocketServer) handleOutgoing(wg *sync.WaitGroup, outgoingMessages <-chan []byte) {
	defer wg.Done()
	for msg := range outgoingMessages {
		// Lock the mutex while iterating through the connections map
		s.connectionsMu.Lock()
		for _, conn := range s.connections {
			err := sendMessageToClient(conn, string(msg))
			if err != nil {
				fmt.Println("Error sending message:", err)
				// Optionally, you might want to remove the connection from the map here
			}
		}
		s.connectionsMu.Unlock()
	}
}

func sendMessageToClient(conn *websocket.Conn, message string) error {
	msg := IFMessage{Text: message}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = conn.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		return err
	}

	return nil
}
