package main

import (
	"github.com/go-gl/glfw/v3.3/glfw"
)

const STICK_DEADZONE float32 = 0.20
const TRIGGER_DEADZONE float32 = 0.40

func abs32(f float32) float32 {
	if f < 0 {
		return -f
	} else {
		return f
	}
}

func multiplexTrust(states map[glfw.Joystick]glfw.GamepadState, multiplexed *glfw.GamepadState) {
	// totals to calculate average
	axesUsed := []float32{0, 0, 0, 0, 0, 0}
	multiplexed.Axes = [6]float32{0, 0, 0, 0, -1, -1}
	multiplexed.Buttons = [15]glfw.Action{glfw.Release}

	for _, state := range states {
		for i := 0; i < len(multiplexed.Buttons); i++ {
			multiplexed.Buttons[i] |= state.Buttons[i]
		}

		for i := 0; i < len(multiplexed.Axes); i++ {
			if abs32(state.Axes[i]) > STICK_DEADZONE {
				multiplexed.Axes[i] += state.Axes[i]
				axesUsed[i] += 1
			}
		}
	}

	// Joysticks are centered at 0
	for _, axis := range JOYSTICK_AXES {
		if axesUsed[axis] == 0 {
			multiplexed.Axes[axis] = 0
		} else {
			// Average all the inputs from controllers
			multiplexed.Axes[axis] = multiplexed.Axes[axis] / axesUsed[axis]
		}
	}

	// Triggers are centered at -1
	for _, axis := range TRIGGER_AXES {
		if axesUsed[axis] == 0 {
			multiplexed.Axes[axis] = -1
		} else {
			// Average all the inputs from controllers
			multiplexed.Axes[axis] = multiplexed.Axes[axis] / axesUsed[axis]
		}
	}
}

func multiplex(rules RulesMap, states map[glfw.Joystick]glfw.GamepadState, multiplexed *glfw.GamepadState) {
	// totals to calculate average
	axesUsed := []float32{0, 0, 0, 0, 0, 0}
	multiplexed.Axes = [6]float32{0, 0, 0, 0, -1, -1}
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
				if rule.Axis == glfw.AxisLeftTrigger || rule.Axis == glfw.AxisRightTrigger {
					// Triggers rest at -1
					if state.Axes[rule.Axis] > -1+TRIGGER_DEADZONE {
						multiplexed.Axes[rule.Axis] += state.Axes[rule.Axis]
					} else {
						multiplexed.Axes[rule.Axis] += -1
					}
					axesUsed[rule.Axis] += 1
				} else {
					// Joysticks rest at 0
					if abs32(state.Axes[rule.Axis]) > STICK_DEADZONE {
						multiplexed.Axes[rule.Axis] += state.Axes[rule.Axis]
						axesUsed[rule.Axis] += 1
					}
				}
			}
		}
	}

	// Joysticks are centered at 0
	for _, axis := range JOYSTICK_AXES {
		if axesUsed[axis] == 0 {
			multiplexed.Axes[axis] = 0
		} else {
			// Average all the inputs from controllers
			multiplexed.Axes[axis] = multiplexed.Axes[axis] / axesUsed[axis]
		}
	}

	// Triggers are centered at -1
	for _, axis := range TRIGGER_AXES {
		if axesUsed[axis] == 0 {
			multiplexed.Axes[axis] = -1
		} else {
			// Average all the inputs from controllers
			multiplexed.Axes[axis] = multiplexed.Axes[axis] / axesUsed[axis]
		}
	}
}
