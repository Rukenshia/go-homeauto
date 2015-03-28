package main

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
)

// GPIOPin is a GPIO-Pin on the Raspberry Pi.
type GPIOPin struct {
	Number int

	direction bool
}

const (
	// INPUT marks a Pin as input.
	INPUT = false
	// OUTPUT marks a Pin as output.
	OUTPUT = true
)

const (
	// HIGH is the "on" state
	HIGH = true
	// LOW is the "off" state
	LOW = false
)

func gpioExec(command string) (string, error) {
	log.Printf("executing %s\n", command)
	cmd := exec.Command("gpio", command)
	output, err := cmd.CombinedOutput()

	return string(output), err
}

func handleError(err error) bool {
	if err != nil {
		log.Printf("Error: '%s'\n", err)
		return true
	}
	return false
}

// CreatePin creates a new GPIOPin.
func CreatePin(num int, direction bool) GPIOPin {
	pin := GPIOPin{}
	pin.Number = num
	pin.SetDirection(direction)

	return pin
}

// State returns the current state of the GPIOPin.
func (pin *GPIOPin) State() bool {
	// TODO: erm, this
	output, err := gpioExec(fmt.Sprintf("read %d", pin.Number))
	if handleError(err) {
		return false
	}

	state, err := strconv.Atoi(output)
	if handleError(err) {
		return false
	}
	return (state == 1)
}

// SetState sets the state of the GPIOPin.
func (pin *GPIOPin) SetState(state bool) {
	var istate int
	if state {
		istate = 1
	}
	_, err := gpioExec(fmt.Sprintf("write %d %d", pin.Number, istate))
	handleError(err)
}

// Direction returns the direction of the GPIOPin.
func (pin *GPIOPin) Direction() bool {
	return pin.direction
}

// SetDirection sets the Direction of a GPIOPin.
func (pin *GPIOPin) SetDirection(direction bool) {
	pin.direction = direction
	var dir = "in"
	if direction {
		dir = "out"
	}
	_, err := gpioExec(fmt.Sprintf("mode %d %s", pin.Number, dir))
	handleError(err)
}
