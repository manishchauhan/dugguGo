package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/manishchauhan/dugguGo/models/adminModel"
	"github.com/manishchauhan/dugguGo/servers/jsonResponse"
	"github.com/manishchauhan/dugguGo/util/mysqlDbManager"
)

func RegisterAdminRoutes(router *mux.Router, dm *mysqlDbManager.DBManager) {
	subrouter := router.PathPrefix("/admin").Subrouter()
	subrouter.HandleFunc("/login", handleUserLogin).Methods("POST")
	subrouter.HandleFunc("/logout", handleUserLogout).Methods("POST")
	subrouter.HandleFunc("/adminlist", fetchAdminList(dm)).Methods("GET")
}
func handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	// Authentication logic
}

func handleAdminLogout(w http.ResponseWriter, r *http.Request) {
	// Logout logic
}
func fetchAdminList(dm *mysqlDbManager.DBManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := dm.Query("SELECT id,username,email FROM admin")
		if err != nil {
			//errorhandler.HandleDatabaseError(w, err)
			return
		}
		defer rows.Close()

		var registers []adminModel.IFAdmin

		for rows.Next() {
			var register adminModel.IFAdmin
			if scanErr := rows.Scan(&register.ID, &register.Username, &register.Email); scanErr != nil {
				//errorhandler.HandleInternalError(w, scanErr)
				return
			}
			registers = append(registers, register)
		}

		if len(registers) == 0 {
			//errorhandler.WriteJSONError(w, http.StatusNotFound, "No data found")
			return
		}
		if jsonErr := jsonResponse.WriteJSONResponse(w, http.StatusOK, registers); jsonErr != nil {
			// Handle the JSON response error
			//	errorhandler.HandleInternalError(w, jsonErr)
		}

	}
}
