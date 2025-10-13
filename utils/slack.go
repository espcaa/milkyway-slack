package utils

import (
	"encoding/json"
	"milkyway-slack/structs"
	"net/http"
)

func SlackAnswer(w http.ResponseWriter, messages []structs.Block) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]any{
		"response_type": "in_channel",
		"blocks":        messages,
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
