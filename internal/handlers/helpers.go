package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func serverError(w http.ResponseWriter, err error, clientMessage string) {
	log.Printf("%s: %v", clientMessage, err)
	http.Error(w, clientMessage, http.StatusInternalServerError)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	return json.NewDecoder(r.Body).Decode(dst)
}
