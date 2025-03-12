package server

import (
	"net/http"
)

// respondJSON sets the appropriate headers and writes JSON data to the response
func respondJSON(w http.ResponseWriter, bytes []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}