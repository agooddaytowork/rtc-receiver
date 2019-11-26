package main

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v2"
)

type Wrap struct {
	*webrtc.DataChannel
}

func (rtc *Wrap) Write(data []byte) (int, error) {
	err := rtc.DataChannel.Send(data)
	return len(data), err
}

var pc *webrtc.PeerConnection

func hub(ws *websocket.Conn) error {
	var msg Session
	for {
		err := ws.ReadJSON(&msg)
		if err != nil {
			_, ok := err.(*websocket.CloseError)
			if !ok {
				fmt.Println("websocket", err)
			}
			break
		}
		err = startRTC(ws, msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func startRTC(ws *websocket.Conn, data Session) error {
	if data.Error != "" {
		return fmt.Errorf(data.Error)
	}

	switch data.Type {
	case "signal_OK":
		// fmt.Printf("Status: RTC Call\r")
		var err error
		pc, err = webrtc.NewPeerConnection(configRTC)
		if err != nil {
			return err
		}
		pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
			// fmt.Printf("Status: ICE %s\r", state.String())
		})

		// dc, err := pc.CreateDataChannel("SSH", nil)
		dc, err := pc.CreateDataChannel("videostream", nil)

		if err != nil {
			fmt.Println(err)
		}

		// DataChannel(dc, ssh)

		VideoChannel(dc)
		offer, err := pc.CreateOffer(nil)
		if err != nil {
			return err
		}
		err = pc.SetLocalDescription(offer)
		if err != nil {
			return err
		}
		if err = ws.WriteJSON(offer); err != nil {
			return err
		}

	case "answer":
		err := pc.SetRemoteDescription(webrtc.SessionDescription{
			Type: webrtc.SDPTypeAnswer,
			SDP:  data.Sdp,
		})
		if err != nil {
			pc.Close()
			return err
		}
		err = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown signaling message '%v'", data.Type)
	}
	return nil
}

// test commands
func VideoChannel(dc *webrtc.DataChannel) {
	dc.OnOpen(func() {
	})
	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		// fmt.Println(msg.Data)
		// fmt.Print(string(msg.Data))
		binary.Write(os.Stdout, binary.LittleEndian, msg.Data)
	})
	dc.OnClose(func() {
		// fmt.Printf("Close video socket\n")

	})
}
