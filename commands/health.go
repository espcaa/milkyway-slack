package commands

import (
	"net/http"
)

type HealthCommand struct{}

func (c HealthCommand) Run(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hii ^-^!"))
}
