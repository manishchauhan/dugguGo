package webrtc

import (
	"sync"

	"github.com/pion/webrtc/v3"
)

// WebRTC is a struct representing your WebRTC-related functionalities
type WebRTC struct {
	rwMutex           sync.RWMutex
	PeerConnectionMap map[int]*webrtc.PeerConnection // Each room contains one peer connection so if there are 50 rooms
	// we need 50 PeerConnection
}

// NewWebRTC creates a new WebRTC instance
func NewWebRTC() *WebRTC {
	// Initialize any necessary settings for your WebRTC instance
	return &WebRTC{
		PeerConnectionMap: make(map[int]*webrtc.PeerConnection),
	}
}

// getPeerConnectionConfig retrieves the default WebRTC configuration
func (w *WebRTC) getPeerConnectionConfig() *webrtc.Configuration {
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
func (w *WebRTC) getMediaEngine() *webrtc.MediaEngine {
	mediaEngine := &webrtc.MediaEngine{}
	return mediaEngine
}

// AddNewPeerConnection adds a new Peer Connection for every room
func (w *WebRTC) AddNewPeerConnection() {
	peerConnection, err := webrtc.NewAPI(webrtc.WithMediaEngine(w.getMediaEngine())).NewPeerConnection(*w.getPeerConnectionConfig())
	if err != nil {
		panic(err)
	}
	println("A video request would start", peerConnection)
}
