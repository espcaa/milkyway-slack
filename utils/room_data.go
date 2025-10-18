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

	formula := fmt.Sprintf(`{user} = RECORD_ID('%s')`, userRecordId)

	projectRecords, err := projectsTable.GetRecords().
		WithFilterFormula(formula).
		ReturnFields("egg", "position").
		Do()

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

	// now get furniture

	furnitureTable := bot.GetAirtableClient().GetTable(dbID, "Furniture")
	furnitureRecords, err := furnitureTable.GetRecords().
		WithFilterFormula(formula).
		ReturnFields("type", "position").
		Do()

	if err != nil {
		return structs.Room{}, fmt.Errorf("failed to get furniture records: %w", err)
	}

	for _, rec := range furnitureRecords.Records {
		texture, ok := rec.Fields["type"].(string)
		if !ok {
			continue
		}

		position, ok := rec.Fields["position"].(string)
		if !ok {
			continue
		}

		room.Furnitures = append(room.Furnitures, structs.Furniture{
			Texture:  texture,
			Position: position,
		})
	}

	room.Floor.Texture = "wood.png"

	return room, nil
}
