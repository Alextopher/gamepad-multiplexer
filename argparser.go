package main

import (
	"io/ioutil"
	"log"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/go-gl/glfw/v3.3/glfw"
	"gopkg.in/yaml.v2"
)

// CommandLine is used to define flags when calling the program
type CommandLine struct {
	Config  string `short:"c" help:"Configuration file location" default:"configs/gpmux.yml"`
	Listen  bool   `short:"l" help:"Specify whether to listen as a server rather than connect"`
	Domain  string `short:"d" help:"The ip or domain to use" default:"localhost"`
	Port    uint16 `short:"p" help:"The port to use" default:"14695"`
	Name    string `short:"n" help:"The name of the client" default:"client"`
	Verbose bool   `short:"v" help:"Increase verbosity level"`
}

// Parse the command line arguments
// This should only be called from the main Parse method in this package
func argParse() (cli CommandLine) {
	ctx := kong.Parse(&cli)
	switch ctx.Command() {
	// case "config":
	// 	log.Println("Foo")
	default:
		return
	}
}

type ClientsMap map[string]RulesMap
type RulesMap map[glfw.Joystick][]MultiplexRule
type ButtonMap map[glfw.GamepadButton]MapRule
type AxisMap map[glfw.GamepadAxis]MapRule

type MultiplexRule struct {
	Type   int
	Button glfw.GamepadButton
	Axis   glfw.GamepadAxis
}

type MapRule struct {
	Key0 string
	Key1 string
}

const (
	Button = iota
	Axis
)

type Config struct {
	Clients map[string]map[string][]string `yaml:"clients"`
	Mapping map[string]string              `yaml:"mapping"`
}

func stringToRule(rule string) MultiplexRule {
	switch rule {
	case "BUTTON_CROSS":
		fallthrough
	case "BUTTON_A":
		return MultiplexRule{Button, glfw.ButtonA, 0}
	case "BUTTON_CIRCLE":
		fallthrough
	case "BUTTON_B":
		return MultiplexRule{Button, glfw.ButtonB, 0}
	case "BUTTON_SQUARE":
		fallthrough
	case "BUTTON_X":
		return MultiplexRule{Button, glfw.ButtonX, 0}
	case "BUTTON_TRIANGLE":
		fallthrough
	case "BUTTON_Y":
		return MultiplexRule{Button, glfw.ButtonY, 0}
	case "BUTTON_LEFT_BUMPER":
		return MultiplexRule{Button, glfw.ButtonLeftBumper, 0}
	case "BUTTON_RIGHT_BUMPER":
		return MultiplexRule{Button, glfw.ButtonRightBumper, 0}
	case "BUTTON_BACK":
		return MultiplexRule{Button, glfw.ButtonBack, 0}
	case "BUTTON_START":
		return MultiplexRule{Button, glfw.ButtonStart, 0}
	case "BUTTON_GUIDE":
		return MultiplexRule{Button, glfw.ButtonGuide, 0}
	case "BUTTON_LEFT_THUMB":
		return MultiplexRule{Button, glfw.ButtonLeftThumb, 0}
	case "BUTTON_RIGHT_THUMB":
		return MultiplexRule{Button, glfw.ButtonRightThumb, 0}
	case "BUTTON_DPAD_UP":
		return MultiplexRule{Button, glfw.ButtonDpadUp, 0}
	case "BUTTON_DPAD_RIGHT":
		return MultiplexRule{Button, glfw.ButtonDpadRight, 0}
	case "BUTTON_DPAD_DOWN":
		return MultiplexRule{Button, glfw.ButtonDpadDown, 0}
	case "BUTTON_DPAD_LEFT":
		return MultiplexRule{Button, glfw.ButtonDpadLeft, 0}
	case "AXIS_LEFT_X":
		return MultiplexRule{Axis, 0, glfw.AxisLeftX}
	case "AXIS_LEFT_Y":
		return MultiplexRule{Axis, 0, glfw.AxisLeftY}
	case "AXIS_RIGHT_X":
		return MultiplexRule{Axis, 0, glfw.AxisRightX}
	case "AXIS_RIGHT_Y":
		return MultiplexRule{Axis, 0, glfw.AxisRightY}
	case "AXIS_LEFT_TRIGGER":
		return MultiplexRule{Axis, 0, glfw.AxisLeftTrigger}
	case "AXIS_RIGHT_TRIGGER":
		return MultiplexRule{Axis, 0, glfw.AxisRightTrigger}
	}

	log.Fatalf("CONFIG ERROR: Unrecognized BUTTON or AXIS %s!\n", rule)
	return MultiplexRule{}
}

func readConfig(filename string) (clientRules ClientsMap, buttonMap ButtonMap, axisMap AxisMap) {
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		// TODO Don't panic
		panic(err)
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalln("CONFIG ERROR: Failed to read config file due to error:", err)
	}

	// Parse client rules
	// id -> controller -> [rules]
	clientRules = make(ClientsMap)
	for id, joysticks := range config.Clients {
		if _, exists := clientRules[id]; exists {
			log.Fatalf("CONFIG ERROR: client %s already defined.", id)
		}

		clientRules[id] = make(map[glfw.Joystick][]MultiplexRule)
		for joystick, rules := range joysticks {
			joystick := glfw.Joystick(joystick[len(joystick)-1] - '0')

			if _, exists := clientRules[id][joystick]; exists {
				log.Fatalf("CONFIG ERROR: client %s joystick %d already defined.", id, joystick)
			}

			clientRules[id][joystick] = make([]MultiplexRule, len(rules))
			for i, rule := range rules {
				clientRules[id][joystick][i] = stringToRule(rule)
			}
		}
	}

	buttonMap = make(map[glfw.GamepadButton]MapRule)
	axisMap = make(map[glfw.GamepadAxis]MapRule)

	// Parse mapping "input gamestate -> output keypress"
	for input, key := range config.Mapping {
		rule := stringToRule(input)

		if rule.Type == Axis {
			if rule.Axis == glfw.AxisLeftTrigger || rule.Axis == glfw.AxisRightTrigger {
				// Triggers map to a single key
				axisMap[rule.Axis] = MapRule{key, ""}
			} else {
				// Joysticks left and right axes require 2 keys to properly handle
				// The first key is for negative axis values, second key is positive values
				keys := strings.Fields(key)
				if len(keys) != 2 {
					log.Fatalf("CONFIG ERROR: rule %s requires 2 key outputs.\nFor example:\n%s: left right", input, input)
				}

				axisMap[rule.Axis] = MapRule{keys[0], keys[1]}
			}
		} else {
			// Buttons and the
			buttonMap[rule.Button] = MapRule{key, ""}
		}
	}

	return clientRules, buttonMap, axisMap
}
