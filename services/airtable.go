package utils

import (
	"milkyway-slack/bot"
	"os"

	"github.com/mehanizm/airtable"
)

func GetTableFromName(m bot.MilkywayBot, name string) *airtable.Table {
	// get airtable database id from .env

	var dbId string = os.Getenv("AIRTABLE_BASE_ID")

	table := m.AirtableClient.GetTable(dbId, name)

	return table
}
