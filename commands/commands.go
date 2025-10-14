package commands

import "milkyway-slack/structs"

func InitCommands(bot structs.BotInterface) map[string]structs.Command {
	return map[string]structs.Command{
		"health": HealthCommand{},
		"room":   RoomCommand{Bot: bot},
	}
}

var CommandRegistry map[string]structs.Command
