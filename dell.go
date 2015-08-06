package dell

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

// EventStruct is our equivalent to node.js's Emitters, of sorts.
// This basically passes back to our Event channel, info about what event was raised
// (e.g. Device, plus an event name) so we can act appropriately
type EventStruct struct {
	Name          string
	ProjectorInfo Projector
}

// Events is our events channel which will notify calling code that we have an event happening
var Events = make(chan EventStruct, 1)

// Projector holds information about our Projectors
type Projector struct {
	UUID  string
	Model string
	Make  string
	Conn  net.Conn
	IP    string
}

// Command is a struct that allows us to Unmarshal some JSON into it. This in turn allows us to use dot notation, such as: SendCommand(printer, Commands.Volume.Up)
type Command struct {
	Input struct {
		VGAA       string
		VGAB       string
		Composite  string
		SVideo     string
		HDMI       string
		Wireless   string
		USBDisplay string
		USBViewer  string
	}
	Volume struct {
		Up     string
		Down   string
		Mute   string
		Unmute string
	}
	Power struct {
		On  string
		Off string
	}
	Menu struct {
		Menu  string
		Up    string
		Down  string
		Left  string
		Right string
		OK    string
	}
	Picture struct {
		Mute     string
		Unmute   string
		Freeze   string
		Unfreeze string
		Contrast struct {
			Up   string
			Down string
		}
		Brightness struct {
			Up   string
			Down string
		}
	}
}

var commandList = []byte(`
  {
	"Input": {
		"Vgaa": "cd13",
		"Vgab": "ce13",
		"Composite": "cf13",
		"Svideo": "d013",
		"Hdmi": "d113",
		"Wireless": "d313",
		"Usbdisplay": "d413",
		"Usbviewer": "d513"
	},
	"Volume": {
		"Up": "fa13",
		"Down": "fb13",
		"Mute": "fc13",
		"Unmute": "fd13"
	},
	"Power": {
		"On": "0400",
		"Off": "0500"
	},
	"Menu": {
		"Menu": "1d14",
		"Up": "1e14",
		"Down": "1f14",
		"Left": "2014",
		"Right": "2114",
		"OK": "2314"
	},
	"Picture": {
		"Mute": "ee13",
		"Unmute": "ef13",
		"Freeze": "f013",
		"Unfreeze": "f113",
		"Contrast": {
			"Up": "f613",
			"Down": "f713"
		},
		"Brightness": {
			"Up": "f513",
			"Down": "f413"
		}
	}
}
`)

// Commands is a list of commands we can use
var Commands Command

// Projectors is a map that contains all of the projectors we've added
var Projectors map[string]Projector

var udpConn *net.UDPConn
var udpAddr *net.UDPAddr

// Init gets the ball rolling by unmarshalling our command JSON and initializing our Projectors map
func Init() (bool, error) {
	Projectors = make(map[string]Projector)
	err := json.Unmarshal(commandList, &Commands)
	if err != nil {
		return false, err
	}

	passMessage("ready", Projector{})

	return true, nil

}

// Listen listens on port 2048 for projectors
func Listen() (bool, error) {
	var err error
	// Resolve our address, ready for listening. We're listening on port 55386
	udpAddr, err = net.ResolveUDPAddr("udp4", "239.255.250.250:9131") // Get our address ready for listening
	if err != nil {
		// Errors. Errors everywhere.
		return false, err
	}

	// Now we're actually listening
	udpConn, err = net.ListenMulticastUDP("udp", nil, udpAddr) // Now we listen on the address we just resolved
	if err != nil {
		return false, err
	}

	go func() {
		for {
			ReadUDP()
		}
	}()

	passMessage("listening", Projector{})

	return true, nil
}

// AddProjector adds a projector <name> to our Projectors list, and connects to the specified IP address
func AddProjector(projector Projector) (bool, error) {

	_, exists := Projectors[projector.UUID]

	if exists == true {
		return false, nil
	}

	tmp, err := net.Dial("tcp", projector.IP+":41794")
	if err != nil {
		fmt.Println(err)
	}

	Projectors[projector.UUID] = Projector{
		Make:  projector.Make,
		Model: projector.Model,
		IP:    projector.IP,
		Conn:  tmp,
	}

	passMessage("projectoradded", Projectors[projector.UUID])

	return true, nil
}

// RemoveProjector does what it says on the tin: Removes a projector from our list (after first closing the connection)
func RemoveProjector(projector Projector) (bool, error) {
	projector.Conn.Close()
	passMessage("projectoradded", projector)
	delete(Projectors, projector.UUID)
	return true, nil
}

// SendCommand issues a command to a projector
func SendCommand(projector Projector, command string) (bool, error) {
	buf, _ := hex.DecodeString("05000600000300" + command)
	_, _ = projector.Conn.Write(buf)
	return true, nil
}

func ReadUDP() (bool, error) { // Now we're checking for messages

	var msg []byte     // Holds the incoming message
	var buf [1024]byte // We want to get 1024 bytes of messages (is this enough? Need to check!)

	var success bool
	var err error

	n, addr, err := udpConn.ReadFromUDP(buf[0:]) // Read 1024 bytes from the buffer
	fmt.Println("O===K")
	if n > 0 { // If we've got more than 0 bytes and it's not from us

		msg = buf[0:n]
		fmt.Println(string(msg))
		// If our message is an AMXB message (a.k.a DDDP, a.k.a Dynamic Device Discovery Protocol)
		if strings.Contains(string(msg), "VideoProjector") {

			// Regex was forged by Lucifer himself. If you can craft a better regex to search for what we need, FOR THE LOVE OF GO, CREATE A PULL REQUEST!
			// To break this down: We start by looking for UUID=, then make a named group called UUID, which holds the contents of whatever comes after UUID=, up until the closing >
			// We then repeat this for Make, Model and SDKClass. Finally, g makes it global so it won't stop after finding the first match
			r, _ := regexp.Compile("UUID=(?P<UUID>(.*?))>|Make=(?P<Make>(.*?))>|Model=(?P<Model>(.*?))>|SDKClass=(?P<SDKClass>(.*?))>")

			match := r.FindAllStringSubmatch(string(msg), -1)
			result := make(map[string]string)
			spew.Dump(match)
			// for i, name := range r.SubexpNames() {
			//
			// 	result[name] = match[i]
			// }

			// (this lets us check to see if we have this printer in our list)
			_, ok := Projectors[result["UUID"]]

			// And if this printer isn't in our list
			if ok != true {
				fmt.Println("Adding new projector. UUID is " + result["UUID"] + ", Model is " + result["Model"] + ", Make is " + result["Make"] + " and IP is " + addr.IP.String())
				// Add it. The key will be the printer name, the value, a Printer struct
				Projectors[result["UUID"]] = Projector{
					UUID:  result["UUID"],
					Model: result["Model"],
					Make:  result["Make"],
					IP:    addr.IP.String(),
				}

				passMessage("projectorfound", Projectors[result["UUID"]])

			} else {
				ReadUDP()
			}
		}

		msg = nil // Clear out our msg property so we don't run handleMessage on old data
	}

	return success, err
}

// passMessage adds items to our Events channel so the calling code can be informed
// It's non-blocking or whatever.
func passMessage(message string, projector Projector) bool {

	select {
	case Events <- EventStruct{message, projector}:

	default:
	}

	return true
}
