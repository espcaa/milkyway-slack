package commands

import (
	"encoding/json"
	"fmt"
	"milkyway-slack/structs"
	"net/http"
	"strings"
)

type LinkCommand struct {
	Bot structs.BotInterface
}

func (c LinkCommand) Run(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("failed to parse form: %w", err)
	}

	userID := r.PostFormValue("user_id")
	text := strings.TrimSpace(r.PostFormValue("text"))

	if text == "" {
		return fmt.Errorf("please provide an email address to link")
	}

	// Validate email format
	if !strings.Contains(text, "@") {
		response := map[string]interface{}{
			"response_type": "ephemeral",
			"text":          "Invalid email format. Please provide a valid email address. /link [email-adress-here]",
		}
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(response)
	}

	// Store in override db
	_, err := c.Bot.GetDB().Exec(
		"INSERT OR REPLACE INTO user_overrides (slack_id, email) VALUES (?, ?)",
		userID, text,
	)
	if err != nil {
		response := map[string]interface{}{
			"response_type": "ephemeral",
			"text":          "whoopsie the db is cooked",
		}
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(response)
	}

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"response_type": "ephemeral",
		"text":          fmt.Sprintf("Successfully linked your Slack account to email: %s", text),
	}
	return json.NewEncoder(w).Encode(response)
}
