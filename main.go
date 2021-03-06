package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// A CommandHandler is a command handler. cpt obvious reporting in
type CommandHandler func(Entity, []string) Response

// A Command is something that is called from the request and has a trigger. It then executes the handler
type Command struct {
	trigger    string
	entityless bool
	handler    CommandHandler
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

func receiveData(data []byte, sender *net.UDPAddr, err error) (Response, bool) {
	if err != nil {
		log.Printf("Error receiving data: '%s'\n", err)
		return Response{Status: "error", Data: "Could not receive your request"}, true
	}

	var request Request
	err = json.Unmarshal(data, &request)
	if err != nil {
		return Response{Status: "error", Data: "invalid request format"}, true
	}

	// Check if entity exists
	var entityExists = false
	var requestEntity Entity
	for _, entity := range entities {
		if entity.Name == request.EntityName {
			entityExists = true
			requestEntity = entity
			break
		}
	}

	// Check if the command is an actual command
	retn := Response{Status: "error", Data: "invalid command"}
	for _, command := range commands {
		if command.trigger == strings.ToLower(request.Command) {
			if entityExists || (!entityExists && command.entityless) {
				retn = command.handler(requestEntity, request.Args)
				break
			}
			retn = Response{Status: "error", Data: "invalid entity"}
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
	var configFile = "./gohome.json"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}
	config = LoadConfig(configFile)
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
	commands = append(commands, Command{trigger: "state", handler: func(entity Entity, args []string) Response {
		if len(args) == 0 {
			var state int
			bstate, err := pins[entity.Pin].State()
			if err != nil {
				return Response{Status: "error", Data: fmt.Sprintf("internal error: %v", err)}
			}

			if bstate {
				state = 1
			}
			return Response{Status: "ok", Data: strconv.Itoa(state)}
		}
		if !entity.IsOutput() {
			return Response{Status: "error", Data: "entity is input"}
		}
		state, err := strconv.Atoi(args[0])
		if err != nil {
			return Response{Status: "error", Data: "state not an integer"}
		}
		if state != 0 {
			pins[entity.Pin].SetState(true)
		} else {
			pins[entity.Pin].SetState(false)
		}
		return Response{Status: "ok", Data: args[0]}

	}})
	// Direction Command
	commands = append(commands, Command{trigger: "direction", handler: func(entity Entity, args []string) Response {
		var direction = "input"
		if pins[entity.Pin].Direction() {
			direction = "output"
		}
		return Response{Status: "ok", Data: direction}
	}})
	// Toggle Command
	commands = append(commands, Command{trigger: "toggle", handler: func(entity Entity, args []string) Response {
		if !entity.IsOutput() {
			return Response{Status: "error", Data: "entity is input"}
		}
		bstate, err := pins[entity.Pin].State()
		if err != nil {
			return Response{Status: "error", Data: fmt.Sprintf("internal error: %v", err)}
		}
		pins[entity.Pin].SetState(!bstate)
		var state int
		if bstate {
			state = 1
		}
		return Response{Status: "ok", Data: strconv.Itoa(state)}
	}})
	// List Command
	commands = append(commands, Command{entityless: true, trigger: "list", handler: func(entity Entity, args []string) Response {
		data, err := json.Marshal(entities)
		if err != nil {
			return Response{Status: "error", Data: fmt.Sprintf("internal error: %v", err)}
		}
		return Response{Status: "ok", Data: string(data)}
	}})

	go server.Process()

	for {
		time.Sleep(1 * time.Millisecond)
	}

}
