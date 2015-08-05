package dell

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
)

// Projector holds information about our Projectors
type Projector struct {
	Name string
	Conn net.Conn
	IP   string
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

// Init gets the ball rolling by unmarshalling our command JSON and initializing our Projectors map
func Init() (bool, error) {
	Projectors = make(map[string]Projector)
	err := json.Unmarshal(commandList, &Commands)
	if err != nil {
		return false, err
	}

	return true, nil
}

// AddProjector adds a projector <name> to our Projectors list, and connects to the specified IP address
func AddProjector(name string, ip string) (bool, error) {

	tmp, err := net.Dial("tcp", ip+":4179")
	if err != nil {
		fmt.Println(err)
	}

	Projectors[name] = Projector{
		Name: name,
		IP:   ip,
		Conn: tmp,
	}

	return true, nil
}

// RemoveProjector does what it says on the tin: Removes a projector from our list (after first closing the connection)
func RemoveProjector(name string) (bool, error) {
	Projectors[name].Conn.Close()
	delete(Projectors, name)
	return true, nil
}

// SendCommand issues a command to a projector
func SendCommand(projector Projector, command string) (bool, error) {
	buf, _ := hex.DecodeString("05000600000300" + command)
	_, _ = projector.Conn.Write(buf)
	return true, nil
}
