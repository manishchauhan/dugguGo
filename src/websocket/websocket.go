// websocket/server.go

package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const maxClients = 50
const workerPoolSize = 10

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

type WebSocketServer struct {
	upgrader websocket.Upgrader
	addr     string
}

func NewWebSocketServer(addr string) *WebSocketServer {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	return &WebSocketServer{
		upgrader: upgrader,
		addr:     addr,
	}
}

func (s *WebSocketServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.HandleWebSocket)
	println("i am running as a web socket")
	fmt.Printf("WebSocket server listening on %s\n", s.addr)
	return http.ListenAndServe(s.addr, mux)
}

func (s *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading connection:", err)
		sendErrorMessage(conn, "Error upgrading connection")
		return
	}
	defer conn.Close()

	fmt.Println("WebSocket connection established")

	incomingMessages := make(chan []byte)
	outgoingMessages := make(chan []byte)

	var wg sync.WaitGroup
	for i := 0; i < workerPoolSize; i++ {
		wg.Add(2)
		go s.handleIncoming(&wg, conn, incomingMessages)
		go s.handleOutgoing(&wg, conn, outgoingMessages)
	}

	go s.generateOutgoingMessages(outgoingMessages)

	wg.Wait()
}

func (s *WebSocketServer) generateOutgoingMessages(outgoingMessages chan<- []byte) {
	for i := 0; i < 1000; i++ {
		message := []byte(fmt.Sprintf("Message #%d from server", i))
		outgoingMessages <- message
	}
	close(outgoingMessages)
}

func (s *WebSocketServer) handleIncoming(wg *sync.WaitGroup, conn *websocket.Conn, incomingMessages chan<- []byte) {
	defer wg.Done()
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			sendErrorMessage(conn, "Error reading message")
			return
		}

		incomingMessages <- p
	}
}

func (s *WebSocketServer) handleOutgoing(wg *sync.WaitGroup, conn *websocket.Conn, outgoingMessages <-chan []byte) {
	defer wg.Done()
	for msg := range outgoingMessages {
		err := conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			fmt.Println("Error writing message:", err)
			sendErrorMessage(conn, "Error writing message")
			return
		}
	}
}
