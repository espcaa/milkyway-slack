package commands

import "milkyway-slack/structs"

var CommandRegistry = map[string]structs.Command{
	"health": HealthCommand{},
	"room":   RoomCommand{},
}
