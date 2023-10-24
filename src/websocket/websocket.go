package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
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
	Time string `json:"time"`
	Text string `json:"text"`
	User string `json:"user"`
}

type WebSocketServer struct {
	upgrader            websocket.Upgrader
	addr                string
	connections         map[string]*websocket.Conn
	connectionsMu       sync.Mutex
	connectionsMuShared sync.RWMutex
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
func (s *WebSocketServer) removeClosedConnections() {
	s.connectionsMu.Lock()
	defer s.connectionsMu.Unlock()
	for connectionKey, conn := range s.connections {
		if conn != nil {
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					// Handle unexpected close errors
					fmt.Println("Unexpected error while handling connection:", err)
				}
				// Connection is closed, remove it from the map
				delete(s.connections, connectionKey)
			}
		}
	}
}
func (s *WebSocketServer) SendWelcomeMessageToAllClients(messageObject IFMessage) {
	// First, remove any closed connections
	s.removeClosedConnections()
	s.connectionsMuShared.RLock()
	defer s.connectionsMuShared.RUnlock()
	for _, conn := range s.connections {
		if conn != nil {
			err := sendMessageToClient(conn, messageObject)
			if err != nil {
				fmt.Println("Error sending welcome message:", err)
			}
		}
	}
}

func (s *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading connection:", err)
		sendErrorMessage(conn, "Error upgrading connection")
		return
	}

	var messageObject IFMessage
	//use refresh token to get user data starts___________________________________
	refreshTokenCookieName := "refresh_token"
	refreshToken, refreshTokenErr := r.Cookie(refreshTokenCookieName)
	if refreshTokenErr != nil {
		messageObject.Text = "Access token is invalid. Please log in again."
		messageObject.Time = getCurrentTime()
		err = sendMessageToClient(conn, messageObject)
		if err != nil {
			fmt.Println("Error sending welcome message:", err)
			return
		}
		return
	}
	claims, parseError := jwtAuth.ParseAndValidateToken(refreshToken.Value)
	if parseError != nil {
		messageObject.Text = "Access token is invalid. Please log in again."
		messageObject.Time = getCurrentTime()
		err = sendMessageToClient(conn, messageObject)
		if err != nil {
			fmt.Println("Error sending welcome message:", err)
			return
		}
		return
	}

	//unique chat connection take a Lock don't allow other thread to write it
	s.connectionsMu.Lock()
	connectionId := uuid.New()
	s.connections[connectionId.String()] = conn
	s.connectionsMu.Unlock()

	fmt.Println("Size of the map:", len(s.connections)) // Print the size

	chatUserName := claims.UserName
	messageObject.Text = "Welcome to the chat room " + chatUserName
	messageObject.Time = getCurrentTime()
	messageObject.User = chatUserName
	s.SendWelcomeMessageToAllClients(messageObject)
	/* Testing purpose only */
	/*
		notificationManager := notification.GetNotifySystemInstance()
		notificationManager.AddNotification("user", chatUserName)
		notificationManager.ReceiveNotification("user", func(data interface{}) {
			if stringValue, ok := data.(string); ok {
				fmt.Printf("Received string notification: %s\n", stringValue)
			}
		})*/

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
	for msgByte := range outgoingMessages {
		// Lock the mutex while iterating through the connections map
		s.connectionsMu.Lock()
		for _, conn := range s.connections {
			var messageObject IFMessage
			if err := json.Unmarshal(msgByte, &messageObject); err != nil {
				fmt.Println("Error unmarshaling JSON:", err)
			} else {
				//fmt.Printf("Received message: %+v\n", messageObject)
			}
			err := sendMessageToClient(conn, messageObject)
			if err != nil {
				fmt.Println("Error sending message:", err)
				// Optionally, you might want to remove the connection from the map here
			}
		}
		s.connectionsMu.Unlock()
	}
}
func getCurrentTime() string {
	// Get the current time
	currentTime := time.Now()
	// Format the current time as a string
	return currentTime.Format("2006-01-02 15:04:05")
}
func sendMessageToClient(conn *websocket.Conn, messageObject IFMessage) error {
	updatedMsgObject := IFMessage{Time: getCurrentTime(), Text: messageObject.Text, User: messageObject.User}
	jsonData, err := json.Marshal(updatedMsgObject)
	if err != nil {
		return err
	}

	err = conn.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		return err
	}

	return nil
}
