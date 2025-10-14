package structs

import "github.com/mehanizm/airtable"

type MilkywayBot struct {
	AirtableClient airtable.Client
	Port           string
}
