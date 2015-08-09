package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

// This file is a bare-bones emulator. It's goal is to announce a fake projector and accept incoming connections
// so that we can fool driver-projector into thinking that we've got a projector on the network, allowing us to test the driver

// ===============================
//           SETTINGS!
// ===============================
// Change these to change the values of your emulated projector
// Check the bottom of this file for more variables

var properties = []string{
	"SDKClass=VideoProjector",
	"UUID=DEADBEEF",
	"Make=DULL",
	"Model=PROJ01",
	"Revision=0.2.0",
}

func main() {
	startTCP()
	// Start our UDP broadcaster
	startUDP()
	// Start our TCP listener

	// Loop forever
	for {

	}
}

func startTCP() {
	// Listen on our port for any incoming connections.
	l, err := net.Listen("tcp", "0.0.0.0:"+tPort)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	// Close the listener when the application closes.

	fmt.Println("Listening on 0.0.0.0:" + tPort)

	go func() {
		for {

			// Listen for an incoming connection.
			conn, err := l.Accept()
			if err != nil {
				fmt.Println("Error accepting: ", err.Error())
				os.Exit(1)
			}

			// Make a buffer to hold incoming data.
			buf := make([]byte, 1024)
			// Read the incoming connection into the buffer.
			_, err = conn.Read(buf)
			if err != nil {
				fmt.Println("Error reading:", err.Error())
			}
			fmt.Println("Reading")
			handleMessage(buf)

		}
	}()

}

func startUDP() {
	fmt.Println("Starting fake projector announcer on Multicast address", mAddr+".", "Projector will announce every", int(announceTime), "seconds")
	addr, err := net.ResolveUDPAddr("udp", mAddr)
	if err != nil {
		log.Fatal(err)
	}
	l, err := net.ListenMulticastUDP("udp", nil, addr)
	// Run this concurrently so we can still accept and parse TCP messages

	for {
		// AMX is a company that specialises in AV control systems. They have a protocol called DDDP (Dynamic Device Discovery Protocol). All packets
		// start with AMXB and are a simple "tag based" design, with each property starting with <-, then the property name, then =, then the value, then a closing >
		var packet = "AMXB"
		for _, p := range properties {
			packet += "<-" + p + ">"
		}
		fmt.Println("Announcing fake projector:", properties)
		l.WriteToUDP([]byte(packet), addr)
		time.Sleep(time.Second * announceTime)
	}

}

// This function takes our message, finds out the last two bytes, then tells us what command was received.
func handleMessage(msg []byte) {
	cID, _ := hex.DecodeString(string(msg[len(msg)-2:]))
	switch string(cID) {

	case "cd13":
		fmt.Println("Input set to VGA-A")
	case "ce13":
		fmt.Println("Input set to VGA-B")
	case "cf13":
		fmt.Println("Input set to Composite")
	case "d013":
		fmt.Println("Input set to S-Video")
	case "d113":
		fmt.Println("Input set to HDMI")
	case "d313":
		fmt.Println("Input set to Wireless")
	case "d413":
		fmt.Println("Input set to USB Display")
	case "d513":
		fmt.Println("Input set to USB Viewer")

	case "fa13":
		fmt.Println("Volume Up")
	case "fb13":
		fmt.Println("Volume Down")
	case "fc13":
		fmt.Println("Volume Muted")
	case "fd13":
		fmt.Println("Volume Unmuted")

	case "0400":
		fmt.Println("Power On")
	case "0500":
		fmt.Println("Power Off")

	case "1d14":
		fmt.Println("Menu button")
	case "1e14":
		fmt.Println("Menu Up")
	case "1f14":
		fmt.Println("Menu Down")
	case "2014":
		fmt.Println("Menu Left")
	case "2114":
		fmt.Println("Menu Right")
	case "2314":
		fmt.Println("OK button")

	case "ee13":
		fmt.Println("Picture Muted")
	case "ef13":
		fmt.Println("Picture Unmuted")
	case "f013":
		fmt.Println("Picture Frozen")
	case "f113":
		fmt.Println("Picture Unmuted")
	case "f613":
		fmt.Println("Contrast Up")
	case "f713":
		fmt.Println("Contrast Down")
	case "f513":
		fmt.Println("Brightness Up")
	case "f413":
		fmt.Println("Brightness Down")

	}

}

// ===============================
//        OTHER SETTINGS!
// ===============================
// Don't change these unless necessary

var mAddr = "239.255.250.250:9131"
var tPort = "41794"
var maxDatagramSize = 8192
var announceTime time.Duration = 30 // Wait how many seconds before announcing again?
