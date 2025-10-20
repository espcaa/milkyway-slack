package main

import (
	"database/sql"
	"errors"
	"log"
	"milkyway-slack/bot"
	"os"

	"github.com/joho/godotenv"
	"github.com/mehanizm/airtable"
	_ "modernc.org/sqlite"
)

func main() {

	godotenv.Load()

	bot, err := createMilkywayBot()
	if err != nil {
		log.Fatal("Error using createMilkywayBot:", err)
	}

	bot.Run()

}

func createMilkywayBot() (*bot.MilkywayBot, error) {

	db, err := sql.Open("sqlite", "file:milkyway.db?_foreign_keys=on")
	if err != nil {
		return nil, errors.New("failed to open SQLite database: " + err.Error())
	}

	err = db.Ping()

	if err != nil {
		return nil, errors.New("failed to ping SQLite database: " + err.Error())
	}

	// if the table user_overrides doesn't exist, create it
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS user_overrides (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		slack_id TEXT UNIQUE,
		email TEXT
	);
	`)

	if err != nil {
		return nil, errors.New("failed to connect to SQLite database: " + err.Error())
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS otps (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			otp TEXT UNIQUE,
			slack_id TEXT,
			action TEXT,
			action_email TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		`)
	if err != nil {
		return nil, errors.New("failed to create otps table: " + err.Error())
	}

	var port = os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	var apiKey = os.Getenv("AIRTABLE_API_KEY")
	if apiKey == "" {
		return nil, errors.New("AIRTABLE_API_KEY environment variable is not set")
	}

	// Initialize Airtable client
	airtableClient := airtable.NewClient(apiKey)

	// Return the bot instance
	return &bot.MilkywayBot{
		AirtableClient: airtableClient,
		Port:           port,
		Sqlite:         db,
	}, nil
}
