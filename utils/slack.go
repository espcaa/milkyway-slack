package utils

import (
	"encoding/json"
	"errors"
	"milkyway-slack/structs"
	"net/http"
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
