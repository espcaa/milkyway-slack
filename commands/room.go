package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"milkyway-slack/structs"
	"milkyway-slack/utils"
	"net/http"
	"os"
	"strings"
)

type RoomCommand struct {
	Bot structs.BotInterface
}

func (c RoomCommand) Run(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("failed to parse form: %w", err)
	}

	responseURL := r.PostFormValue("response_url")
	userID := r.PostFormValue("user_id")
	if responseURL == "" {
		return fmt.Errorf("no response_url provided")
	}

	// Send immediate response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"text": "Processing room command...",
	}); err != nil {
		return fmt.Errorf("failed to send immediate response: %w", err)
	}

	go func(url string, userID string, bot structs.BotInterface) {
		// get the user's airtable record
		email, err := getUserEmail(userID, bot)
		if err != nil {
			log.Printf("Error getting user email: %v", err)
			sendErrorResponse(url, "Failed to get your email :(")
			return
		}

		recordID, err := getUserRecordID(email, bot)
		if err != nil {
			log.Printf("Error getting user record ID: %v", err)
			sendErrorResponse(url, "Failed to find your user record :(")
			return
		}

		log.Printf("User email: %s, Record ID: %s", email, recordID)

		room, err := utils.GetRoomData(bot, recordID)

		if err != nil {
			log.Printf("Error getting room data: %v", err)
			sendErrorResponse(url, "Failed to get your room data :(")
			return
		}

		// transform room data into a message

		var markdownRoomInfo string
		for _, value := range room.Projects {
			markdownRoomInfo += fmt.Sprintf("  * Project: %s\n", value.Egg_texture)
		}
		for _, value := range room.Furnitures {
			markdownRoomInfo += fmt.Sprintf("  * Furniture: %s\n", value.Texture)
		}
		if markdownRoomInfo == "" {
			markdownRoomInfo = "Your room is empty. Start adding projects!"
		}

		blocks := []Block{
			{
				Type: "section",
				Text: &Text{
					Type: "mrkdwn",
					Text: markdownRoomInfo,
				},
			},
		}

		payload := map[string]any{
			"response_type": "in_channel",
			"blocks":        blocks,
			"text":          "room",
		}

		if err := sendSlackResponse(url, payload); err != nil {
			log.Printf("Error sending final response: %v", err)
			sendErrorResponse(url, "Failed to send room message :(")
		}
	}(responseURL, userID, c.Bot)

	return nil
}

func getUserEmail(userID string, bot structs.BotInterface) (string, error) {
	// check in the override sqlite db (else check in the airtable)

	var email string

	// db query here to get email from userID
	// if found, return it else continue

	err := bot.GetDB().QueryRow(
		"SELECT email FROM user_overrides WHERE slack_id = ?", userID,
	).Scan(&email)

	if err == nil {
		return email, nil
	}

	// now get it from slack api

	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		return "", fmt.Errorf("SLACK_TOKEN not set")
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://slack.com/api/users.info?user="+userID, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+slackToken)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		OK   bool `json:"ok"`
		User struct {
			Profile struct {
				Email string `json:"email"`
			} `json:"profile"`
		} `json:"user"`
		Error string `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.OK {
		return "", fmt.Errorf("slack api error: %s", result.Error)
	}

	email = result.User.Profile.Email
	if email == "" {
		return "", fmt.Errorf("email not found in slack profile")
	}

	// store in override db for future use

	_, err = bot.GetDB().Exec(
		"INSERT INTO user_overrides (slack_id, email) VALUES (?, ?)",
		userID, email,
	)

	if err != nil {
		log.Printf("failed to insert user override: %v", err)
	}

	return email, nil
}

func getUserRecordID(email string, bot structs.BotInterface) (string, error) {

	var recordID string
	var dbID = os.Getenv("AIRTABLE_BASE_ID")
	if dbID == "" {
		return "", fmt.Errorf("AIRTABLE_BASE_ID not set")
	}

	var UserTable = bot.GetAirtableClient().GetTable("User", dbID)
	records, err := UserTable.GetRecords().
		WithFilterFormula(fmt.Sprintf(`{email}='%s'`, strings.ReplaceAll(email, "'", "\\'"))).
		ReturnFields("Name").
		Do()
	if err != nil {
		return "", fmt.Errorf("failed to get user records: %w", err)
	}

	if len(records.Records) == 0 {
		return "", fmt.Errorf("no user record found for email: %s", email)
	}

	recordID = records.Records[0].ID

	return recordID, nil
}
