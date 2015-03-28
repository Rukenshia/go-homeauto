package main

import (
	"encoding/json"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

// A CommandHandler is a command handler. cpt obvious reporting in
type CommandHandler func(Entity, []string) Answer

// A Command is something that is called from the request and has a trigger. It then executes the handler
type Command struct {
	trigger string
	handler CommandHandler
}

// Request represents a Request to the server.
type Request struct {
	Command    string
	EntityName string
	Args       []string
}

var config Config
var pins []GPIOPin
var entities []Entity
var commands []Command

func receiveData(data []byte, sender *net.UDPAddr, err error) (Answer, bool) {
	if err != nil {
		log.Printf("Error receiving data: '%s'\n", err)
		return Answer{Status: "error", Data: "Could not receive your request"}, true
	}

	var request Request
	err = json.Unmarshal(data, &request)
	if err != nil {
		return Answer{Status: "error", Data: "invalid request format"}, true
	}

	// Check if entity exists
	var exists = false
	var requestEntity Entity
	for _, entity := range entities {
		if entity.Name == request.EntityName {
			exists = true
			requestEntity = entity
			break
		}
	}
	if !exists {
		return Answer{Status: "error", Data: "invalid entity"}, true
	}

	// Check if the command is an actual command
	retn := Answer{Status: "error", Data: "invalid command"}
	for _, command := range commands {
		log.Printf("%s,%s\n", request.Command, command.trigger)
		if command.trigger == strings.ToLower(request.Command) {
			retn = command.handler(requestEntity, request.Args)
			break
		}
	}

	return retn, true
}

func createPins(revision uint8) []GPIOPin {
	var retn []GPIOPin
	var pins = 16

	switch revision {
	case 2:
		{
			// RPi 1 Revision 2
			pins = 20
		}
	case 3:
		{
			// RPi 1 B+ / RPi 2
			pins = 31
		}
	}

	for i := 0; i < pins+1; i++ {
		if i >= 17 && i <= 20 {
			// Pins do actually not exist, makes the pi crash
			continue
		}
		retn = append(retn, CreatePin(i, INPUT))
	}

	return retn
}

func main() {
	config = LoadConfig("./config.json")
	entities = config.Entities

	pins = createPins(config.Board)

	log.Println("Loaded Config, configuring all pins")
	for _, pin := range pins {
		pin.SetState(LOW)
	}

	for _, entity := range entities {
		if int(entity.Pin) >= len(pins) {
			log.Fatalf("Misconfiguration: we have %d pins but the entity %s points to pin %d", len(pins), entity.Name, entity.Pin)
		} else {
			if entity.IsOutput() {
				pins[entity.Pin].SetDirection(OUTPUT)
			}
		}
	}

	log.Println("All Pins configured. Starting Server.")

	server := CreateServer("0.0.0.0", 4200)
	server.RegisterCallback(receiveData)
	err := server.Start()

	if err != nil {
		log.Fatalf("Could not create Server: '%s'\n", err)
	}
	log.Printf("Server on %s:%d created.\n", server.info.IP, server.info.Port)

	// Register Commands
	commands = append(commands, Command{trigger: "state", handler: func(entity Entity, args []string) Answer {
		if len(args) == 0 {
			var state int
			bstate, err := pins[entity.Pin].State()
			if err != nil {
				return Answer{Status: "error", Data: "internal error"}
			}

			if bstate {
				state = 1
			}
			return Answer{Status: "ok", Data: strconv.Itoa(state)}
		}
		if !entity.IsOutput() {
			return Answer{Status: "error", Data: "entity is input"}
		}
		state, err := strconv.Atoi(args[0])
		if err != nil {
			return Answer{Status: "error", Data: "state not an integer"}
		}
		if state != 0 {
			pins[entity.Pin].SetState(true)
		} else {
			pins[entity.Pin].SetState(false)
		}
		return Answer{Status: "ok"}

	}})
	// Direction Command
	commands = append(commands, Command{trigger: "direction", handler: func(entity Entity, args []string) Answer {
		var direction int
		if pins[entity.Pin].Direction() {
			direction = 1
		}
		return Answer{Status: "ok", Data: strconv.Itoa(direction)}
	}})
	// Toggle Command
	commands = append(commands, Command{trigger: "toggle", handler: func(entity Entity, args []string) Answer {
		if !entity.IsOutput() {
			return Answer{Status: "error", Data: "entity is input"}
		}
		bstate, err := pins[entity.Pin].State()
		if err != nil {
			return Answer{Status: "error", Data: "internal error"}
		}
		pins[entity.Pin].SetState(!bstate)
		var state int
		if bstate {
			state = 1
		}
		return Answer{Status: "ok", Data: strconv.Itoa(state)}
	}})

	go server.Process()

	for {
		time.Sleep(1 * time.Millisecond)
	}

}
