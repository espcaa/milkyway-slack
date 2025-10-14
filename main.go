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

	if err != nil {
		return nil, errors.New("failed to connect to SQLite database: " + err.Error())
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
