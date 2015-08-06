package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Grayda/go-dell"
)

var ready bool

func main() {
	fmt.Println("Preparing commands..")
	_, err := dell.Init()
	if err != nil {
		fmt.Println("Error!", err)
	}

	for { // Loop forever
		select { // This lets us do non-blocking channel reads. If we have a message, process it. If not, check for UDP data and loop
		case msg := <-dell.Events:
			switch msg.Name {
			case "ready":
				fmt.Println("Ready to go!")
				_, err = dell.Listen()
				if err != nil {
					fmt.Println(err)
				}
			case "projectorfound":
				_, err = dell.AddProjector(msg.ProjectorInfo)
				if err != nil {
					fmt.Println("Error connecting to printer:", err)
					os.Exit(1)
				}
			case "listening":
				fmt.Println("YES!!")
			case "projectoradded":
				fmt.Println("Sending command to turn on the projector..")
				dell.SendCommand(msg.ProjectorInfo, dell.Commands.Power.On)
				fmt.Println("Waiting 30 seconds for the projector to turn on..")
				time.Sleep(time.Second * 30)

				fmt.Println("Sending command to set input to VGA A..")
				dell.SendCommand(msg.ProjectorInfo, dell.Commands.Input.VGAA)
				fmt.Println("Waiting 3 seconds for the input to change..")
				time.Sleep(time.Second * 3)

				fmt.Println("Sending command to set input to VGA B..")
				dell.SendCommand(msg.ProjectorInfo, dell.Commands.Input.VGAB)
				fmt.Println("Waiting 3 seconds for the input to change..")
				time.Sleep(time.Second * 3)

				fmt.Println("Sending command to set input to HDMI..")
				dell.SendCommand(msg.ProjectorInfo, dell.Commands.Input.HDMI)
				fmt.Println("Waiting 3 seconds for the input to change..")
				time.Sleep(time.Second * 3)

				fmt.Println("Turning the projector off..")
				dell.SendCommand(msg.ProjectorInfo, dell.Commands.Power.Off)
			default:

			}
		}
	}

}
