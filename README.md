# gpmux

Work in progress gamepad multiplexer. Allows multiple players to control the same character in games.

## todo list:

- [x] connect & disconnect controllers
- [x] detect gamepad
- [x] read buttons and axis state from gamepads
- [x] parse rules from yaml file
- [x] multiplex gamepad data to single virtual gamepad
- [ ] map multiplexed gamepad to keyboard and mouse events
- [ ] output keyboard events to the OS
- [ ] create a server to receive button presses
- [ ] create a client to send button presses

## Valid rules:
```
BUTTON_A
BUTTON_B
BUTTON_X
BUTTON_Y
BUTTON_LEFT_BUMPER
BUTTON_RIGHT_BUMPER
BUTTON_BACK
BUTTON_START
BUTTON_GUIDE
BUTTON_LEFT_THUMB
BUTTON_RIGHT_THUMB
BUTTON_DPAD_UP
BUTTON_DPAD_RIGHT
BUTTON_DPAD_DOWN
BUTTON_DPAD_LEFT
BUTTON_CROSS - BUTTON_A
BUTTON_CIRCLE - BUTTON_B
BUTTON_SQUARE - BUTTON_X
BUTTON_TRIANGLE - BUTTON_Y

AXIS_LEFT_X
AXIS_LEFT_Y
AXIS_RIGHT_X
AXIS_RIGHT_Y
AXIS_LEFT_TRIGGER
AXIS_RIGHT_TRIGGER
```