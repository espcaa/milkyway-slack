package structs

import "net/http"

type Command interface {
	Run(w http.ResponseWriter, r *http.Request) (err error)
}

type Block struct {
	Type     string `json:"type"`
	Text     *Text  `json:"text,omitempty"`
	ImageURL string `json:"image_url,omitempty"` // for image blocks
	AltText  string `json:"alt_text,omitempty"`  // for image blocks
	Title    *Text  `json:"title,omitempty"`     // optional title for image blocks
}

type Text struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
