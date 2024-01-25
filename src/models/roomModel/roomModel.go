package roomModel

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type EnumMessageType int
type IFroomModel struct {
	Chatroom_id        int    `json:"chatroom_id"`        //room_id
	Chatroom_name      string `json:"chatroom_name"`      //name
	Created_by_user_id int    `json:"created_by_user_id"` //user_id
	Chatroom_details   string `json:"chatroom_details"`   //details
}
type IFSdp struct {
	Sdp string `json:"sdp"`
}
type IFMessage struct {
	Content interface{} `json:"content,omitempty"` //any type of content

}
type IFWebRtcMessage struct {
	Candidate webrtc.ICECandidateInit   `json:"candidate,omitempty"` //
	Sdp       webrtc.SessionDescription `json:"spd,omitempty"`
}
type IFVideoParticipant struct {
	Time        string          `json:"time,omitempty"` // time of message
	MessageType EnumMessageType `json:"messagetype"`
	User        string          `json:"user,omitempty"`   // user
	RoomId      int             `json:"roomid,omitempty"` // Roomid
}
type IFWebsocketMessage struct {
	Time         string          `json:"time,omitempty"` // time of message
	MessageType  EnumMessageType `json:"messagetype"`
	User         string          `json:"user,omitempty"`   // user
	RoomId       int             `json:"roomid,omitempty"` // Roomid
	Data         string          `json:"data"`
	ConnectionID string          `json:"connectionid"` // connection
	RTCPeerID    string          `json:"rtcpeerid"`    //

}

// User List
type IFChatUser struct {
	ChatUser string
	Conn     *ThreadSafeWriter
}

// Room List
type IFChatRoom struct {
	RoomId   int
	UserList map[string]IFChatUser
}

// Helper to make Gorilla Websockets threadsafe
type ThreadSafeWriter struct {
	*websocket.Conn
	sync.Mutex
}
type PeerConnectionState struct {
	PeerConnection *webrtc.PeerConnection
	Websocket      *ThreadSafeWriter
	RTCPeerID      string
}
