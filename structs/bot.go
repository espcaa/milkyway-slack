package structs

import (
	"database/sql"

	"github.com/mehanizm/airtable"
)

type BotInterface interface {
	GetAirtableClient() *airtable.Client
	GetDB() *sql.DB
}
