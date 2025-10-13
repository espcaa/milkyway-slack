package structs

import "net/http"

type Command interface {
	Run(w http.ResponseWriter, r *http.Request)
}
