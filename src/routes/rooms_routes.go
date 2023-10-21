package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/manishchauhan/dugguGo/models/roomModel"
	"github.com/manishchauhan/dugguGo/servers/errorhandler"
	"github.com/manishchauhan/dugguGo/servers/jsonResponse"
	"github.com/manishchauhan/dugguGo/util/auth/jwtAuth"
	"github.com/manishchauhan/dugguGo/util/mysqlDbManager"
)

func RegisterRoomsRoutes(router *mux.Router, dm *mysqlDbManager.DBManager) {
	subrouter := router.PathPrefix("/chatrooms").Subrouter()
	subrouter.Handle("/add", jwtAuth.AuthMiddleware(http.HandlerFunc(addRoom(dm)))).Methods("POST")  //add new room
	subrouter.Handle("/list", jwtAuth.AuthMiddleware(http.HandlerFunc(getRooms(dm)))).Methods("GET") //add new room
	subrouter.HandleFunc("/delete", deleteRooms(dm)).Methods("POST")                                 //multiple rooms can be deleted or only one can be deleted
	subrouter.HandleFunc("/edit", editRoom(dm)).Methods("POST")                                      //based on the room id
	//subrouter.HandleFunc("/select", selectRooms(dm)).Methods("POST")' //based on the room id
}

// edit already exits room
func editRoom(dm *mysqlDbManager.DBManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorhandler.SendErrorResponse(w, http.StatusMethodNotAllowed, jsonResponse.Methodnotallowed)
			return
		}
		var updateRequest struct {
			Chatroom_id      int    `json:"chatroom_id"`
			Chatroom_name    string `json:"chatroom_name"`
			Chatroom_details string `json:"chatroom_details"`
		}
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&updateRequest); err != nil {
			errorhandler.SendErrorResponse(w, http.StatusBadRequest, jsonResponse.Errordecoding+err.Error())
			return
		}

		// Define the update columns and their values
		updateColumns := map[string]interface{}{
			"chatroom_details": updateRequest.Chatroom_details,
			"chatroom_name":    updateRequest.Chatroom_name,
		}
		// Define the conditions to update the specific chatroom by its chatroom_id
		conditions := "chatroom_id = ?"
		// Execute the update query
		_, err := dm.ExecuteUpdateWithWhere("chatroom", updateColumns, conditions, updateRequest.Chatroom_id)
		if err != nil {
			// Handle the error
			errorhandler.SendErrorResponse(w, http.StatusInternalServerError, "Failed to update chatroom")
			return
		}

		successResponse := jsonResponse.ResponseMessage{
			LoginStatus: true,
			Message:     jsonResponse.ChatRoomUpdated,
		}

		// Send back a success response if everything was successful
		jsonResponse.WriteJSONResponse(w, http.StatusOK, successResponse)
	}
}

// delete room or rooms
func deleteRooms(dm *mysqlDbManager.DBManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorhandler.SendErrorResponse(w, http.StatusMethodNotAllowed, jsonResponse.Methodnotallowed)
			return
		}

		var chatroomIDs []int
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&chatroomIDs); err != nil {
			errorhandler.SendErrorResponse(w, http.StatusBadRequest, jsonResponse.Errordecoding+err.Error())
			return
		}

		if len(chatroomIDs) == 0 {
			// Handle the case where no chatroom IDs were provided
			errorhandler.SendErrorResponse(w, http.StatusBadRequest, "No chatroom IDs provided.")
			return
		}

		// Define the table name
		tableName := "chatroom"

		// Create the placeholders for the DELETE query
		placeholders := make([]string, len(chatroomIDs))
		args := make([]interface{}, len(chatroomIDs))
		for i, id := range chatroomIDs {
			placeholders[i] = "?"
			args[i] = id
		}

		// Construct the SQL query with placeholders
		condition := fmt.Sprintf("chatroom_id IN (%s)", strings.Join(placeholders, ", "))

		// Call the ExecuteDeleteWithWhere function with the prepared statement and args
		_, err := dm.ExecuteDeleteWithWhere(tableName, condition, args...)
		if err != nil {
			errorhandler.SendErrorResponse(w, http.StatusInternalServerError, "Error deleting records: "+err.Error())
			return
		}

		successResponse := jsonResponse.ResponseMessage{
			LoginStatus: true,
			Message:     jsonResponse.DeletedRoomsMessage,
		}

		// Send back a success response if everything was successful
		jsonResponse.WriteJSONResponse(w, http.StatusOK, successResponse)
	}
}

// add a new room
func addRoom(dm *mysqlDbManager.DBManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorhandler.SendErrorResponse(w, http.StatusMethodNotAllowed, jsonResponse.Methodnotallowed)
			return
		}

		var roomData roomModel.IFroomModel

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&roomData); err != nil {
			errorhandler.SendErrorResponse(w, http.StatusBadRequest, jsonResponse.Errordecoding+err.Error())
			return
		}
		defer r.Body.Close()
		// Specify the table name
		tableName := "chatroom"
		// Insert the data into the table
		_, err := dm.ExecuteInsert(tableName, &roomData)
		if err != nil {
			errorhandler.SendErrorResponse(w, http.StatusInternalServerError, jsonResponse.DBinsertError+err.Error())
			return
		}
		// send back response if everything was successful
		successResponse := jsonResponse.ResponseMessage{
			LoginStatus: true,
			Message:     jsonResponse.RoomCreated,
		}
		jsonResponse.WriteJSONResponse(w, http.StatusOK, successResponse)

	}
}

// get all rooms list
func getRooms(dm *mysqlDbManager.DBManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := "SELECT chatroom_id, chatroom_name, created_by_user_id, chatroom_details FROM chatroom"

		rows, err := dm.Query(query)
		if err != nil {
			errorhandler.SendErrorResponse(w, http.StatusInternalServerError, "Database error")
			return
		}
		defer rows.Close()

		var allRooms []roomModel.IFroomModel

		for rows.Next() {
			var roomData roomModel.IFroomModel
			if scanErr := rows.Scan(&roomData.Chatroom_id, &roomData.Chatroom_name, &roomData.Created_by_user_id, &roomData.Chatroom_details); scanErr != nil {
				errorhandler.SendErrorResponse(w, http.StatusInternalServerError, "Failed to scan data")
				return
			}
			allRooms = append(allRooms, roomData)
		}

		if jsonErr := jsonResponse.WriteJSONResponse(w, http.StatusOK, allRooms); jsonErr != nil {
			errorhandler.SendErrorResponse(w, http.StatusInternalServerError, "JSON response error")
			return
		}
	}
}
