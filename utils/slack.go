package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"milkyway-slack/structs"
	"mime/multipart"
	"net/http"
	"os"
)

func SlackAnswer(w http.ResponseWriter, messages []structs.Block) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]any{
		"response_type": "in_channel",
		"blocks":        messages,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return errors.New("failed to encode response: " + err.Error())
	}

	return nil
}

func UploadImageFromBuffer(img image.Image, fileName string, channels []string) (string, error) {
	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		return "", fmt.Errorf("SLACK_TOKEN not set")
	}

	var imgBuf bytes.Buffer
	if err := png.Encode(&imgBuf, img); err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, &imgBuf); err != nil {
		return "", fmt.Errorf("failed to copy image buffer: %w", err)
	}

	if len(channels) > 0 {
		if err := writer.WriteField("channels", stringJoin(channels, ",")); err != nil {
			return "", fmt.Errorf("failed to write channels field: %w", err)
		}
	}

	writer.Close()

	req, err := http.NewRequest("POST", "https://slack.com/api/files.upload", &body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+slackToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		OK   bool `json:"ok"`
		File struct {
			ID string `json:"id"`
		} `json:"file"`
		Error string `json:"error,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode Slack response: %w", err)
	}

	if !result.OK {
		return "", fmt.Errorf("Slack API error: %s", result.Error)
	}

	return result.File.ID, nil
}

func stringJoin(items []string, sep string) string {
	if len(items) == 0 {
		return ""
	}
	res := items[0]
	for _, s := range items[1:] {
		res += sep + s
	}
	return res
}

func MakeSlackFilePublic(fileID string) (string, error) {
	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		return "", fmt.Errorf("SLACK_TOKEN not set")
	}

	payload := map[string]string{
		"file": fileID,
	}

	bodyBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "https://slack.com/api/files.sharedPublicURL", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+slackToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		OK   bool `json:"ok"`
		File struct {
			PermalinkPublic string `json:"permalink_public"`
		} `json:"file"`
		Error string `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.OK {
		return "", fmt.Errorf("Slack API error: %s", result.Error)
	}

	return result.File.PermalinkPublic, nil
}
