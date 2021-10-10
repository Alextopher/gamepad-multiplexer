package main

import (
	"io/ioutil"
	"log"

	"github.com/go-gl/glfw/v3.3/glfw"
)

// Joysticks and triggers behave fairly differently
var (
	JOYSTICK_AXES = [4]glfw.GamepadAxis{glfw.AxisLeftX, glfw.AxisLeftY, glfw.AxisRightX, glfw.AxisLeftY}
	TRIGGER_AXES  = [2]glfw.GamepadAxis{glfw.AxisLeftTrigger, glfw.AxisRightTrigger}
)

// Sets up handlers for joystick connect & disconnect
func joystickCallbacks(joy glfw.Joystick, event glfw.PeripheralEvent) {
	if event == glfw.Connected {
		if !joy.IsGamepad() {
			log.Printf("ERROR: Connected device is not supported! Check for updates to gamecontrollerdb.txt")
		} else {
			log.Printf("New %s connected! Device %d", joy.GetGamepadName(), joy)
		}
	} else if event == glfw.Disconnected {
		log.Printf("Device %d disconnected!\n", joy)
	} else {
		log.Panicf("JoystickCallbacks joystick %d unknown event %d\n", joy, event)
	}
}

func joysticksInit() (joysticks [16]glfw.Joystick) {
	content, err := ioutil.ReadFile("gamecontrollerdb.txt")
	if err != nil {
		log.Panic("Could not read contents of gamecontrollerdb.txt", err)
	}
	glfw.UpdateGamepadMappings(string(content))

	glfw.SetJoystickCallback(joystickCallbacks)
	for i := 0; i < 16; i++ {
		joysticks[i] = glfw.Joystick(i)
	}

	return joysticks
}
