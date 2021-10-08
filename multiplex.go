package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/go-gl/glfw/v3.3/glfw"
	"gopkg.in/yaml.v2"
)

const DEADZONE float32 = 0.20

var rules map[int][]multiplexRule

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
	switch {
	case rule == "BUTTON_A" || rule == "BUTTON_CROSS":
		return multiplexRule{Button, glfw.ButtonA, 0}
	case rule == "BUTTON_B" || rule == "BUTTON_CIRCLE":
		return multiplexRule{Button, glfw.ButtonB, 0}
	case rule == "BUTTON_X" || rule == "BUTTON_SQUARE":
		return multiplexRule{Button, glfw.ButtonX, 0}
	case rule == "BUTTON_Y" || rule == "BUTTON_TRIANGLE":
		return multiplexRule{Button, glfw.ButtonY, 0}
	case rule == "BUTTON_LEFT_BUMPER":
		return multiplexRule{Button, glfw.ButtonLeftBumper, 0}
	case rule == "BUTTON_RIGHT_BUMPER":
		return multiplexRule{Button, glfw.ButtonRightBumper, 0}
	case rule == "BUTTON_BACK":
		return multiplexRule{Button, glfw.ButtonBack, 0}
	case rule == "BUTTON_START":
		return multiplexRule{Button, glfw.ButtonStart, 0}
	case rule == "BUTTON_GUIDE":
		return multiplexRule{Button, glfw.ButtonGuide, 0}
	case rule == "BUTTON_LEFT_THUMB":
		return multiplexRule{Button, glfw.ButtonLeftThumb, 0}
	case rule == "BUTTON_RIGHT_THUMB":
		return multiplexRule{Button, glfw.ButtonRightThumb, 0}
	case rule == "BUTTON_DPAD_UP":
		return multiplexRule{Button, glfw.ButtonDpadUp, 0}
	case rule == "BUTTON_DPAD_RIGHT":
		return multiplexRule{Button, glfw.ButtonDpadRight, 0}
	case rule == "BUTTON_DPAD_DOWN":
		return multiplexRule{Button, glfw.ButtonDpadDown, 0}
	case rule == "BUTTON_DPAD_LEFT":
		return multiplexRule{Button, glfw.ButtonDpadLeft, 0}
	case rule == "AXIS_LEFT_X":
		return multiplexRule{Axis, 0, glfw.AxisLeftX}
	case rule == "AXIS_LEFT_Y":
		return multiplexRule{Axis, 0, glfw.AxisLeftY}
	case rule == "AXIS_RIGHT_X":
		return multiplexRule{Axis, 0, glfw.AxisRightX}
	case rule == "AXIS_RIGHT_Y":
		return multiplexRule{Axis, 0, glfw.AxisRightY}
	case rule == "AXIS_LEFT_TRIGGER":
		return multiplexRule{Axis, 0, glfw.AxisLeftTrigger}
	case rule == "AXIS_RIGHT_TRIGGER":
		return multiplexRule{Axis, 0, glfw.AxisRightTrigger}
	}

	log.Fatalf("Unrecognized rule %s!\n", rule)
	os.Exit(1)
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
		// TODO Don't panic
		panic(err)
	}

	rules = make(map[int][]multiplexRule)
	for joystick, newRules := range config.Controllers {
		// `joystick0` <- get last character as int
		id := glfw.Joystick(joystick[len(joystick)-1] - '0')

		rules[int(id)] = make([]multiplexRule, len(newRules))
		for i, rule := range newRules {
			rules[int(id)][i] = stringToRule(rule)
		}
	}
}

func multiplex(states map[int]glfw.GamepadState, multiplexed *glfw.GamepadState) {
	// totals to calculate average
	axes_n := []float32{0, 0, 0, 0, 0, 0}
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
				if multiplexed.Buttons[rule.Button] == glfw.Press || state.Buttons[rule.Button] == glfw.Press {
					multiplexed.Buttons[rule.Button] = glfw.Press
				} else {
					multiplexed.Buttons[rule.Button] = glfw.Release
				}
			case Axis:
				// Axes get put through a deadzone filter then averaged
				// This way if player 1 and player are moving opposite they will cancel
				// However player 1 not moving and player 2 moving won't result in half speed
				if abs32(state.Axes[rule.Axis]) > DEADZONE {
					multiplexed.Axes[rule.Axis] += state.Axes[rule.Axis]
					axes_n[rule.Axis] += 1
				}
			}
		}
	}

	// Average the axes
	for i := 0; i < 6; i++ {
		if axes_n[i] == 0 {
			multiplexed.Axes[i] = 0
		} else {
			multiplexed.Axes[i] = multiplexed.Axes[i] / axes_n[i]
		}
	}
}
