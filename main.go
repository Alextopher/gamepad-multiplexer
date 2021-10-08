package main

import (
  "log"
	"runtime"

	"github.com/go-gl/glfw/v3.3/glfw"
)

func joystick() {
  // clear state
	joy := glfw.Joystick(glfw.Joystick1)
  log.Println(joy)

	axes := joy.GetAxes()
	buttons := joy.GetButtons()

  log.Println(axes)
  log.Println(buttons)
}

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

func main() {
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

  joystick()
}
