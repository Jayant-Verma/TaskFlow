package utils

import (
	"encoding/json"
	"errors"
	"io"
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

func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			WriteError(w, http.StatusBadRequest, err.Error(), nil)
			return false
		}
		WriteError(w, http.StatusBadRequest, err.Error(), nil)
		return false
	}
	return true
}
