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
		fmt.Println("Error preparing commands. Error is:", err)
	}

	for { // Loop forever
		select { // This lets us do non-blocking channel reads. If we have a message, process it. If not, check for UDP data and loop
		case msg := <-dell.Events:
			switch msg.Name {
			case "ready":
				fmt.Println("Ready to start listening for commands..")

				go dell.Listen()

				if err != nil {
					fmt.Println(err)
				}
			case "projectorfound":

				_, err = dell.AddProjector(msg.ProjectorInfo)
				if err != nil {
					fmt.Println("Error connecting to printer:", err)
					os.Exit(1)
				}
				fmt.Println("Projector was found at " + msg.ProjectorInfo.IP + ". Make: " + msg.ProjectorInfo.Make + ". Model:" + msg.ProjectorInfo.Model + ". Revision:" + msg.ProjectorInfo.Revision)
			case "listening":
				fmt.Println("Listening for projectors via DDDP")
			case "commandsent":
				fmt.Println("Command sent!")
			case "projectorremoved":
				fmt.Println("Projector Removed:", msg.ProjectorInfo.UUID)
			case "projectoradded":
				fmt.Println("Connected to projector. Sending command to turn on the projector..")
				// dell.SendCommand(msg.ProjectorInfo, dell.Commands.Power.On)
				// fmt.Println("Waiting 30 seconds for the projector to turn on..")
				// time.Sleep(time.Second * 30)
				//
				// fmt.Println("Sending command to projector to get status..")
				// dell.GetStatus(msg.ProjectorInfo)
				// time.Sleep(time.Second * 3)
				//
				// fmt.Println("Sending command to set input to VGA A..")
				// dell.SendCommand(msg.ProjectorInfo, dell.Commands.Input.VGAA)
				// fmt.Println("Waiting 3 seconds for the input to change..")
				// time.Sleep(time.Second * 3)
				//
				// fmt.Println("Sending command to set input to VGA B..")
				// dell.SendCommand(msg.ProjectorInfo, dell.Commands.Input.VGAB)
				// fmt.Println("Waiting 3 seconds for the input to change..")
				// time.Sleep(time.Second * 3)
				//
				// fmt.Println("Sending command to set input to HDMI..")

				// fmt.Println("Waiting 3 seconds for the input to change..")
				// time.Sleep(time.Second * 3)
				//
				// fmt.Println("Turning the projector off..")
				for {
					dell.GetStatus(msg.ProjectorInfo)
					time.Sleep(time.Second * 10)
				}

			default:

			}
		}
	}

}
