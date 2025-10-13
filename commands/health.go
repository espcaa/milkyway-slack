package commands

import (
	"log"
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

	err := utils.SlackAnswer(w, []structs.Block{answer})
	if err != nil {
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
		log.Println("Error sending response:", err)
		return
	}

}
