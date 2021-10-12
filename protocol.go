package main

import (
	"encoding/binary"
	"errors"
	"math"
	"regexp"
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"
)

const Interval time.Duration = 100 * time.Millisecond

var namePattern = regexp.MustCompile("[a-zA-Z0-9-]+")

const (
	REGISTER              = 1
	SET_ID                = 2
	CONFIGURATION         = 3
	PERIPHERAL_CONNECT    = 4
	PERIPHERAL_DISCONNECT = 5
	DONE                  = 6
	ERROR                 = 255
)

type ControlProtocol struct {
	Type uint8
	Len  uint32
	Data []byte
}

// Parse the data from the packet and convert to valid struct
func (p *ControlProtocol) Parse(data []byte) error {
	if len(data) < 5 {
		return errors.New("packet improperly formatted")
	}

	pos := 0

	// Get the type
	p.Type = data[pos]
	pos++

	// Get the length of the data
	p.Len = binary.BigEndian.Uint32(data[pos : pos+4])
	pos += 4

	// Make sure the packet size is right
	if len(data)-5 < 0 || uint32(len(data)-5) < p.Len {
		return errors.New("packet improperly formatted")
	}

	// Get the data
	p.Data = data[pos:(pos + int(p.Len))]

	return nil
}

// Bytes turns the data from the packet into the byte slice it represents
func (p *ControlProtocol) Bytes() []byte {
	data := make([]byte, 5, 5+p.Len)
	// Set the type
	data[0] = p.Type

	// Set the length
	binary.BigEndian.PutUint32(data[1:], p.Len)

	// Set the message
	data = append(data, p.Data...)
	return data
}

// Error returns an error packet to send
func (p *ControlProtocol) Error(msg string) []byte {
	p.Type = ERROR
	p.Len = uint32(len(msg))
	p.Data = []byte(msg)

	return p.Bytes()
}

// Register returns a REGISTER packet to send
func (p *ControlProtocol) Register(name string) []byte {
	p.Type = REGISTER
	p.Len = uint32(len(name))
	p.Data = []byte(name)

	return p.Bytes()
}

// SetId returns a SET_ID packet to send
func (p *ControlProtocol) SetId(id uint8) []byte {
	p.Type = SET_ID
	p.Len = 1
	p.Data = []byte{id}

	return p.Bytes()
}

// Configure returns a CONFIGURATION packet to send
func (p *ControlProtocol) Configure(data []byte) []byte {
	p.Type = CONFIGURATION
	p.Len = uint32(len(data))
	p.Data = data

	return p.Bytes()
}

type GamestateProtocol struct {
	PacketId     uint32
	JoystickId   uint8
	GamepadState glfw.GamepadState
}

// Parse the data from the packet and convert to valid struct
func (p *GamestateProtocol) Parse(data []byte) error {
	// Bad packet length
	if len(data) != 31 {
		return errors.New("invalid packet length")
	}

	pos := 0
	// Get the packet id by shifting each byte into the appropriate place of the uint32
	p.PacketId = binary.BigEndian.Uint32(data)
	pos += 4

	// Get the joystick id
	p.JoystickId = data[pos]
	pos++

	// Get the buttons by turning each bit into the correct position in the array
	for i := 0; i < 15; i++ {
		p.GamepadState.Buttons[i] = glfw.Action((data[pos+int(i/8)] >> (7 - (i % 8))) & 1)
	}
	pos += 2

	for i := 0; i < 6; i++ {
		// Get bytes as uint32
		var n uint32 = binary.BigEndian.Uint32(data[pos:])
		// Convert to float32
		p.GamepadState.Axes[i] = math.Float32frombits(n)
		// Bump up the position
		pos += 4
	}

	return nil
}

// Bytes turns the data from the packet into the byte slice it represents
func (p GamestateProtocol) Bytes() []byte {
	// Create our byte slice
	b := make([]byte, 31, 31)
	pos := 0

	// Get our packet id
	binary.BigEndian.PutUint32(b, p.PacketId)
	pos += 4

	// Get the joystick id
	b[pos] = p.JoystickId
	pos++

	// Get the buttons by turning each bit into the correct position in the array
	for i := 0; i < 15; i++ {
		b[pos+int(i/8)] |= byte(p.GamepadState.Buttons[i] << (7 - (i % 8)))
	}
	pos += 2

	for i := 0; i < 6; i++ {
		n := math.Float32bits(p.GamepadState.Axes[i])
		binary.BigEndian.PutUint32(b[pos:], n)
		pos += 4
	}

	return b
}

func (rules RulesMap) Bytes() []byte {
	length := 0
	for _, array := range rules {
		length += len(array) + 2
	}

	b := make([]byte, 0, length)
	for joystick, array := range rules {
		// 1 byte joystick id [0, 15]
		b = append(b, byte(joystick))
		for _, rule := range array {
			// 1 byte per rule.
			// 1st bit represents if rule is an Button (0) Axis (1)
			if rule.Type == Button {
				b = append(b, byte(rule.Button))
			} else {
				b = append(b, byte(rule.Axis|1<<7))
			}
		}
		b = append(b, 0xFF)
	}

	return b
}

func ParseRulesMap(bytes []byte) (RulesMap, error) {
	rules := make(RulesMap)

	for i := 0; i < len(bytes); i++ {
		joystick := glfw.Joystick(i)
		rules[joystick] = make([]MultiplexRule, 0)

		i++
		for bytes[i] != 255 {
			var rule MultiplexRule
			if bytes[i]&128 == 0 {
				rule = MultiplexRule{Button, glfw.GamepadButton(bytes[i] & 127), 0}
			} else {
				rule = MultiplexRule{Axis, 0, glfw.GamepadAxis(bytes[i] & 127)}
			}
			rules[joystick] = append(rules[joystick], rule)
			i++
		}
	}

	return rules, nil
}
