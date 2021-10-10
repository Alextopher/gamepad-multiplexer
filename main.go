package main

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-vgo/robotgo"
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

	// Initialize the joystick handlers
	joysticks := joysticksInit()

	// Virtual Gamepad
	var multiplexed glfw.GamepadState
	// TODO Next line not needed once we call `multiplex`
	multiplexed.Axes = [6]float32{0, 0, 0, 0, -1, -1}

	if cli.Listen {
		// Read in the configs
		rules, buttonMap, axisMap := readConfig(cli.Config)

		// Run the server to listen for joystick inputs
		go listen(cli.Domain, cli.Port, rules)

		for {
			glfw.PollEvents()
			// TODO : multiplexed = multiplex(rules, gamepadStates, &multiplexed)
			if cli.Verbose {
				log.Println(multiplexed)
			}

			// Button events
			for i := 0; i < len(multiplexed.Buttons); i++ {
				button := glfw.GamepadButton(i)
				if rule, exists := buttonMap[button]; exists {
					if multiplexed.Buttons[button] == glfw.Press {
						robotgo.KeyDown(rule.Key0)
					} else {
						robotgo.KeyUp(rule.Key0)
					}
				}
			}

			// Joystick events
			for _, axis := range JOYSTICK_AXES {
				if rule, exists := axisMap[axis]; exists {
					if multiplexed.Axes[axis] > 0 {
						// positives "right or down"
						robotgo.KeyUp(rule.Key0)
						robotgo.KeyDown(rule.Key1)
					} else if multiplexed.Axes[axis] < 0 {
						// negatives "left or up"
						robotgo.KeyDown(rule.Key0)
						robotgo.KeyUp(rule.Key1)
					} else {
						// zero "nothing"
						robotgo.KeyUp(rule.Key0)
						robotgo.KeyUp(rule.Key1)
					}
				}
			}

			// Trigger events
			for _, axis := range TRIGGER_AXES {
				if rule, exists := axisMap[axis]; exists {
					if multiplexed.Axes[axis] != -1 {
						robotgo.KeyDown(rule.Key0)
					} else {
						robotgo.KeyUp(rule.Key0)
					}
				}
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
				if joy.Present() {
					gamepadStates[uint8(i)] = *joy.GetGamepadState()
				}
			}
			// Multiplex the states
			// TODO : multiplexed = multiplex(rules, gamepadStates, &multiplexed)
			if cli.Verbose {
				log.Println(multiplexed)
			}

			// Create the multiplexed packet
			pkt := GamestatePacket{
				PacketId:     uint32(count),
				JoystickId:   0,
				GamepadState: multiplexed,
			}
			// Update the packet count
			count++

			// Send the packet to the server
			go func() {
				_, err := conn.Write(pkt.Bytes())
				if err != nil {
					log.Fatalln("Failed to send packet due to error:", err)
				}
			}()

			// Wait until trying again
			time.Sleep(Interval)
		}
	}
}
