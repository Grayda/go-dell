go-dell
=======

"Go Banana!" - Ralph Wiggum

go-dell is a small Golang library that lets you control some Dell projectors via Ethernet. It may also work with other projectors that have a Crestron web UI built in.

Compatibility
=============

go-dell has been tested with the following projectors:

 - Dell s300wi
 - Dell s500wi

Not all commands will be available on all projectors, and not all commands may work as intended due to differences in hardware. If you find a projector that does / doesn't work that isn't on this list, please let me know so this driver and list can be updated.

Usage
=====

Simply import `github.com/Grayda/go-dell`, then call `dell.Init()` to prepare, `dell.AddProjector("yourname", "192.168.1.2")` to connect to a projector, then finally `dell.SendCommand(dell.Projectors["yourname"], dell.Commands.Power.On)` to tell the projector to turn on. "yourname" can be whatever you like, as long as it's unique

See `tests/main.go` for a full example

List of available commands
==========================

Commands are accessed like so: `dell.Commands.Input.USBViewer`

 * Input
    * VGAA
    * VGAB
    * Composite  
    * SVideo
    * HDMI
    * Wireless
    * USBDisplay
    * USBViewer  

 * Volume
    * Up
    * Down
    * Mute
    * Unmute

 * Power
    * On  
    * Off

 * Menu
    * Menu  
    * Up
    * Down  
    * Left  
    * Right
    * OK

 * Picture
    * Mute
    * Unmute
    * Freeze
    * Unfreeze
    * Contrast
      * Up
      * Down

    * Brightness
      * Up
      * Down

Supporting development
======================

If your projector has an Ethernet port on it, plug it in to your network and try out this code. If it works, please let me know so I can update the compatibility list
