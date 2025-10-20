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
	// --- 1. Load the base room image (ressources/room.png) ---
	baseRoomFile, err := os.Open("ressources/room.png")
	if err != nil {
		// Return an error if the base image can't be opened
		return nil, err
	}
	defer baseRoomFile.Close()
	baseRoomImg, err := png.Decode(baseRoomFile)
	if err != nil {
		// Return an error if the base image can't be decoded
		return nil, err
	}

	// Create the canvas based on the base room image's bounds
	canvas := image.NewRGBA(baseRoomImg.Bounds())
	// Draw the base room image first (e.g., walls, shadows)
	draw.Draw(canvas, canvas.Bounds(), baseRoomImg, image.Point{}, draw.Src)

	// --- 2. Apply the ground tiling ---

	// Determine the starting position (origin for the tile grid).
	// For a 6x6 grid, the center of the rhombus (at the middle-most tile) should be near the canvas center.
	// We'll calculate the top-left offset to center the grid.
	// The rhombus width at its widest point is FloorGridSize * TileWidth.
	// The rhombus height is FloorGridSize * TileHeight.

	// Assuming the canvas center is (CanvasWidth/2, CanvasHeight/2).
	// We need to offset the entire grid to be centered on the floor area of the room.png.
	// This offset will be highly dependent on the 'room.png' dimensions and content.
	// A common approach for a centered 6x6 grid:
	// The top point of the 6x6 rhombus is at (3 * TileWidth, 0) relative to its container's top-left.
	// The container's top-left in the Svelte CSS has a margin-left of -1 * TileWidth.
	// Let's use a manual offset based on a typical room layout:
	// Assuming room.png is 700px high (like the Svelte CSS) and ~700-800px wide.
	const CanvasCenterX = 350 // Estimate half of a 700px wide canvas
	const CanvasCenterY = 300 // Estimate the y-coordinate of the room floor's center

	// Calculate the top-left corner of the isometric projection of the tile grid.
	// (0,0) tile position (relative to the grid container) should be centered.
	// Container's origin (for the tile grid) should be:
	// X: CanvasCenterX - (GridTotalWidth / 2) + MarginAdjustment (from Svelte CSS `margin-left: calc(-1 * var(--tile-width));`)
	// GridTotalWidth is FloorGridSize * TileWidth
	// Let's simplify and use an empirical offset based on a centered 6x6:
	const GridOffsetFloorX = CanvasCenterX - (TileWidth / 2) - TileWidth // Adjust for 6x6 center and Svelte -1 * TileWidth margin
	const GridOffsetFloorY = CanvasCenterY - (FloorGridSize * TileHeight / 2)

	// Use a hardcoded default floor texture if room.Floor.Texture is empty, or assume it's set.
	floorTextureName := room.Floor.Texture
	if floorTextureName == "" {
		floorTextureName = "wood" // Fallback to 'wood'
	}

	// Pre-load the floor tile image (assuming all tiles in the grid use the same one)
	tileFile, err := os.Open("ressources/synced/floor/" + floorTextureName + ".png")
	if err != nil {
		// Log the error but continue if floor tile can't be found (it won't render)
		// Or you might return the error, depending on how critical the floor is.
		// For this example, we'll continue with the base room image if the specific floor file fails.
		// return nil, err // Uncomment this if floor is critical
	} else {
		defer tileFile.Close()
		tileImg, err := png.Decode(tileFile)
		if err != nil {
			// Continue if decoding fails
		} else {
			// Iterate over the grid (i = y, j = x in your Svelte code)
			for i := 0; i < FloorGridSize; i++ {
				for j := 0; j < FloorGridSize; j++ {
					// Isometric coordinate calculation, matching the Svelte CSS:
					// left: ((var(--x) - var(--y)) * var(--tile-width) / 2 )
					// top: ((var(--x) + var(--y)) * var(--tile-height) / 2)

					// x = j, y = i
					tileRelX := (j - i) * TileWidth / 2
					tileRelY := (j + i) * TileHeight / 2

					// Calculate the absolute position on the canvas
					absX := GridOffsetFloorX + tileRelX
					absY := GridOffsetFloorY + tileRelY

					pos := image.Pt(absX, absY)
					r := image.Rectangle{Min: pos, Max: pos.Add(tileImg.Bounds().Size())}

					// Draw the tile
					draw.Draw(canvas, r, tileImg, image.Point{}, draw.Over)
				}
			}
		}
	}

	// --- 3. Draw Projects ---
	// The projects' x, y are relative to the room's center, which we defined as (CanvasCenterX, CanvasCenterY)
	// These coordinates are the same ones clamped by the Svelte drag logic.
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
		// The coordinates here (x, y) are relative to the room's center (0, 0) in the Svelte code.
		xRel, err1 := strconv.Atoi(parts[0])
		yRel, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			continue
		}

		// Calculate absolute position on the canvas: CanvasCenter + RelativePos - (ImageSize/2)
		// We subtract half the image size to center the image at the coordinate (xRel, yRel).
		imgBounds := projectImg.Bounds()
		xAbs := CanvasCenterX + xRel - imgBounds.Dx()/2
		yAbs := CanvasCenterY + yRel - imgBounds.Dy()/2

		pos := image.Pt(xAbs, yAbs)
		r := image.Rectangle{Min: pos, Max: pos.Add(imgBounds.Size())}
		draw.Draw(canvas, r, projectImg, image.Point{}, draw.Over)
	}

	// --- 4. Draw Furnitures ---
	// Same logic as projects, using the room center as the origin.
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
		// Note: The furniture position might also include 'flipped' (e.g., "x,y,0") based on your Svelte code.
		// For simplicity, we only parse x, y here, assuming the first two parts.
		if len(parts) < 2 {
			continue
		}
		xRel, err1 := strconv.Atoi(parts[0])
		yRel, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			continue
		}

		// Calculate absolute position on the canvas: CanvasCenter + RelativePos - (ImageSize/2)
		imgBounds := furnImg.Bounds()
		xAbs := CanvasCenterX + xRel - imgBounds.Dx()/2
		yAbs := CanvasCenterY + yRel - imgBounds.Dy()/2

		pos := image.Pt(xAbs, yAbs)
		r := image.Rectangle{Min: pos, Max: pos.Add(imgBounds.Size())}

		// Note: If you need to handle `flipped`, you would apply a transform here before drawing.
		draw.Draw(canvas, r, furnImg, image.Point{}, draw.Over)
	}

	return canvas, nil
}
