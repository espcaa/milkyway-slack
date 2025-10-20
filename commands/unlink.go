package commands

import (
	"encoding/json"
	"fmt"
	"milkyway-slack/structs"
	"net/http"
)

type LinkCommand struct {
	Bot structs.BotInterface
}

func (c LinkCommand) Run(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("failed to parse form: %w", err)
	}

	text := r.PostFormValue("text")

	userID := r.PostFormValue("user_id")

	// just delete any existing overrides

	_, err := c.Bot.GetDB().Exec(
		"DELETE FROM user_overrides WHERE slack_id = ?",
		userID,
	)
	if err != nil {
		response := map[string]interface{}{
			"response_type": "ephemeral",
			"text":          "whoopsie the db is cooked",
		}
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(response)
	}

	if text == "" {
		response := map[string]interface{}{
			"response_type": "ephemeral",
			"text":          "please provide an email address to link",
		}
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(response)
	}

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"response_type": "ephemeral",
		"text":          fmt.Sprintf("Successfully unlinked your Slack account to your milkyway account"),
	}
	return json.NewEncoder(w).Encode(response)
}
