package main

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"log"
	"runtime"
	"time"
)

var gamepadStates map[uint8]glfw.GamepadState = make(map[uint8]glfw.GamepadState)

func main() {
	runtime.LockOSThread()
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	// Read command line args
	cli := argParse()

	// Read in the configs
	readConfig(cli.Config)

	// Initialize the joystick handlers
	joysticks := joysticksInit()

	if cli.Listen {
		// Run the server to listen for joystick inputs
		go listen(cli.Domain, cli.Port)

		for {
			glfw.PollEvents()
			gamepadStates[0] = *joysticks[0].GetGamepadState()
			multiplexed := multiplex(gamepadStates)
			if cli.Verbose {
				log.Println(multiplexed)
			}
			time.Sleep(Interval)
		}
	} else {
		// Connect to the server
		conn := connect(cli.Domain, cli.Port)

		// Create a counter
		count := 1
		for {
			glfw.PollEvents()
			// Get joystick states
			for i, joy := range joysticks {
				gamepadStates[uint8(i)] = *joy.GetGamepadState()
			}
			// Multiplex the states
			multiplexed := multiplex(gamepadStates)
			if cli.Verbose {
				log.Println(multiplexed)
			}

			// Create the multiplexed packet
			pkt := Packet{
				PacketId:     uint32(count),
				JoystickId:   0,
				GamepadState: multiplexed,
			}

			// Send the packet to the server
			go func() {
				_, err := conn.WriteTo(pkt.Bytes(), conn.RemoteAddr())
				if err != nil {
					log.Fatalln("Failed to send packet due to error:", err)
				}
			}()

			// Wait until trying again
			time.Sleep(Interval)
		}
	}
}
