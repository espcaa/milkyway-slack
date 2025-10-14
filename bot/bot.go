package bot

import (
	"log"
	"milkyway-slack/commands"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mehanizm/airtable"
)

type MilkywayBot struct {
	AirtableClient airtable.Client
	Port           string
}

func (bot MilkywayBot) Run() {
	// Bot logic here

	r := chi.NewRouter()

	// add the post endpoints for each command
	for name, cmd := range commands.CommandRegistry {

		log.Println("Registering command:", name)

		cmdCopy := cmd
		r.Post("/commands/"+name, func(w http.ResponseWriter, r *http.Request) {
			log.Println("Received request for command:", name)
			cmdCopy.Run(w, r)
		})
	}

	log.Println("Server starting at http://localhost:" + bot.Port)
	http.ListenAndServe(":"+bot.Port, r)
}
