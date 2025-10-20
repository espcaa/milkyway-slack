package utils

import (
	"image"
	"image/draw"
	"image/png"
	"milkyway-slack/structs"
	"os"
	"strconv"
	"strings"
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
	if projectsField, ok := userRecord.Fields["projects"].([]any); ok {
		var ProjectTable = bot.GetAirtableClient().GetTable(dbID, "Projects")
		for _, projectID := range projectsField {
			if projectIDStr, ok := projectID.(string); ok {
				projectRecord, err := ProjectTable.GetRecord(projectIDStr)
				if err != nil {
					continue
				}
				project := structs.Project{}
				if eggTexture, ok := projectRecord.Fields["egg"].(string); ok {
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
	if furnituresField, ok := userRecord.Fields["Furniture"].([]interface{}); ok {
		var FurnitureTable = bot.GetAirtableClient().GetTable(dbID, "Furniture")
		for _, furnitureID := range furnituresField {
			if furnitureIDStr, ok := furnitureID.(string); ok {
				furnitureRecord, err := FurnitureTable.GetRecord(furnitureIDStr)
				if err != nil {
					continue
				}
				furniture := structs.Furniture{}
				if texture, ok := furnitureRecord.Fields["type"].(string); ok {
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

func GenerateRoomImage(room structs.Room) (image.Image, error) {
	floorFile, err := os.Open("ressources/synced/floor/" + room.Floor.Texture)
	if err != nil {
		return nil, err
	}
	defer floorFile.Close()
	floorImg, err := png.Decode(floorFile)
	if err != nil {
		return nil, err
	}

	canvas := image.NewRGBA(floorImg.Bounds())
	draw.Draw(canvas, canvas.Bounds(), floorImg, image.Point{}, draw.Over)

	for _, project := range room.Projects {
		if project.Egg_texture == "" || project.Position == "" {
			continue
		}

		projectFile, err := os.Open("ressources/synced/projects/" + project.Egg_texture)
		if err != nil {
			continue
		}
		projectImg, err := png.Decode(projectFile)
		projectFile.Close()
		if err != nil {
			continue
		}

		parts := strings.Split(project.Position, ",")
		if len(parts) != 2 {
			continue
		}
		x, err1 := strconv.Atoi(parts[0])
		y, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			continue
		}

		pos := image.Pt(x, y)
		r := image.Rectangle{Min: pos, Max: pos.Add(projectImg.Bounds().Size())}
		draw.Draw(canvas, r, projectImg, image.Point{}, draw.Over)
	}

	for _, furniture := range room.Furnitures {
		if furniture.Texture == "" || furniture.Position == "" {
			continue
		}

		furnFile, err := os.Open("ressources/synced/room/" + furniture.Texture)
		if err != nil {
			continue
		}
		furnImg, err := png.Decode(furnFile)
		furnFile.Close()
		if err != nil {
			continue
		}

		parts := strings.Split(furniture.Position, ",")
		if len(parts) != 2 {
			continue
		}
		x, err1 := strconv.Atoi(parts[0])
		y, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			continue
		}

		pos := image.Pt(x, y)
		r := image.Rectangle{Min: pos, Max: pos.Add(furnImg.Bounds().Size())}
		draw.Draw(canvas, r, furnImg, image.Point{}, draw.Over)
	}

	return canvas, nil
}
