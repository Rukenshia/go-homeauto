package main

import (
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

func gpioExec(args []string) (string, error) {
	cmd := exec.Command("gpio", args...)
	output, err := cmd.Output()

	if err != nil {
		log.Printf("Error: '%s (args: %v)'\n", err, args)
	}
	if output[len(output)-1] == '\n' {
		output = output[:len(output)-1]
	}
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
func (pin *GPIOPin) State() (bool, error) {
	// TODO: erm, this
	output, err := gpioExec([]string{"read", strconv.Itoa(pin.Number)})
	if handleError(err) {
		return false, err
	}

	state, err := strconv.Atoi(output)
	return (state == 1), err
}

// SetState sets the state of the GPIOPin.
func (pin *GPIOPin) SetState(state bool) error {
	var sstate = "0"
	if state {
		sstate = "1"
	}
	_, err := gpioExec([]string{"write", strconv.Itoa(pin.Number), sstate})
	return err
}

// Direction returns the direction of the GPIOPin.
func (pin *GPIOPin) Direction() bool {
	return pin.direction
}

// SetDirection sets the Direction of a GPIOPin.
func (pin *GPIOPin) SetDirection(direction bool) error {
	pin.direction = direction
	var dir = "in"
	if direction {
		dir = "out"
	}
	_, err := gpioExec([]string{"mode", strconv.Itoa(pin.Number), dir})
	return err
}
