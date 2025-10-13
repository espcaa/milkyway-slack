package main

import (
	"errors"
	"log"
	"os"

	"github.com/mehanizm/airtable"
)

type MilkywayBot struct {
	AirtableClient airtable.Client
}

func main() {

	bot, err := createMilkywayBot()
	if err != nil {
		log.Fatal("Error using createMilkywayBot:", err)
	}

	bot.Run()

}

func createMilkywayBot() (*MilkywayBot, error) {

	var apiKey = os.Getenv("AIRTABLE_API_KEY")
	if apiKey == "" {
		return nil, errors.New("AIRTABLE_API_KEY environment variable is not set")
	}

	// Initialize Airtable client
	airtableClient := airtable.NewClient(apiKey)

	// Return the bot instance
	return &MilkywayBot{
		AirtableClient: *airtableClient,
	}, nil
}

func (bot *MilkywayBot) Run() {
	// Bot logic here

}
