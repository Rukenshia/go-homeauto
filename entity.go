package main

import (
	"strings"
)

// An Entity is 'something' connected to a specific GPIOPin with a Name.
type Entity struct {
	Name         string
	FriendlyName string
	Pin          uint
	Type         string
}

// IsInput returns whether the Entity is an Input.
func (entity *Entity) IsInput() bool {
	return (strings.ToLower(entity.Type) == "input")
}

// IsOutput returns whether the Entity is an Output.
func (entity *Entity) IsOutput() bool {
	return (strings.ToLower(entity.Type) == "output")
}
