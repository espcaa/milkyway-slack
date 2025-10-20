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

// GenerateRoomImage creates the room image with base background, tiled floor, projects, and furniture.
func GenerateRoomImage(room structs.Room) (image.Image, error) {
	const (
		TileWidth = 96
		// FIX: Use integer arithmetic or explicitly cast the float result.
		// We'll calculate it using integer arithmetic by multiplying by 587 and dividing by 1000.
		// 96 * 0.587 = 56.352. (96 * 587) / 1000 = 56.352 (Go truncates the integer division to 56).
		TileHeight    = TileWidth * 587 / 1000
		FloorGridSize = 6

		// Estimated center of the room area within the final image canvas.
		CanvasCenterX = 377
		CanvasCenterY = 600
	)
	// --- 1. Load the base room image (ressources/room.png) ---
	baseRoomFile, err := os.Open("ressources/room.png")
	if err != nil {
		return nil, err
	}
	defer baseRoomFile.Close()
	baseRoomImg, err := png.Decode(baseRoomFile)
	if err != nil {
		return nil, err
	}

	// Create the canvas based on the base room image's bounds
	canvas := image.NewRGBA(baseRoomImg.Bounds())
	// Draw the base room image first (e.g., walls, shadows)
	draw.Draw(canvas, canvas.Bounds(), baseRoomImg, image.Point{}, draw.Src)

	// --- 2. Apply the ground tiling ---

	// Calculate the top-left corner of the isometric projection of the tile grid.
	// This empirically centers a 6x6 grid in the floor area.
	const GridOffsetFloorX = CanvasCenterX - (TileWidth / 2) - TileWidth
	const GridOffsetFloorY = CanvasCenterY + (FloorGridSize * TileHeight / 2)

	floorTextureName := room.Floor.Texture
	if floorTextureName == "" {
		floorTextureName = "wood" // Fallback
	}

	tileFile, err := os.Open("ressources/synced/floor/" + floorTextureName)
	if err == nil {
		defer tileFile.Close()
		tileImg, err := png.Decode(tileFile)
		if err == nil {
			// Iterate over the grid (i = y, j = x in your Svelte code)
			for i := 0; i < FloorGridSize; i++ {
				for j := 0; j < FloorGridSize; j++ {
					// Isometric coordinate calculation:
					// left: ((var(--x) - var(--y)) * var(--tile-width) / 2 )
					// top: ((var(--x) + var(--y)) * var(--tile-height) / 2)

					// x = j, y = i
					tileRelX := (j - i) * TileWidth
					tileRelY := (j + i) * TileHeight

					// Calculate the absolute position on the canvas
					absX := GridOffsetFloorX + tileRelX
					absY := GridOffsetFloorY + tileRelY

					pos := image.Pt(absX, absY)
					r := image.Rectangle{Min: pos, Max: pos.Add(tileImg.Bounds().Size())}

					// resize tileImg to TileWidth x TileHeight
					resizedTileImg := resize.Resize(TileWidth, TileHeight, tileImg, resize.Lanczos3)

					tileImg = resizedTileImg
					// Draw the tile over the base image
					draw.Draw(canvas, r, tileImg, image.Point{}, draw.Over)
				}
			}
		}
	} // Floor tiling complete

	// --- 3. Draw Projects ---
	// Items are placed using x, y which are direct pixel offsets from the room's center (CanvasCenterX, CanvasCenterY).
	for _, project := range room.Projects {
		if project.Egg_texture == "" || project.Position == "" {
			continue
		}

		projectFile, err := os.Open("ressources/synced" + project.Egg_texture)
		if err != nil {
			continue
		}

		// Use image.Decode as requested (no encoding step needed)
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

		// ==========================================================
		// ✨ RESIZE IMPLEMENTATION ✨
		// Scale the image so its width is TileWidth (96), preserving aspect ratio (height=0).
		// NOTE: This requires importing a resizing library like 'github.com/oliamb/resize'.

		resizedImg := resize.Resize(TileWidth, 0, projectImg, resize.Lanczos3)

		// Reassign projectImg to the new, resized image
		projectImg = resizedImg
		// ==========================================================

		// Calculate absolute position on the canvas: CanvasCenter + RelativePos - (ImageSize/2)
		// We use the bounds of the NOW RESIZED image.
		imgBounds := projectImg.Bounds()
		xAbs := CanvasCenterX + xRel - imgBounds.Dx()/2*2
		yAbs := CanvasCenterY + yRel - imgBounds.Dy()/2*2

		pos := image.Pt(xAbs, yAbs)
		r := image.Rectangle{Min: pos, Max: pos.Add(imgBounds.Size())}
		draw.Draw(canvas, r, projectImg, image.Point{}, draw.Over)
	}
	// --- 4. Draw Furnitures ---
	// Same 2D placement logic as projects.
	for _, furniture := range room.Furnitures {
		if furniture.Texture == "" || furniture.Position == "" {
			continue
		}

		furnFile, err := os.Open("ressources/synced/room/" + furniture.Texture + ".png")
		if err != nil {
			continue
		}

		// Use image.Decode as requested
		furnImg, _, err := image.Decode(furnFile)
		furnFile.Close()
		if err != nil {
			continue
		}

		parts := strings.Split(furniture.Position, ",")
		// Use only the first two parts for x, y, ignoring the optional 'flipped' state.
		if len(parts) < 2 {
			continue
		}
		xRel, err1 := strconv.Atoi(parts[0])
		yRel, err2 := strconv.Atoi(parts[1])

		xRel *= 2 // Furnitures use double scale for positioning
		yRel *= 2

		if err1 != nil || err2 != nil {
			continue
		}

		// ==========================================================
		// ✨ RESIZE IMPLEMENTATION FOR FURNITURE ✨
		// Scale the image so its width is TileWidth (96), preserving aspect ratio (height=0).
		// NOTE: This requires importing a resizing library like 'github.com/oliamb/resize'.

		resizedImg := resize.Resize(TileWidth, 0, furnImg, resize.Lanczos3)

		// Reassign furnImg to the new, resized image
		furnImg = resizedImg
		// ==========================================================

		// Calculate absolute position on the canvas: CanvasCenter + RelativePos - (ImageSize/2)
		// We use the bounds of the NOW RESIZED image.
		imgBounds := furnImg.Bounds()
		xAbs := CanvasCenterX + xRel - imgBounds.Dx()/2
		yAbs := CanvasCenterY + yRel - imgBounds.Dy()/2

		pos := image.Pt(xAbs, yAbs)
		r := image.Rectangle{Min: pos, Max: pos.Add(imgBounds.Size())}
		draw.Draw(canvas, r, furnImg, image.Point{}, draw.Over)
	}

	return canvas, nil
}
