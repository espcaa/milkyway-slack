package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"milkyway-slack/utils"
	"net/http"
)

type HealthCommand struct{}

type Block struct {
	Type     string `json:"type"`
	Text     *Text  `json:"text,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
	AltText  string `json:"alt_text,omitempty"`
}

type Text struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (c HealthCommand) Run(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("failed to parse form: %w", err)
	}

	responseURL := r.PostFormValue("response_url")
	if responseURL == "" {
		return fmt.Errorf("no response_url provided")
	}

	// answer quickly
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"text": "checking health...",
	}); err != nil {
		return fmt.Errorf("failed to send immediate response: %w", err)
	}

	url, err := utils.UploadFile("health.png")
	if err != nil {
		log.Printf("Error uploading file: %v", err)
		return sendErrorResponse(responseURL, "Failed to upload health status image :(")
	}

	blocks := []Block{
		{
			Type: "section",
			Text: &Text{
				Type: "mrkdwn",
				Text: "hiii, everything should work ^-^",
			},
		},
		{
			Type:     "image",
			ImageURL: url,
			AltText:  "health status",
		},
	}

	payload := map[string]interface{}{
		"response_type": "in_channel",
		"blocks":        blocks,
		"text":         "health check",
	}

	if err := sendSlackResponse(responseURL, payload); err != nil {
		log.Printf("Error sending final response: %v", err)
		return sendErrorResponse(responseURL, "Failed to send health status :(")
	}

	return nil
}

func sendSlackResponse(url string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func sendErrorResponse(url, message string) error {
	payload := map[string]string{
		"response_type": "ephemeral",
		"text":         message,
	}

	return sendSlackResponse(url, payload)
}
