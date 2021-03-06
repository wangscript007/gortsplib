// +build ignore

package main

import (
	"fmt"
	"net"

	"github.com/aler9/gortsplib"
	"github.com/aler9/gortsplib/rtph264"
)

// This example shows how to generate RTP/H264 frames from a file with Gstreamer,
// create a RTSP client, connect to a server, announce a H264 track and write
// the frames with the TCP protocol.

func main() {
	// open a listener to receive RTP/H264 frames
	pc, err := net.ListenPacket("udp4", "127.0.0.1:9000")
	if err != nil {
		panic(err)
	}
	defer pc.Close()

	fmt.Println("Waiting for a rtp/h264 stream on port 9000 - you can send one with gstreamer:\n" +
		"gst-launch-1.0 filesrc location=video.mp4 ! qtdemux ! video/x-h264" +
		" ! h264parse config-interval=1 ! rtph264pay ! udpsink host=127.0.0.1 port=9000")

	// wait for RTP/H264 frames
	decoder := rtph264.NewDecoderFromPacketConn(pc)
	sps, pps, err := decoder.ReadSPSPPS()
	if err != nil {
		panic(err)
	}
	fmt.Println("stream connected")

	// create a H264 track
	track, err := gortsplib.NewTrackH264(0, sps, pps)
	if err != nil {
		panic(err)
	}

	// connect to the server and start publishing the track
	conn, err := gortsplib.DialPublish("rtsp://localhost:8554/mystream",
		gortsplib.StreamProtocolTCP, gortsplib.Tracks{track})
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	buf := make([]byte, 2048)
	for {
		// read frames from the source
		n, _, err := pc.ReadFrom(buf)
		if err != nil {
			break
		}

		// write frames to the server
		err = conn.WriteFrame(track.Id, gortsplib.StreamTypeRtp, buf[:n])
		if err != nil {
			fmt.Println("connection is closed (%s)", err)
			break
		}
	}
}
