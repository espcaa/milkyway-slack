package utils

import (
	"image"
	"image/draw"
	"image/png"
	"milkyway-slack/structs"
	"os"
	"strconv"
	"strings"

	"github.com/nfnt/resize"
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
	const (
		TileWidth     = 96
		TileHeight    = TileWidth * 587 / 1000
		FloorGridSize = 6
		CanvasCenterX = 377
		CanvasCenterY = 600
	)
	baseRoomFile, err := os.Open("ressources/room.png")
	if err != nil {
		return nil, err
	}
	defer baseRoomFile.Close()
	baseRoomImg, err := png.Decode(baseRoomFile)
	if err != nil {
		return nil, err
	}

	canvas := image.NewRGBA(baseRoomImg.Bounds())
	draw.Draw(canvas, canvas.Bounds(), baseRoomImg, image.Point{}, draw.Src)

	const GridOffsetFloorY = CanvasCenterY - (TileHeight * FloorGridSize / 2)
	const GridOffsetFloorX = CanvasCenterX - (TileWidth * FloorGridSize / 2)

	floorTextureName := room.Floor.Texture
	if floorTextureName == "" {
		floorTextureName = "wood"
	}

	tileFile, err := os.Open("ressources/synced/floor/" + floorTextureName)
	if err == nil {
		defer tileFile.Close()
		tileImg, err := png.Decode(tileFile)
		if err == nil {
			// Pre-resize once
			resizedTileImg := resize.Resize(TileWidth, TileHeight, tileImg, resize.Lanczos3)
			tileBounds := resizedTileImg.Bounds()
			tileW, tileH := tileBounds.Dx(), tileBounds.Dy()

			for i := 0; i < FloorGridSize; i++ {
				for j := 0; j < FloorGridSize; j++ {
					tileRelX := (j - i) * tileW / 2
					tileRelY := (j + i) * tileH / 2

					absX := GridOffsetFloorX + tileRelX
					absY := GridOffsetFloorY + tileRelY

					pos := image.Pt(absX, absY)
					r := image.Rectangle{Min: pos, Max: pos.Add(tileBounds.Size())}
					draw.Draw(canvas, r, resizedTileImg, image.Point{}, draw.Over)
				}
			}
		}
	}
	for _, project := range room.Projects {
		if project.Egg_texture == "" || project.Position == "" {
			continue
		}

		projectFile, err := os.Open("ressources/synced" + project.Egg_texture)
		if err != nil {
			continue
		}

		projectImg, _, err := image.Decode(projectFile)
		projectFile.Close()
		if err != nil {
			continue
		}

		parts := strings.Split(project.Position, ",")
		if len(parts) != 2 {
			continue
		}
		xRel, err1 := strconv.Atoi(parts[0])
		yRel, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			continue
		}

		resizedImg := resize.Resize(TileWidth*2, 0, projectImg, resize.Lanczos3)

		projectImg = resizedImg

		imgBounds := projectImg.Bounds()
		xAbs := CanvasCenterX + xRel - imgBounds.Dx()/2*2
		yAbs := CanvasCenterY + yRel - imgBounds.Dy()/2*2

		pos := image.Pt(xAbs, yAbs)
		r := image.Rectangle{Min: pos, Max: pos.Add(imgBounds.Size())}
		draw.Draw(canvas, r, projectImg, image.Point{}, draw.Over)
	}

	for _, furniture := range room.Furnitures {
		if furniture.Texture == "" || furniture.Position == "" {
			continue
		}

		furnFile, err := os.Open("ressources/synced/room/" + furniture.Texture + ".png")
		if err != nil {
			continue
		}

		furnImg, _, err := image.Decode(furnFile)
		furnFile.Close()
		if err != nil {
			continue
		}

		parts := strings.Split(furniture.Position, ",")
		if len(parts) < 2 {
			continue
		}
		xRel, err1 := strconv.Atoi(parts[0])
		yRel, err2 := strconv.Atoi(parts[1])

		if err1 != nil || err2 != nil {
			continue
		}

		resizedImg := resize.Resize(TileWidth*2, 0, furnImg, resize.Lanczos3)

		furnImg = resizedImg

		imgBounds := furnImg.Bounds()
		xAbs := CanvasCenterX + xRel - imgBounds.Dx()/2
		yAbs := CanvasCenterY + yRel - imgBounds.Dy()/2

		pos := image.Pt(xAbs, yAbs)
		r := image.Rectangle{Min: pos, Max: pos.Add(imgBounds.Size())}
		draw.Draw(canvas, r, furnImg, image.Point{}, draw.Over)
	}

	return canvas, nil
}
