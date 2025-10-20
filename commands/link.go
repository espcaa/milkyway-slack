package commands

import (
	"encoding/json"
	"fmt"
	"milkyway-slack/structs"
	"milkyway-slack/utils"
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
	email := strings.TrimSpace(r.PostFormValue("text"))

	if email == "" {
		response := map[string]interface{}{
			"response_type": "ephemeral",
			"text":          "Please provide an email address to link. /link [email-address-here]",
		}
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(response)
	}

	// Validate email format
	if !strings.Contains(email, "@") {
		response := map[string]interface{}{
			"response_type": "ephemeral",
			"text":          "Invalid email format. Please provide a valid email address. /link [email-address-here]",
		}
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(response)
	}

	// Delete any existing overrides for this user/email
	_, err := c.Bot.GetDB().Exec(
		"DELETE FROM user_overrides WHERE slack_id = ? AND email = ?",
		userID, email,
	)
	if err != nil {
		response := map[string]interface{}{
			"response_type": "ephemeral",
			"text":          "Database error occurred while removing old overrides",
		}
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(response)
	}

	// generate OTP
	otp, err := generateOTP()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// insert OTP into database
	_, err = c.Bot.GetDB().Exec(
		"INSERT INTO otps (otp, slack_id, action, action_email) VALUES (?, ?, ?, ?)",
		otp, userID, "link", email,
	)
	if err != nil {
		response := map[string]interface{}{
			"response_type": "ephemeral",
			"text":          "Database error occurred while creating OTP",
		}
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(response)
	}
	var otpLink = fmt.Sprintf("https://milkybot.alice.hackclub.app/otp/verify?otp=%s", otp)

	// send OTP email
	if err := utils.SendOtpEmail(email, otpLink); err != nil {
		response := map[string]interface{}{
			"response_type": "ephemeral",
			"text":          "Failed to send email, please try again later",
		}
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(response)
	}

	// success response
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"response_type": "ephemeral",
		"text":          fmt.Sprintf("Check your email (%s) for a link to confirm linking your Slack account.", email),
	}
	return json.NewEncoder(w).Encode(response)
}
