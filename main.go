package main

import (
	"errors"
	"log"
	"milkyway-slack/bot"
	"os"

	"github.com/joho/godotenv"
	"github.com/mehanizm/airtable"
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
		AirtableClient: *airtableClient,
	}, nil
}
