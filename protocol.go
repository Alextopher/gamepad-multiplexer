package main

import (
	"errors"
	"github.com/go-gl/glfw/v3.3/glfw"
	"net"
	"time"
)

const Interval time.Duration = 100 * time.Millisecond

type Packet struct {
	PacketId     uint32
	JoystickId   uint8
	GamepadState glfw.GamepadState
	LAddr        net.Addr
	RAddr        net.Addr
}

// Parse the data from the packet and convert to valid struct
func (p *Packet) Parse(data []byte) error {
	// Bad packet length
	if len(data) != 31 {
		return errors.New("invalid packet length")
	}

	pos := 0
	// Get the packet id by shifting each byte into the appropriate place of the uint32
	for i := 0; i < 4; i++ {
		p.PacketId |= uint32(uint32(data[i]) << (8 * (3 - i)))
		pos++
	}

	// Get the joystick id
	p.JoystickId = data[pos]
	pos++

	// Get the buttons by turning each bit into the correct position in the array
	for i := 0; i < 15; i++ {
		p.GamepadState.Buttons[i] = glfw.Action((data[pos+int(i/8)] >> (7 - (i % 8))) & 1)
	}
	pos += 2

	for i := 0; i < 6; i++ {
		var n uint32 = 0
		// Get all of the bits to a uint32
		for j := 0; j < 4; j++ {
			n |= uint32(uint32(data[pos]) << (8 * (3 - j)))
			pos++
		}
		// Convert to float32
		p.GamepadState.Axes[i] = float32(n)
	}

	return nil
}

// Bytes turns the data from the packet into the byte slice it represents
func (p Packet) Bytes() []byte {
	// Create our byte array
	b := make([]byte, 31, 31)

	pos := 0
	// Make the packet id by shifting over the appropriate number of bytes and masking with 255
	for i := 0; i < 4; i++ {
		b[pos] = byte(uint32(p.PacketId) >> (8 * (3 - i)) & 255)
		pos++
	}

	// Get the joystick id
	b[pos] = p.JoystickId
	pos++

	// Get the buttons by turning each bit into the correct position in the array
	for i := 0; i < 15; i++ {
		b[pos+int(i/8)] |= byte(p.GamepadState.Buttons[i] << (7 - (i % 8)))
	}
	pos += 2

	for i := 0; i < 6; i++ {
		var n uint32 = uint32(p.GamepadState.Axes[i])
		for j := 0; j < 4; j++ {
			b[pos] = byte((n >> (8 * (3 - j))) & 255)
			pos++
		}
	}

	return b
}
