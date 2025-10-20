package commands

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"milkyway-slack/structs"
	"milkyway-slack/utils"
	"net/http"
)

type UnLinkCommand struct {
	Bot structs.BotInterface
}

// generate a long cryptic OTP
func generateOTP() (string, error) {
	b := make([]byte, 32) // 32 bytes = 64 hex characters
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (c UnLinkCommand) Run(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("failed to parse form: %w", err)
	}

	userID := r.PostFormValue("user_id")

	otp, err := generateOTP()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// get user email from existing override
	_, err = c.Bot.GetDB().Exec(
		"SELECT email FROM user_overrides WHERE slack_id = ?",
		userID,
	)
	if err != nil {
		response := map[string]interface{}{
			"response_type": "ephemeral",
			"text":          "No linked account found to unlink.",
		}
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(response)
	}
	var userEmail string
	err = c.Bot.GetDB().QueryRow(
		"SELECT email FROM user_overrides WHERE slack_id = ?",
		userID,
	).Scan(&userEmail)
	if err != nil {
		response := map[string]interface{}{
			"response_type": "ephemeral",
			"text":          "Database error occurred while retrieving email",
		}
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(response)
	}

	// store OTP in database

	_, err = c.Bot.GetDB().Exec(
		"INSERT INTO otps (otp, slack_id, action, action_email) VALUES (?, ?, ?, ?)",
		otp, userID, "unlink", userEmail,
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
	if err := utils.SendOtpEmail(userEmail, otpLink); err != nil {
		response := map[string]interface{}{
			"response_type": "ephemeral",
			"text":          "Failed to send email, please try again later",
		}
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(response)
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"response_type": "ephemeral",
		"text":          "Check your email for a link to confirm unlinking your account.",
	}
	return json.NewEncoder(w).Encode(response)
}
