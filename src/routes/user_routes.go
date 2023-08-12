package routes

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/manishchauhan/dugguGo/util/mysqlDbManager"
)

func RegisterUserRoutes(router *mux.Router, db *mysqlDbManager.DBManager) {

	subrouter := router.PathPrefix("/user").Subrouter()
	subrouter.HandleFunc("/login", handleUserLogin).Methods("POST")
	subrouter.HandleFunc("/logout", handleUserLogout).Methods("POST")
	subrouter.HandleFunc("/favorite-games", fetchFavoriteGames).Methods("GET")
	subrouter.HandleFunc("/rooms", fetchRoomList).Methods("GET")
}

func handleUserLogin(w http.ResponseWriter, r *http.Request) {
	// Authentication logic
	fmt.Fprintln(w, "User Login")
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
