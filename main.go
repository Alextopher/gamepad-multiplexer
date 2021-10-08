package main

import (
	"log"
	"runtime"
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"
)

func main() {
	runtime.LockOSThread()
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	readConfig("configs/test.yaml")
	states := make(map[int]glfw.GamepadState)
	joysticks := joysticksInit()
	multiplexed := glfw.GamepadState{}

	for {
		glfw.PollEvents()
		states[0] = *joysticks[0].GetGamepadState()
		multiplex(states, &multiplexed)
		log.Print(multiplexed)
		time.Sleep(100 * time.Millisecond)
	}
}
