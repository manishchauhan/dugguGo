// response/response.go
package jsonResponse

import (
	"encoding/json"
	"net/http"

	"github.com/manishchauhan/dugguGo/models/userModel"
)

const (
	UserRegistered   = "User registered successfully"
	Methodnotallowed = "Method not allowed"
	Errordecoding    = "Error decoding request body: "
	DBinsertError    = "Error inserting user data: "
)

type ResponseMessage struct {
	Message     string `json:"message"` // msg if needed
	LoginStatus bool   `json:"status"`  //true if successful
	User        userModel.IFUser
}

// use this method for select or multiselect
func WriteJSONResponse(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.Write(jsonBytes)
	return err
}

// use this method for add, delete, update, etc.
func SendJSONResponse(w http.ResponseWriter, status int, message string) {
	response := ResponseMessage{Message: message}
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
