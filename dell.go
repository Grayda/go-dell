package dell

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strings"
	"time"
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
// Is there a neater way to do this?
type Projector struct {
	Conn     net.Conn
	IP       string
	Name     string
	UUID     string // AKA MAC Address
	Model    string
	Make     string
	Revision string
	// Properties
	PowerState   bool
	VolumeMuted  bool
	PictureMuted bool
	Frozen       bool
	Volume       int
	Contrast     int
	Brightness   int
	Location     string
	Resolution   string
	LampHours    string
	Error        string
	Source       string
}

// Command is a struct that allows us to Unmarshal some JSON into it. This in turn allows us to use dot notation, such as: SendCommand(printer, dell.Commands.Volume.Up)
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

// CommandList is a JSON object that contains
// the various hex codes required to control the projector
// It's exported so that you can overwrite it if necessary
var CommandList = []byte(`
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

// PropertyList contains a list of properties present in the DCE/RPC packet that comes back.
// This list helps us parse the packet and extract info like lamp hours and such back from it.
var PropertyList = []byte(`

	{
		"Power": "138a",
		"Lamp": "138b",
		"Input": "13af",
		"MAC": "13b4",
		"Name": "13ba",
		"Assigned": "13bb"
		"Location": "13bc",
		"Position": "13bd",
		"Resolution": "13be",
		"Firmware": "13bf"
	}

`)

// Commands is a list of commands we can use
var Commands Command

// Projectors is a map that contains all of the projectors we've added
var Projectors map[string]Projector

// UDP and TCP connections for reading and writing
var udpConn *net.UDPConn
var udpAddr *net.UDPAddr

// commandPrefix
var commandPrefix = "05000600000300"

// Init gets the ball rolling by unmarshalling our command JSON and initializing our Projectors map
func Init() (bool, error) {
	Projectors = make(map[string]Projector)
	err := json.Unmarshal(CommandList, &Commands)
	if err != nil {
		return false, err
	}

	// Tell our calling code that we're ready!
	passMessage("ready", Projector{})

	return true, nil

}

// Listen listens on port 9131 for projectors
func Listen() (bool, error) {
	var err error
	// Resolve our address, ready for listening. We're listening on port 9131 on the multicast address below
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

	passMessage("listening", Projector{})
	// Because we need to be on the lookout for incoming projector discovery packets, we run
	// this in a goroutine and loop forever

	for {

		readUDP()
		// for _, p := range Projectors {
		// 	go readTCP(p)
		// }

	}

	return true, nil
}

// AddProjector adds a projector <name> to our Projectors list, and connects to the specified IP address
func AddProjector(projector Projector) (bool, error) {

	// Does this projector already exist?
	_, exists := Projectors[projector.UUID]

	// Yes?
	if exists == true {
		// Return. We're not interested
		return false, nil
	}

	// Connect to the projector
	tmp, err := net.Dial("tcp", projector.IP+":41794")
	if err != nil {
		fmt.Println(err)
	}

	// Add the projector to our list.
	Projectors[projector.UUID] = Projector{
		UUID:  projector.UUID,
		Name:  projector.UUID, // Because we don't know the name yet, but we do know the UUID
		Make:  projector.Make,
		Model: projector.Model,
		IP:    projector.IP,
		Conn:  tmp,
	}

	passMessage("projectoradded", Projectors[projector.UUID])

	go func() {
		for {
			readTCP(Projectors[projector.UUID])
			if isDisconnected(Projectors[projector.UUID]) {
				RemoveProjector(Projectors[projector.UUID])
			}
		}
	}()

	return true, nil
}

// RemoveProjector does what it says on the tin: Removes a projector from our list (after first closing the connection)
func RemoveProjector(projector Projector) (bool, error) {
	projector.Conn.Close()
	passMessage("projectorremoved", projector)
	delete(Projectors, projector.UUID)
	return true, nil
}

// SendCommand issues a command to a projector
func SendCommand(projector Projector, command string) (bool, error) {
	fmt.Println("Sending Message to", projector.IP, ":", commandPrefix+command)
	SendRaw(commandPrefix+command, projector)
	passMessage("commandsent", projector)
	return true, nil
}

// SendRaw sends raw data
func SendRaw(msg string, projector Projector) {
	buf, _ := hex.DecodeString(msg)
	_, _ = projector.Conn.Write(buf)

}

func readUDP() (bool, error) { // Now we're checking for messages

	var msg []byte // Holds the incoming message
	buf := make([]byte, 1024)

	var success bool
	var err error

	n, addr, err := udpConn.ReadFromUDP(buf[:]) // Read 1024 bytes from the buffer
	buf2 := buf[:n]
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if n > 0 { // If we've got more than 0 bytes and it's not from us
		fmt.Println("MES!")
		msg = buf2[0:n]

		// If our message is an AMXB message (a.k.a DDDP, a.k.a Dynamic Device Discovery Protocol)
		if strings.Contains(string(msg), "VideoProjector") {

			// Regex was forged by Lucifer himself. If you can craft a better regex to search for what we need, FOR THE LOVE OF GO, CREATE A PULL REQUEST!
			// To break this down: We start by looking for UUID=, then make a named group called UUID, which holds the contents of whatever comes after UUID=, up until the closing >
			// We then repeat this for Make, Model and SDKClass. Finally, g makes it global so it won't stop after finding the first match
			r, _ := regexp.Compile("<-([^=]+)=([^>]+)>")
			result := make(map[string]string)

			for _, p := range r.FindAllStringSubmatch(string(msg), -1) {
				result[p[1]] = p[2]
			}
			//
			// 	result[name] = match[i]
			// }

			// (this lets us check to see if we have this printer in our list)
			_, ok := Projectors[result["UUID"]]

			// And if this printer isn't in our list
			if ok != true {
				//	fmt.Println("Adding new projector. UUID is " + result["UUID"] + ", Model is " + result["Model"] + ", Make is " + result["Make"] + " and IP is " + addr.IP.String())
				// Add it. The key will be the printer name, the value, a Printer struct
				tmp := Projector{
					UUID:     result["UUID"],
					Model:    result["Model"],
					Make:     result["Make"],
					Revision: result["Revision"],
					IP:       addr.IP.String(),
				}

				passMessage("projectorfound", tmp)

			} else {

			}
		}
		fmt.Println("OK")
		msg = nil // Clear out our msg property so we don't run handleMessage on old data

	}

	return success, err
}

func readTCP(projector Projector) (bool, error) { // Now we're checking for messages

	buf := make([]byte, 4098)

	var success bool
	var err error

	n, err := projector.Conn.Read(buf) // Read 1024 bytes from the buffer

	buf2 := buf[:n]
	if err != nil && n > 0 {
		fmt.Println(err)
		os.Exit(1)
	}
	if n > 0 { // If we've got more than 0 bytes and it's not from us

		msg := hex.EncodeToString(buf2)

		handleMessage(msg, projector)

		// fmt.Println(string("Message!: "), msg, n)
		// 050005000002031d

		// If our message is an AMXB message (a.k.a DDDP, a.k.a Dynamic Device Discovery Protocol)

		msg = "" // Clear out our msg property so we don't run handleMessage on old data
	}

	return success, err
}

// GetStatus asks the projector for everything it knows. It's returned as a DCE/RPC packet
func GetStatus(projector Projector) {
	SendRaw("050005000002031e", projector)

}

// This function parses the RPC message we get back from GetStatus and updates 'projector' accordingly.
// Each "line" from our RPC packet is separated by "03" (in hex) and the end of the readable data is separated by a "05" (again, in hex)
// So to retrieve your data, you split the whole packet up by 03, loop through and find a line that ends with the ID you want (e.g. a line that has "13b4" at the end will contain your MAC)
// then you split that line up by 05, and the first part is your MAC address.
// The exception to the rule is the firmware revision, which has the ID at the START of the line. I'd check for this, but different projectors might put different properties last
func handleMessage(msg string, projector Projector) {
	// Holds the IDs we're looking for
	var ids map[string]string

	json.Unmarshal(PropertyList, &ids)

	parts := strings.Split(msg, "03")
	for _, p := range parts {
		for c := range ids {
			if strings.HasSuffix(p, ids[c]) {
				// Now, we could use the reflect library to dynamically set these properties, but I'll try that later.
				data := strings.Split(p, "05")[0]
				dec, _ := hex.DecodeString(data)
				switch ids[c] {
				case "Input":
					projector.Source = string(dec)
				case "Power":
					if string(dec) == "On" {
						projector.PowerState = true
					} else {
						projector.PowerState = false
					}
				case "Name":
					projector.Name = string(dec)
					passMessage("namechanged", projector)
				case "Lamp":
					projector.LampHours = string(dec)
				}
			}
		}
	}

}

func isDisconnected(projector Projector) bool {
	fmt.Println("Checking for disconnection")
	one := make([]byte, 1)
	projector.Conn.SetReadDeadline(time.Now())
	if _, err := projector.Conn.Read(one); err == io.EOF {
		projector.Conn.SetReadDeadline(time.Time{})
		return true
	}

	projector.Conn.SetReadDeadline(time.Time{})
	return false

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
