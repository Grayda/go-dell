package main

import (
	"fmt"
	"time"

	"github.com/Grayda/go-dell"
)

func main() {
	_, err := dell.Init()
	if err != nil {
		fmt.Println("Error!", err)
	}

	_, err = dell.AddProjector("demo", "192.168.1.2")
	if err != nil {
		fmt.Println("Error connecting to printer:", err)
	}

	fmt.Println("Sending command to turn on the projector..")
	dell.SendCommand(dell.Projectors["demo"], dell.Commands.Power.On)
	fmt.Println("Waiting 30 seconds for the projector to turn on..")
	time.Sleep(time.Second * 30)

	fmt.Println("Sending command to set input to VGA A..")
	dell.SendCommand(dell.Projectors["demo"], dell.Commands.Input.VGAA)
	fmt.Println("Waiting 2 seconds for the input to change..")
	time.Sleep(time.Second * 2)

	fmt.Println("Sending command to set input to VGA B..")
	dell.SendCommand(dell.Projectors["demo"], dell.Commands.Input.VGAB)
	fmt.Println("Waiting 2 seconds for the input to change..")
	time.Sleep(time.Second * 2)

	fmt.Println("Sending command to set input to HDMI..")
	dell.SendCommand(dell.Projectors["demo"], dell.Commands.Input.HDMI)
	fmt.Println("Waiting 2 seconds for the input to change..")
	time.Sleep(time.Second * 2)

	fmt.Println("Turning the projector off..")
	dell.SendCommand(dell.Projectors["demo"], dell.Commands.Power.Off)

}
