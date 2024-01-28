package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/manishchauhan/dugguGo/models/roomModel"
	"github.com/manishchauhan/dugguGo/webrtcServer"
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

const (
	TextMessage roomModel.EnumMessageType = iota // simple message
	JoinRoom                                     // welcome message when user joins a channel
	LeaveRoom                                    //  message when user leaves a channel
	Request                                      //   request to join a channel
	VideoCall
	Candidate     // Sdpoffer
	Offer         //Offer
	Answer        //Answer
	DeleteChannel // delete a channel webrtc
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

func sendErrorMessage(threadSafeWriter *roomModel.ThreadSafeWriter, errorMsg string) error {
	message := ErrorMessage{Error: errorMsg}
	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return threadSafeWriter.Conn.WriteMessage(websocket.TextMessage, jsonData)
}

type WebSocketServer struct {
	upgrader websocket.Upgrader
	addr     string
	//connections   map[string]*websocket.Conn
	roomList             map[int]roomModel.IFChatRoom //Room List
	oneTwoOne            []roomModel.IFChatUser       //One Two One Two
	connectionsMu        sync.RWMutex
	UserList             []string //all user List
	SelectedConnectionId string
	webRTCInstance       *webrtcServer.WebRTCManager
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
		roomList: make(map[int]roomModel.IFChatRoom),
	}
}

func (s *WebSocketServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.HandleWebSocket)
	fmt.Printf("WebSocket server listening on %s\n", s.addr)
	corsHandler := cors.Default().Handler(mux)
	if s.webRTCInstance == nil {
		s.webRTCInstance = webrtcServer.NewWebRTCManager()
	}
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

func (s *WebSocketServer) addNewChatRoom(roomId int, chatUserName string, threadSafeWriter *roomModel.ThreadSafeWriter) string {
	// Generate a unique ID for the connection
	connectionID := uuid.New()
	connectionIDString := connectionID.String()

	// Create a new ChatUser
	chatUser := roomModel.IFChatUser{
		ChatUser: chatUserName,
		Conn:     threadSafeWriter,
	}
	// Check if the chat room already exists
	room, roomAlreadyExists := s.roomList[roomId]
	if roomAlreadyExists {
		// Add the chat user to the existing room
		room.UserList[connectionIDString] = chatUser
	} else {
		// Create a new chat room and add the user
		newRoom := roomModel.IFChatRoom{
			RoomId:   roomId,
			UserList: map[string]roomModel.IFChatUser{connectionIDString: chatUser},
		}
		s.roomList[roomId] = newRoom
	}
	return connectionIDString
}

func (s *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	unsafeConn, err := s.upgrader.Upgrade(w, r, nil)
	conn := &roomModel.ThreadSafeWriter{Conn: unsafeConn, Mutex: sync.Mutex{}}
	if err != nil {
		fmt.Println("Error upgrading connection:", err)
		sendErrorMessage(conn, "Error upgrading connection")
		return
	}

	// If handshake happens successfully, start a new instance of WebRTC.
	// Only one instance is enough, so you might want to check if s.webRTCInstance is already set.

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
func (s *WebSocketServer) startVideoConference(websocketMessage *roomModel.IFWebsocketMessage, threadSafeWriter *roomModel.ThreadSafeWriter) bool {

	room, roomAlreadyExists := s.roomList[websocketMessage.RoomId]
	if roomAlreadyExists {

		s.webRTCInstance.SetupReceiversAndAssignConnections(websocketMessage, room, threadSafeWriter)
		return true
	}
	return false
}
func (s *WebSocketServer) unmarshalMessage(data []byte, target interface{}) error {
	if err := json.Unmarshal(data, target); err != nil {
		return err
	}
	return nil
}

func (s *WebSocketServer) handleIncoming(wg *sync.WaitGroup, threadSafeWriter *roomModel.ThreadSafeWriter, incomingMessages chan<- []byte, outgoingMessages chan<- []byte) {
	defer wg.Done()
	websocketMessage := &roomModel.IFWebsocketMessage{}
	for {
		_, raw, err := threadSafeWriter.Conn.ReadMessage()

		if err != nil {
			fmt.Println("Error reading message:", err)
			sendErrorMessage(threadSafeWriter, "Error reading message")
			return
		}
		if err := s.unmarshalMessage(raw, &websocketMessage); err != nil {
			fmt.Println(err)
			return
		}
		// check message type
		switch websocketMessage.MessageType {
		case TextMessage:
			s.connectionsMu.Lock()
			messageObject := roomModel.IFMessage{}

			if err := s.unmarshalMessage([]byte(websocketMessage.Data), &messageObject); err != nil {
				log.Println(err)
				return
			}

			s.connectionsMu.Unlock()
			break
		case JoinRoom:
			s.connectionsMu.Lock()
			messageObject := roomModel.IFMessage{}

			if err := s.unmarshalMessage([]byte(websocketMessage.Data), &messageObject); err != nil {
				log.Println(err)
				return
			}
			s.SelectedConnectionId = s.addNewChatRoom(websocketMessage.RoomId, websocketMessage.User, threadSafeWriter)

			s.connectionsMu.Unlock()
			break
		case LeaveRoom:
			s.connectionsMu.Lock() //  message when user leaves a channel
			messageObject := roomModel.IFMessage{}
			if err := s.unmarshalMessage([]byte(websocketMessage.Data), &messageObject); err != nil {
				log.Println(err)
				return
			}
			s.deleteRoom(websocketMessage.RoomId)
			s.connectionsMu.Unlock()
			continue
		case VideoCall:
			s.startVideoConference(websocketMessage, threadSafeWriter)
			continue
		case Candidate:
			if s.webRTCInstance != nil {
				s.webRTCInstance.AddICECandidate([]byte(websocketMessage.Data), websocketMessage.RoomId, websocketMessage.RTCPeerID)
			}

			continue
		case Answer:
			if s.webRTCInstance != nil {

				s.webRTCInstance.SetRemoteDescription([]byte(websocketMessage.Data), websocketMessage.RoomId, websocketMessage.RTCPeerID)
			}
			continue
		case DeleteChannel:
			if s.webRTCInstance != nil {
				s.webRTCInstance.DeleteChannel(websocketMessage.RoomId, websocketMessage.RTCPeerID, threadSafeWriter)
			}
			continue
		default:
			break
		}
		// Send incoming message to outgoingMessages channel
		select {
		case outgoingMessages <- raw:
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
		messageObject := roomModel.IFWebsocketMessage{}
		messageObject.Time = getCurrentTime()

		if err := json.Unmarshal(msgByte, &messageObject); err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			continue // Skip to the next message if unmarshaling fails
		}
		for _, chatRoom := range s.roomList[messageObject.RoomId].UserList {
			err := s.sendMessageToClient(chatRoom.Conn, &messageObject, s.SelectedConnectionId)
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
func (s *WebSocketServer) sendMessageToClient(threadSafeWriter *roomModel.ThreadSafeWriter, messageObject *roomModel.IFWebsocketMessage, SelectedConnectionId string) error {
	threadSafeWriter.Lock()
	defer threadSafeWriter.Unlock()
	messageObject.ConnectionID = SelectedConnectionId
	err := threadSafeWriter.Conn.WriteJSON(messageObject)
	if err != nil {
		return err
	}
	return nil
}
