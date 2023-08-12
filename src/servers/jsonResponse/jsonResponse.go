// response/response.go
package jsonResponse

import (
	"encoding/json"
	"net/http"
)

func WriteJSONResponse(w http.ResponseWriter, code int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.Write(jsonBytes)
	return err
}
