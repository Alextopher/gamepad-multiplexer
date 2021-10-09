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

	readConfig("configs/test.yaml")
	joysticks := joysticksInit()

	for {
		glfw.PollEvents()
		gamepadStates[0] = *joysticks[0].GetGamepadState()
		multiplexed := multiplex(gamepadStates)
		log.Println(multiplexed)
		time.Sleep(100 * time.Millisecond)
	}
}
