package main

import (
	"log"
	"sync"

	"github.com/go-gl/glfw/v3.3/glfw"
)

var joystick_interrupts map[glfw.Joystick]chan bool
var joystick_interrupts_mutex sync.RWMutex

func joystickHandler(joy glfw.Joystick) {
	for {
		joystick_interrupts_mutex.RLock()
		select {
		case <-joystick_interrupts[joy]:
			joystick_interrupts_mutex.RUnlock()
			joystick_interrupts_mutex.Lock()
			delete(joystick_interrupts, joy)
			joystick_interrupts_mutex.Unlock()
			return
		default:
			joystick_interrupts_mutex.RUnlock()
			log.Print(joy.GetButtons())
		}
	}
}

// Sets up handlers for joystick connect & disconnect
func joystickCallbacks(joy glfw.Joystick, event glfw.PeripheralEvent) {
	if event == glfw.Connected {
		log.Printf("New device connected! %d\n", joy)
		joystick_interrupts_mutex.Lock()
		joystick_interrupts[joy] = make(chan bool)
		go joystickHandler(joy)
		joystick_interrupts_mutex.Unlock()
	} else if event == glfw.Disconnected {
		log.Printf("Device disconnected! %d\n", joy)
		joystick_interrupts_mutex.Lock()
		if c, exists := joystick_interrupts[joy]; exists {
			c <- true
		}
		joystick_interrupts_mutex.Unlock()
	} else {
		log.Panicf("JoystickCallbacks joystick %d unknown event %d\n", joy, event)
	}
}

func Init() {
	joystick_interrupts = make(map[glfw.Joystick]chan bool)
	glfw.SetJoystickCallback(joystickCallbacks)
}
