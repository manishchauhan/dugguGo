package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/rs/cors"
)

/*
Task done
================================================================
//next video and audio call - next target
//aws task
//share html
//share format text
//share images and videos (all formats) //mutiple videos
//share hyperlinks

pending:-
1. Remove a room and all its connection when Room is deleted at frontend
2. Code refactor
3. Change roomname to roomid as its helps to mantain connection when we edit room
4. optimize connection and gorutines to handle more and more connection
5. send user name active list-1
6. one two one chat option-2
*/
type EnumMessageType int

const (
	TextMessage  EnumMessageType = iota // simple message
	JoinRoom                            // welcome message when user joins a channel
	LeaveRoom                           //  message when user leaves a channel
	Request                             //   request to join a channel
	videoRequest                        // video call
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
	Time         string          `json:"time"`         //time of message
	Text         string          `json:"text"`         //text
	User         string          `json:"user"`         //user
	RoomId       int             `json:"roomid"`       //Roomid
	MessageType  EnumMessageType `json:"messagetype"`  //
	ConnectionID string          `json:"connectionid"` //connection
}

// User List
type IFChatUser struct {
	ChatUser string
	Conn     *websocket.Conn
}

// Room List
type IFChatRoom struct {
	RoomId   int
	UserList map[string]IFChatUser
}

type WebSocketServer struct {
	upgrader websocket.Upgrader
	addr     string
	//connections   map[string]*websocket.Conn
	roomList             map[int]IFChatRoom //Room List
	oneTwoOne            []IFChatUser       //One Two One Two
	connectionsMu        sync.RWMutex
	UserList             []string //all user List
	SelectedConnectionId string
	//webRTCInstance       *webrtc.WebRTC
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
		UserList: make([]string, 0),
		roomList: make(map[int]IFChatRoom),
	}
}

func (s *WebSocketServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.HandleWebSocket)
	fmt.Printf("WebSocket server listening on %s\n", s.addr)
	corsHandler := cors.Default().Handler(mux)
	return http.ListenAndServe(s.addr, corsHandler)
}

// Check if connection is Already established and exits in map
func (s *WebSocketServer) IsConnectionAlreadyExists(roomId int, connectionKey string) bool {
	room, existsRoom := s.roomList[roomId]
	if !existsRoom {
		return false
	}
	_, existsConnection := room.UserList[connectionKey]
	if existsConnection {
		return true
	}
	return true
}

// Delete room and delete all connection realted to it
func (s *WebSocketServer) deleteRoom(roomId int) {
	_, exists := s.roomList[roomId]
	if exists {
		delete(s.roomList, roomId)
	}
}

// Delete a connection
func (s *WebSocketServer) deleteAndClosedConnection(roomId int, connectionKey string) bool {
	room, exists := s.roomList[roomId]
	if !exists {
		return false
	}
	// Now, you can safely delete the connection
	room.UserList[connectionKey].Conn.Close()
	delete(room.UserList, connectionKey)
	return true
}
func (s *WebSocketServer) removeClosedConnections(roomId int) {
	s.connectionsMu.Lock()
	defer s.connectionsMu.Unlock()
	for connectionKey, chatRoom := range s.roomList[roomId].UserList {
		Conn := chatRoom.Conn
		if Conn != nil {
			if err := Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					// Handle unexpected close errors
					fmt.Println("Unexpected error while handling connection:", err)
				}
				// Connection is closed, remove it from the map
				delete(s.roomList[roomId].UserList, connectionKey)
			}
		}
	}
}

func (s *WebSocketServer) addNewChatRoom(roomId int, chatUserName string, conn *websocket.Conn) string {
	// Generate a unique ID for the connection
	connectionID := uuid.New()
	connectionIDString := connectionID.String()

	// Create a new ChatUser
	chatUser := IFChatUser{
		ChatUser: chatUserName,
		Conn:     conn,
	}
	// Check if the chat room already exists
	room, roomAlreadyExists := s.roomList[roomId]
	if roomAlreadyExists {
		// Add the chat user to the existing room
		room.UserList[connectionIDString] = chatUser
	} else {
		// Create a new chat room and add the user
		newRoom := IFChatRoom{
			RoomId:   roomId,
			UserList: map[string]IFChatUser{connectionIDString: chatUser},
		}
		s.roomList[roomId] = newRoom
	}
	return connectionIDString
}

func (s *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading connection:", err)
		sendErrorMessage(conn, "Error upgrading connection")
		return
	}
	/*if handshake happen successfully, start a new instance of webrtc only one instance is enough*/
	//s.webRTCInstance = webrtc.NewWebRTC()
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
func (s *WebSocketServer) startVideoConference() {
	//s.webRTCInstance.AddNewPeerConnection()
}
func (s *WebSocketServer) handleIncoming(wg *sync.WaitGroup, conn *websocket.Conn, incomingMessages chan<- []byte, outgoingMessages chan<- []byte) {
	defer wg.Done()
	for {

		_, p, err := conn.ReadMessage()
		var messageObject IFMessage
		if err != nil {
			fmt.Println("Error reading message:", err)
			sendErrorMessage(conn, "Error reading message")
			return
		}
		json.Unmarshal(p, &messageObject)
		// check message type

		switch messageObject.MessageType {
		case TextMessage:
			s.connectionsMu.Lock()
			if messageObject.ConnectionID != "" {
				s.SelectedConnectionId = messageObject.ConnectionID
			}
			s.connectionsMu.Unlock()
			break
		case JoinRoom:
			s.connectionsMu.Lock()
			s.SelectedConnectionId = s.addNewChatRoom(messageObject.RoomId, messageObject.User, conn)
			s.connectionsMu.Unlock()
			break
		case LeaveRoom:
			s.connectionsMu.Lock() //  message when user leaves a channel
			s.deleteRoom(messageObject.RoomId)
			s.connectionsMu.Unlock()
			continue
		case videoRequest:
			//s.startVideoConference()
		default:
			break
		}
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
		var messageObject IFMessage
		if err := json.Unmarshal(msgByte, &messageObject); err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
		} else {
			//fmt.Printf("Received message: %+v\n", messageObject)
		}
		for _, chatRoom := range s.roomList[messageObject.RoomId].UserList {
			err := sendMessageToClient(chatRoom.Conn, messageObject, s.SelectedConnectionId)
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
func sendMessageToClient(conn *websocket.Conn, messageObject IFMessage, SelectedConnectionId string) error {
	updatedMsgObject := IFMessage{ConnectionID: SelectedConnectionId, MessageType: messageObject.MessageType, Time: getCurrentTime(), Text: messageObject.Text, User: messageObject.User, RoomId: messageObject.RoomId}
	jsonData, err := json.Marshal(updatedMsgObject)
	if err != nil {
		return err
	}
	//fmt.Println("messageObject.RoomName", conn)
	err = conn.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		return err
	}
	return nil
}
