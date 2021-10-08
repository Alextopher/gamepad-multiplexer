package main

import (
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

	Init()
	for {
		glfw.PollEvents()
		time.Sleep(10 * time.Millisecond)
	}
}
