package routes

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/manishchauhan/dugguGo/util/mysqlDbManager"
)

func RegisterAdminRoutes(router *mux.Router, dm *mysqlDbManager.DBManager) {
	subrouter := router.PathPrefix("/admin").Subrouter()
	subrouter.HandleFunc("/login", handleUserLogin).Methods("POST")
	subrouter.HandleFunc("/logout", handleUserLogout).Methods("POST")
	subrouter.HandleFunc("/adminlist", fetchAdminList).Methods("GET")
}
func handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	// Authentication logic
}

func handleAdminLogout(w http.ResponseWriter, r *http.Request) {
	// Logout logic
}
func fetchAdminList(w http.ResponseWriter, r *http.Request) {
	// Authentication logic
	fmt.Fprintln(w, "Admin list")
}
