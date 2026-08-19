package handlers

import (
	"log"
	"net/http"
)

func serverError(w http.ResponseWriter, err error, clientMessage string) {
	log.Printf("%s: %v", clientMessage, err)
	http.Error(w, clientMessage, http.StatusInternalServerError)
}
