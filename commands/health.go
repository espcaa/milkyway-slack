package commands

import (
	"milkyway-slack/structs"
	"milkyway-slack/utils"
	"net/http"
)

type HealthCommand struct{}

func (c HealthCommand) Run(w http.ResponseWriter, r *http.Request) {
	// send a 200 status code with a message "OK" to acknoweldge the request

	var answer = structs.Block{
		Type: "section",
		Text: &structs.Text{
			Type: "mrkdwn",
			Text: "working ^-^",
		},
	}

	utils.SlackAnswer(w, []structs.Block{answer})

}
