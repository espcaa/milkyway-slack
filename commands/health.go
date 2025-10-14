package commands

import (
	"bytes"
	"encoding/json"
	"log"
	"milkyway-slack/structs"
	"milkyway-slack/utils"
	"net/http"
)

type HealthCommand struct{}

func (c HealthCommand) Run(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	responseURL := r.PostFormValue("response_url")
	// send a 200 status code with a message "OK" to acknoweldge the request

	// answer quicjly

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"response_type": "ephemeral",
		"text":          "Processing your request…",
	})

	url, err := utils.UploadFile("health.png")
	if err != nil {
		log.Println("Error uploading file:", err)
		http.Post(responseURL, "application/json", bytes.NewBuffer([]byte(`{"text":"Failed to upload file :("}`)))
		return
	}

	blocks := []structs.Block{
		{
			Type: "section",
			Text: &structs.Text{
				Type: "mrkdwn",
				Text: "it's working ^-^",
			},
		},
		{
			Type:     "image",
			ImageURL: url,
			AltText:  "health image",
		},
	}

	payload := map[string]any{
		"response_type": "in_channel",
		"blocks":        blocks,
	}

	data, _ := json.Marshal(payload)
	_, err = http.Post(responseURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Println("Error sending delayed response:", err)
	}
}
