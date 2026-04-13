package utils

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func WriteError(w http.ResponseWriter, status int, message string, fields map[string]string) {
	resp := map[string]any{"error": message}
	if len(fields) > 0 {
		resp["fields"] = fields
	}
	WriteJSON(w, status, resp)
}
