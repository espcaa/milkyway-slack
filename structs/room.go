package structs

type Room struct {
	Projects   []Project
	Furnitures []Furniture
	Floor      Floor
}

type Project struct {
	Egg_texture string
	Position    string
}

type Furniture struct {
	Texture  string
	Position string
}

type Floor struct {
	Texture string
}
