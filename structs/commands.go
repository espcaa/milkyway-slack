package structs

import "net/http"

type Command interface {
	Run(w http.ResponseWriter, r *http.Request)
}

type Block struct {
	Type string `json:"type"`
	Text *Text  `json:"text,omitempty"`
}

type Text struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
