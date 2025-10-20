package bot

import (
	"database/sql"
	"log"
	"milkyway-slack/commands"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mehanizm/airtable"
)

type MilkywayBot struct {
	AirtableClient *airtable.Client
	Port           string
	Sqlite         *sql.DB
}

func (bot *MilkywayBot) GetAirtableClient() *airtable.Client {
	return bot.AirtableClient
}

func (bot *MilkywayBot) GetDB() *sql.DB {
	return bot.Sqlite
}

func (bot *MilkywayBot) Run() {
	// Bot logic here

	r := chi.NewRouter()

	commands.CommandRegistry = commands.InitCommands(bot)

	// add the post endpoints for each command
	for name, cmd := range commands.CommandRegistry {

		log.Println("Registering command:", name)

		cmdCopy := cmd
		r.Post("/commands/"+name, func(w http.ResponseWriter, r *http.Request) {
			log.Println("Received request for command:", name)
			err := cmdCopy.Run(w, r)
			if err != nil {
				log.Println("Error executing command", name, ":", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			} else {
				log.Println("Successfully executed command:", name)
			}
		})

	}

	r.Get("/otp/verify", func(w http.ResponseWriter, r *http.Request) {
		HandleOtpVerifying(w, r, bot)
	})

	log.Println("Server starting at http://localhost:" + bot.Port)
	http.ListenAndServe(":"+bot.Port, r)
}
