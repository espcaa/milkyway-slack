package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type RoomCommand struct{}

func (c RoomCommand) Run(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("failed to parse form: %w", err)
	}

	responseURL := r.PostFormValue("response_url")
	if responseURL == "" {
		return fmt.Errorf("no response_url provided")
	}

	// Send immediate response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"text": "Processing room command...",
	}); err != nil {
		return fmt.Errorf("failed to send immediate response: %w", err)
	}

	go func(url string) {
		blocks := []Block{
			{
				Type: "section",
				Text: &Text{
					Type: "mrkdwn",
					Text: "🌠 room",
				},
			},
		}

		payload := map[string]interface{}{
			"response_type": "in_channel",
			"blocks":        blocks,
			"text":          "room",
		}

		if err := sendSlackResponse(url, payload); err != nil {
			log.Printf("Error sending final response: %v", err)
			sendErrorResponse(url, "Failed to send room message :(")
		}
	}(responseURL)

	return nil
}
