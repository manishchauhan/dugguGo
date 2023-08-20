// errorhandler/errorhandler.go
package errorhandler

import (
	"encoding/json"
	"net/http"
)

func SendErrorResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	errorResponse := map[string]interface{}{
		"errorCode": code,
		"errorMsg":  message,
	}
	json.NewEncoder(w).Encode(errorResponse)
}
