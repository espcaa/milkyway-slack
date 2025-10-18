package utils

import (
	"milkyway-slack/structs"
	"os"
)

func GetRoomData(bot structs.BotInterface, userRecordId string) (structs.Room, error) {
	room := structs.Room{
		Projects: make([]structs.Project, 0),
	}

	var dbID = os.Getenv("AIRTABLE_BASE_ID")

	// get the user record with the user_record id

	var UserTable = bot.GetAirtableClient().GetTable(dbID, "User")
	userRecord, err := UserTable.GetRecord(userRecordId)
	if err != nil {
		return room, err
	}

	// now go to the user_record "projects" and "Furniture" fields to get the linked records to populate the room data

	// Projects
	if projectsField, ok := userRecord.Fields["projects"].([]interface{}); ok {
		var ProjectTable = bot.GetAirtableClient().GetTable(dbID, "Project")
		for _, projectID := range projectsField {
			if projectIDStr, ok := projectID.(string); ok {
				projectRecord, err := ProjectTable.GetRecord(projectIDStr)
				if err != nil {
					continue
				}
				project := structs.Project{}
				if eggTexture, ok := projectRecord.Fields["egg_texture"].(string); ok {
					project.Egg_texture = eggTexture
				}
				if position, ok := projectRecord.Fields["position"].(string); ok {
					project.Position = position
				}
				room.Projects = append(room.Projects, project)
			}
		}
	}

	// Furnitures
	room.Furnitures = make([]structs.Furniture, 0)
	if furnituresField, ok := userRecord.Fields["furnitures"].([]interface{}); ok {
		var FurnitureTable = bot.GetAirtableClient().GetTable(dbID, "Furniture")
		for _, furnitureID := range furnituresField {
			if furnitureIDStr, ok := furnitureID.(string); ok {
				furnitureRecord, err := FurnitureTable.GetRecord(furnitureIDStr)
				if err != nil {
					continue
				}
				furniture := structs.Furniture{}
				if texture, ok := furnitureRecord.Fields["texture"].(string); ok {
					furniture.Texture = texture
				}
				if position, ok := furnitureRecord.Fields["position"].(string); ok {
					furniture.Position = position
				}
				room.Furnitures = append(room.Furnitures, furniture)
			}
		}
	}

	// Floor
	room.Floor = structs.Floor{}
	room.Floor.Texture = "wood.png"

	return room, nil
}
