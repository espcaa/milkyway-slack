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

	// load the file "health.png" from the current directory + send it via catbox file upload to slack
	url, err := utils.UploadFile("health.png")

	// create a slack block with the image

	if err != nil {
		log.Println("Error uploading file:", err)
		return
	}

	var imageBlock = structs.Block{
		Type:     "image",
		ImageURL: url,
		AltText:  "health image",
	}

	err = utils.SlackAnswer(w, []structs.Block{answer, imageBlock})
	if err != nil {
		http.Error(w, "Failed to send image response", http.StatusInternalServerError)
		log.Println("Error sending image response:", err)
		return
	}

}
