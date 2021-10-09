package main

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

const DEADZONE float32 = 0.20

var rules map[uint8][]multiplexRule

type multiplexRule struct {
	Type   int
	Button glfw.GamepadButton
	Axis   glfw.GamepadAxis
}

const (
	Button = iota
	Axis
)

type Config struct {
	Controllers map[string][]string `yaml:"controllers"`
}

func abs32(f float32) float32 {
	if f < 0 {
		return -f
	} else {
		return f
	}
}

func stringToRule(rule string) multiplexRule {
	switch rule {
	case "BUTTON_CROSS":
		fallthrough
	case "BUTTON_A":
		return multiplexRule{Button, glfw.ButtonA, 0}
	case "BUTTON_CIRCLE":
		fallthrough
	case "BUTTON_B":
		return multiplexRule{Button, glfw.ButtonB, 0}
	case "BUTTON_SQUARE":
		fallthrough
	case "BUTTON_X":
		return multiplexRule{Button, glfw.ButtonX, 0}
	case "BUTTON_TRIANGLE":
		fallthrough
	case "BUTTON_Y":
		return multiplexRule{Button, glfw.ButtonY, 0}
	case "BUTTON_LEFT_BUMPER":
		return multiplexRule{Button, glfw.ButtonLeftBumper, 0}
	case "BUTTON_RIGHT_BUMPER":
		return multiplexRule{Button, glfw.ButtonRightBumper, 0}
	case "BUTTON_BACK":
		return multiplexRule{Button, glfw.ButtonBack, 0}
	case "BUTTON_START":
		return multiplexRule{Button, glfw.ButtonStart, 0}
	case "BUTTON_GUIDE":
		return multiplexRule{Button, glfw.ButtonGuide, 0}
	case "BUTTON_LEFT_THUMB":
		return multiplexRule{Button, glfw.ButtonLeftThumb, 0}
	case "BUTTON_RIGHT_THUMB":
		return multiplexRule{Button, glfw.ButtonRightThumb, 0}
	case "BUTTON_DPAD_UP":
		return multiplexRule{Button, glfw.ButtonDpadUp, 0}
	case "BUTTON_DPAD_RIGHT":
		return multiplexRule{Button, glfw.ButtonDpadRight, 0}
	case "BUTTON_DPAD_DOWN":
		return multiplexRule{Button, glfw.ButtonDpadDown, 0}
	case "BUTTON_DPAD_LEFT":
		return multiplexRule{Button, glfw.ButtonDpadLeft, 0}
	case "AXIS_LEFT_X":
		return multiplexRule{Axis, 0, glfw.AxisLeftX}
	case "AXIS_LEFT_Y":
		return multiplexRule{Axis, 0, glfw.AxisLeftY}
	case "AXIS_RIGHT_X":
		return multiplexRule{Axis, 0, glfw.AxisRightX}
	case "AXIS_RIGHT_Y":
		return multiplexRule{Axis, 0, glfw.AxisRightY}
	case "AXIS_LEFT_TRIGGER":
		return multiplexRule{Axis, 0, glfw.AxisLeftTrigger}
	case "AXIS_RIGHT_TRIGGER":
		return multiplexRule{Axis, 0, glfw.AxisRightTrigger}
	}

	log.Fatalf("Unrecognized rule %s!\n", rule)
	return multiplexRule{}
}

func readConfig(filename string) {
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		// TODO Don't panic
		panic(err)
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalln("Failed to read config file due to error:", err)
	}

	rules = make(map[uint8][]multiplexRule)
	for joystick, newRules := range config.Controllers {
		// `joystick0` <- get last character as int
		id := glfw.Joystick(joystick[len(joystick)-1] - '0')

		rules[uint8(id)] = make([]multiplexRule, len(newRules))
		for i, rule := range newRules {
			rules[uint8(id)][i] = stringToRule(rule)
		}
	}
}

func multiplex(states map[uint8]glfw.GamepadState) (multiplexed glfw.GamepadState) {
	// totals to calculate average
	axesUsed := []float32{0, 0, 0, 0, 0, 0}
	multiplexed.Axes = [6]float32{0, 0, 0, 0, 0, 0}
	multiplexed.Buttons = [15]glfw.Action{glfw.Release}

	for id, state := range states {
		if rules[id] == nil {
			continue
		}

		// Apply rules
		for _, rule := range rules[id] {
			switch rule.Type {
			case Button:
				// If anyone is pressing the button, then it is pressed
				multiplexed.Buttons[rule.Button] |= state.Buttons[rule.Button]
			case Axis:
				// Axes get put through a deadzone filter then averaged
				// This way if player 1 and player are moving opposite they will cancel
				// However player 1 not moving and player 2 moving won't result in half speed
				if abs32(state.Axes[rule.Axis]) > DEADZONE {
					multiplexed.Axes[rule.Axis] += state.Axes[rule.Axis]
					axesUsed[rule.Axis] += 1
				}
			}
		}
	}

	// Average the axes
	for i := 0; i < 6; i++ {
		if axesUsed[i] == 0 {
			multiplexed.Axes[i] = 0
		} else {
			multiplexed.Axes[i] = multiplexed.Axes[i] / axesUsed[i]
		}
	}
	return multiplexed
}
