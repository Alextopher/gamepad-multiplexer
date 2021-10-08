package main

import (
  "log"
	"runtime"

	"github.com/go-gl/glfw/v3.3/glfw"
)


func getJoysticks() (joys []glfw.Joystick) {
  for i := 0; i < 16; i++ {
    joy := glfw.Joystick(i)
    if joy.Present() {
      joys = append(joys, joy)
    }
  }
  return
}

func joystick(joy *glfw.Joystick) {
	axes := joy.GetAxes()
	buttons := joy.GetButtons()

  log.Println(axes)
  log.Println(buttons)
  log.Println(joy.Present(), joy.IsGamepad(), joy.GetGUID())
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

  joys := getJoysticks()
  for _, joy := range joys {
    joystick(&joy)
  }
}
