// errorhandler/errorhandler.go
package errorhandler

import (
	"encoding/json"
	"net/http"
)

func WriteJSONError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	errorResponse := map[string]interface{}{
		"errorCode": code,
		"errorMsg":  message,
	}
	json.NewEncoder(w).Encode(errorResponse)
}

func HandleDatabaseError(w http.ResponseWriter, err error) {
	WriteJSONError(w, http.StatusInternalServerError, "Database error")
}

func HandleInternalError(w http.ResponseWriter, err error) {
	WriteJSONError(w, http.StatusInternalServerError, "Internal server error")
}
