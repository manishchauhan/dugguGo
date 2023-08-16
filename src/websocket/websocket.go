package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
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
	upgrader websocket.Upgrader
	addr     string
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
		upgrader: upgrader,
		addr:     addr,
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
	defer conn.Close()

	// Send welcome message to client
	err = sendMessageToClient(conn, "Welcome to the Chat Center")
	if err != nil {
		fmt.Println("Error sending welcome message:", err)
		return
	}

	incomingMessages := make(chan []byte)
	outgoingMessages := make(chan []byte)

	var wgIncoming sync.WaitGroup
	var wgOutgoing sync.WaitGroup

	wgIncoming.Add(1)
	wgOutgoing.Add(1)

	go s.handleIncoming(&wgIncoming, conn, incomingMessages, outgoingMessages)
	go s.handleOutgoing(&wgOutgoing, conn, outgoingMessages)

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

func (s *WebSocketServer) handleOutgoing(wg *sync.WaitGroup, conn *websocket.Conn, outgoingMessages <-chan []byte) {
	defer wg.Done()
	for msg := range outgoingMessages {
		err := sendMessageToClient(conn, string(msg))
		if err != nil {
			fmt.Println("Error sending message:", err)
			return
		}
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
