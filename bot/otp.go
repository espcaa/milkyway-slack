package bot

import (
	"database/sql"
	"fmt"
	"net/http"
)

func HandleOtpVerifying(w http.ResponseWriter, r *http.Request, bot *MilkywayBot) {
	otp := r.URL.Query().Get("otp")
	if otp == "" {
		http.Error(w, "OTP is required", http.StatusBadRequest)
		return
	}

	// Query the OTP record
	var slackID, actionEmail, action string
	err := bot.GetDB().QueryRow(
		"SELECT slack_id, action_email, action FROM otps WHERE otp = ?",
		otp,
	).Scan(&slackID, &actionEmail, &action)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid OTP", http.StatusBadRequest)
		} else {
			http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Perform the action
	switch action {
	case "link":
		_, err := bot.GetDB().Exec(
			"INSERT INTO user_overrides (slack_id, email) VALUES (?, ?)",
			slackID, actionEmail,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to link account: %v", err), http.StatusInternalServerError)
			return
		}
		// return some text
		w.Write([]byte(fmt.Sprintf("Successfully linked Slack account to email: %s", actionEmail)))
		return
	case "unlink":
		_, err := bot.GetDB().Exec(
			"DELETE FROM user_overrides WHERE slack_id = ? AND email = ?",
			slackID, actionEmail,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to unlink account: %v", err), http.StatusInternalServerError)
			return
		}
		// return some text
		w.Write([]byte(fmt.Sprintf("Successfully unlinked Slack account from email: %s", actionEmail)))
		return
	default:
		http.Error(w, "Unknown OTP action", http.StatusBadRequest)
		return
	}

	// Delete the OTP from the DB
	_, err = bot.GetDB().Exec("DELETE FROM otps WHERE otp = ?", otp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete OTP: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OTP verified successfully"))
}
