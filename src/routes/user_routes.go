package routes

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/manishchauhan/dugguGo/models/userModel"
	"github.com/manishchauhan/dugguGo/servers/errorhandler"
	"github.com/manishchauhan/dugguGo/servers/jsonResponse"

	"github.com/manishchauhan/dugguGo/util/mysqlDbManager"
)

func RegisterUserRoutes(router *mux.Router, dm *mysqlDbManager.DBManager) {

	subrouter := router.PathPrefix("/user").Subrouter()
	subrouter.HandleFunc("/login", handleUserLogin).Methods("POST")
	subrouter.HandleFunc("/logout", handleUserLogout).Methods("POST")
	subrouter.HandleFunc("/favorite-games", fetchFavoriteGames).Methods("GET")
	subrouter.HandleFunc("/rooms", fetchRoomList).Methods("GET")
	subrouter.HandleFunc("/userlist", fetchAllUsers(dm)).Methods("GET")
}

func handleUserLogin(w http.ResponseWriter, r *http.Request) {
	// Authentication logic
	fmt.Fprintln(w, "User Login")
	return
}
func fetchAllUsers(dm *mysqlDbManager.DBManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := dm.Query("SELECT id,username,email FROM register")
		if err != nil {
			errorhandler.HandleDatabaseError(w, err)
			return
		}
		defer rows.Close()

		var registers []userModel.IFUser

		for rows.Next() {
			var register userModel.IFUser
			if scanErr := rows.Scan(&register.ID, &register.Username, &register.Email); scanErr != nil {
				errorhandler.HandleInternalError(w, scanErr)
				return
			}
			registers = append(registers, register)
		}

		if len(registers) == 0 {
			errorhandler.WriteJSONError(w, http.StatusNotFound, "No data found")
			return
		}
		if jsonErr := jsonResponse.WriteJSONResponse(w, http.StatusOK, registers); jsonErr != nil {
			// Handle the JSON response error
			errorhandler.HandleInternalError(w, jsonErr)
		}

	}
}

func handleUserLogout(w http.ResponseWriter, r *http.Request) {
	// Logout logic
	fmt.Fprintln(w, "User Logout")
}

func fetchFavoriteGames(w http.ResponseWriter, r *http.Request) {
	// Logic to fetch favorite games
	fmt.Fprintln(w, "Fetch favorite games")
}

func fetchRoomList(w http.ResponseWriter, r *http.Request) {
	// Logic to fetch room list
	fmt.Fprintln(w, "Fetch room list")
}
