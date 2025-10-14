package utils

import (
	"fmt"
	"milkyway-slack/structs"
	"os"
)

func GetRoomData(bot structs.BotInterface, userRecordId string) (structs.Room, error) {
	room := structs.Room{
		Projects: make([]structs.Project, 0),
	}

	var dbID = os.Getenv("AIRTABLE_BASE_ID")
	// Get the projects table
	projectsTable := bot.GetAirtableClient().GetTable(dbID, "Projects")

	formula := fmt.Sprintf(`FIND("%s", ARRAYJOIN({Users}))`, userRecordId)

	projectRecords, err := projectsTable.GetRecords().
		WithFilterFormula(formula).
		ReturnFields("egg", "position").
		Do()
	if err != nil {
		return structs.Room{}, fmt.Errorf("failed to get project records: %w", err)
	}

	if err != nil {
		return structs.Room{}, fmt.Errorf("failed to get project records: %w", err)
	}

	for _, rec := range projectRecords.Records {
		eggTexture, ok := rec.Fields["egg"].(string)
		if !ok {
			continue
		}

		position, ok := rec.Fields["position"].(string)
		if !ok {
			continue
		}

		room.Projects = append(room.Projects, structs.Project{
			Egg_texture: eggTexture,
			Position:    position,
		})
	}

	return room, nil
}
