package webrtcServer

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/manishchauhan/dugguGo/models/roomModel"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
)

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

// RTCChannel represents the WebRTC channel with sender and receiver peers
type RTCChannel struct {
	peerConnections map[string]roomModel.PeerConnectionState
	trackLocals     map[string]*webrtc.TrackLocalStaticRTP
}

// WebRTCManager is a struct representing your WebRTC-related functionalities
type WebRTCManager struct {
	channels map[int]RTCChannel
	lock     sync.RWMutex
}

// NewWebRTCManager creates a new WebRTCManager instance
func NewWebRTCManager() *WebRTCManager {
	// Initialize any necessary settings for your WebRTC instance
	return &WebRTCManager{
		channels: make(map[int]RTCChannel),
	}
}

// getDefaultPeerConnectionConfig retrieves the default WebRTC configuration
func (w *WebRTCManager) getDefaultPeerConnectionConfig() *webrtc.Configuration {
	peerConnectionConfig := &webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302", "stun:stun1.l.google.com:19302"},
			},
		},
	}
	return peerConnectionConfig
}

// getMediaEngine retrieves a new instance of the WebRTC MediaEngine
func (w *WebRTCManager) getMediaEngine() *webrtc.MediaEngine {
	mediaEngine := &webrtc.MediaEngine{}
	return mediaEngine
}

// initializePeerConnection initializes a new Peer Connection for every room
func (w *WebRTCManager) initializePeerConnection() *webrtc.PeerConnection {
	peerConnection, err := webrtc.NewAPI(webrtc.WithMediaEngine(w.getMediaEngine())).NewPeerConnection(*w.getDefaultPeerConnectionConfig())
	if err != nil {
		panic(err)
	}
	return peerConnection
}

// dispatchKeyFrame sends a keyframe to all PeerConnections, used everytime a new user joins the call
func (w *WebRTCManager) dispatchKeyFrame(roomid int) {
	w.lock.Lock()
	defer w.lock.Unlock()
	room, roomExists := w.channels[roomid]
	if roomExists {
		for _, Connection := range room.peerConnections {
			for _, receiver := range Connection.PeerConnection.GetReceivers() {
				if receiver.Track() == nil {
					continue
				}

				_ = Connection.PeerConnection.WriteRTCP([]rtcp.Packet{
					&rtcp.PictureLossIndication{
						MediaSSRC: uint32(receiver.Track().SSRC()),
					},
				})
			}
		}
	}
}

// signalPeerConnections updates each PeerConnection so that it is getting all the expected media tracks
func (w *WebRTCManager) signalPeerConnections(roomid int) {
	w.lock.Lock()
	defer func() {
		w.lock.Unlock()
		w.dispatchKeyFrame(roomid)
	}()

	room, roomExists := w.channels[roomid]
	if roomExists {
		attemptSync := func() (tryAgain bool) {
			for ConnectionKey, Connection := range room.peerConnections {
				if Connection.PeerConnection.ConnectionState() == webrtc.PeerConnectionStateClosed {
					delete(room.peerConnections, ConnectionKey)
					return true
				}

				// map of sender we already are seanding, so we don't double send
				existingSenders := map[string]bool{}

				for _, sender := range Connection.PeerConnection.GetSenders() {
					if sender.Track() == nil {
						continue
					}

					existingSenders[sender.Track().ID()] = true

					// If we have a RTPSender that doesn't map to a existing track remove and signal
					if _, ok := room.trackLocals[sender.Track().ID()]; !ok {
						if err := Connection.PeerConnection.RemoveTrack(sender); err != nil {
							return true
						}
					}
				}

				// Don't receive videos we are sending, make sure we don't have loopback
				for _, receiver := range Connection.PeerConnection.GetReceivers() {
					if receiver.Track() == nil {

						continue
					}

					existingSenders[receiver.Track().ID()] = true
				}

				// Add all track we aren't sending yet to the PeerConnection
				for trackID := range room.trackLocals {
					if _, ok := existingSenders[trackID]; !ok {
						if _, err := Connection.PeerConnection.AddTrack(room.trackLocals[trackID]); err != nil {
							return true
						}
					}
				}

				offer, err := Connection.PeerConnection.CreateOffer(nil)
				if err != nil {
					return true
				}

				if err = Connection.PeerConnection.SetLocalDescription(offer); err != nil {
					return true
				}

				offerString, err := json.Marshal(offer)
				if err != nil {
					return true
				}
				//fmt.Println("2**************8", room.peerConnections[i].PeerConnection)
				if err = Connection.Websocket.WriteJSON(&roomModel.IFWebsocketMessage{
					MessageType: Offer,
					RTCPeerID:   Connection.RTCPeerID,
					Data:        string(offerString),
				}); err != nil {
					fmt.Println("Web socket is claosed")
					return true
				}
			}

			return
		}

		for syncAttempt := 0; ; syncAttempt++ {
			if syncAttempt == 25 {
				// Release the lock and attempt a sync in 3 seconds. We might be blocking a RemoveTrack or AddTrack
				go func() {
					time.Sleep(time.Second * 3)
					w.signalPeerConnections(roomid)
				}()
				return
			}

			if !attemptSync() {
				break
			}
		}
	}

}

// Add to list of tracks and fire renegotation for all PeerConnections
func (w *WebRTCManager) addTrack(roomId int, t *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {
	w.lock.Lock()
	defer func() {
		w.lock.Unlock()
		w.signalPeerConnections(roomId)
	}()

	// Create a new TrackLocal with the same codec as our incoming
	trackLocal, err := webrtc.NewTrackLocalStaticRTP(t.Codec().RTPCodecCapability, t.ID(), t.StreamID())
	if err != nil {
		panic(err)
	}
	w.channels[roomId].trackLocals[t.ID()] = trackLocal
	return trackLocal
}

// Remove from list of tracks and fire renegotation for all PeerConnections
func (w *WebRTCManager) removeTrack(roomId int, t *webrtc.TrackLocalStaticRTP) {
	w.lock.Lock()
	defer func() {
		w.lock.Unlock()
		w.signalPeerConnections(roomId)
	}()

	delete(w.channels[roomId].trackLocals, t.ID())
}

// SetupReceiversAndAssignConnections associates each participant in the specified room with a corresponding PeerConnection.
func (w *WebRTCManager) SetupReceiversAndAssignConnections(websocketMessage *roomModel.IFWebsocketMessage, chatRoom roomModel.IFChatRoom, threadSafeWriter *roomModel.ThreadSafeWriter) {
	uuidID := uuid.New()
	RTCPeerID := uuidID.String()
	// If not, initialize a new peer connection
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		log.Print(err)
		return
	}
	// When this frame returns close the PeerConnection

	//defer peerConnection.Close() //nolint
	// Accept one audio and one video track incoming
	for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
		if _, err := peerConnection.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		}); err != nil {
			log.Print(err)
			return
		}
	}

	// Save the new peer connection to the map
	w.lock.Lock()
	channel, exists := w.channels[chatRoom.RoomId]
	if !exists {
		channel = RTCChannel{
			peerConnections: make(map[string]roomModel.PeerConnectionState, 0),
			trackLocals:     make(map[string]*webrtc.TrackLocalStaticRTP),
		}
	}

	channel.peerConnections[RTCPeerID] = roomModel.PeerConnectionState{
		PeerConnection: peerConnection,
		Websocket:      threadSafeWriter,
		RTCPeerID:      RTCPeerID,
	}
	w.channels[chatRoom.RoomId] = channel
	w.lock.Unlock()
	length1 := len(w.channels[chatRoom.RoomId].peerConnections)
	length2 := len(w.channels[chatRoom.RoomId].trackLocals)
	fmt.Println("id........", RTCPeerID)
	fmt.Println("peerConnections........", length1)
	fmt.Println("trackLocals........", length2)
	//OnICECandidate
	peerConnection.OnICECandidate(func(ice *webrtc.ICECandidate) {
		if ice == nil {
			return
		}
		candidateString, err := json.Marshal(ice.ToJSON())
		if err != nil {
			log.Println(err)
			return
		}
		if writeErr := w.WriteJSON(threadSafeWriter, &roomModel.IFWebsocketMessage{
			MessageType: Candidate,
			User:        websocketMessage.User,
			RoomId:      websocketMessage.RoomId,
			Time:        websocketMessage.Time,
			Data:        string(candidateString),
			//ConnectionID: selectedConnectionID,
		}); writeErr != nil {
			log.Println("web socket closed", writeErr)
		}
	})
	// If PeerConnection is closed remove it from global list
	peerConnection.OnConnectionStateChange(func(p webrtc.PeerConnectionState) {
		switch p {
		case webrtc.PeerConnectionStateFailed:
			if err := peerConnection.Close(); err != nil {
				log.Print(err)
			}
		case webrtc.PeerConnectionStateClosed:
			w.signalPeerConnections(chatRoom.RoomId)
		default:
		}
	})
	peerConnection.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		// Create a track to fan out our incoming video to all peers
		trackLocal := w.addTrack(chatRoom.RoomId, t)
		defer w.removeTrack(chatRoom.RoomId, trackLocal)

		buf := make([]byte, 1500)
		for {
			i, _, err := t.Read(buf)
			if err != nil {
				return
			}

			if _, err = trackLocal.Write(buf[:i]); err != nil {
				return
			}
		}
	})
	//println("1----", peerConnection)
	// Signal for the new PeerConnection
	w.signalPeerConnections(chatRoom.RoomId)
	//println("122----", peerConnection)
	//println("122----", peerConnection)
}
func (w *WebRTCManager) AddICECandidate(msgData []byte, roomid int, RTCPeerID string) {
	fmt.Println("RTCPeerCandidate", RTCPeerID)
	candidate := webrtc.ICECandidateInit{}
	if err := json.Unmarshal([]byte(msgData), &candidate); err != nil {
		log.Println(err)
		return
	}
	channel, exists := w.channels[roomid]
	if exists {
		if err := channel.peerConnections[RTCPeerID].PeerConnection.AddICECandidate(candidate); err != nil {
			log.Println("Candidate", err)
			return
		}
	}
}

func (w *WebRTCManager) SetRemoteDescription(msgData []byte, roomid int, RTCPeerID string) {
	fmt.Println("RTCPeerIDAnswer", RTCPeerID)
	answer := webrtc.SessionDescription{}
	if err := json.Unmarshal([]byte(msgData), &answer); err != nil {
		log.Println(err)
		return
	}
	channel, exists := w.channels[roomid]
	if exists {
		rtcChannel := channel.peerConnections[RTCPeerID]
		if err := rtcChannel.PeerConnection.SetRemoteDescription(answer); err != nil {
			log.Println("Answer", err)
			return
		}
	}

}

func (w *WebRTCManager) WriteJSON(threadSafeWriter *roomModel.ThreadSafeWriter, v interface{}) error {
	threadSafeWriter.Lock()
	defer threadSafeWriter.Unlock()
	return threadSafeWriter.Conn.WriteJSON(v)
}

// remove a channel from channel list when user close a room
func (w *WebRTCManager) DeleteChannel(roomId int) {
	if _, exists := w.channels[roomId]; exists {
		delete(w.channels, roomId)
	}
}
