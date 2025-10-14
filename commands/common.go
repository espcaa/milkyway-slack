package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

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
		"text":          message,
	}

	return sendSlackResponse(url, payload)
}
