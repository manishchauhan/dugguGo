// mysqlhttpserver/server.go

package mysqlhttpserver

import (
	"fmt"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/manishchauhan/dugguGo/routes"
	"github.com/manishchauhan/dugguGo/util/mysqlDbManager"
)

func StartServer(port string, dm *mysqlDbManager.DBManager) {
	rootRouter := mux.NewRouter()
	routes.RegisterRoutes(rootRouter, dm)
	// Create a CORS middleware with allowed origins, methods, and headers
	allowedOrigins := []string{"http://localhost:3000", "http://192.168.29.216:3000"}
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins(allowedOrigins), // Allow any origin
		//dhandlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type"}),
		handlers.AllowCredentials(),
	)
	http.Handle("/", corsHandler(rootRouter))
	// Add a route for the home page
	rootRouter.HandleFunc("/", homeHandler).Methods("GET")
	fmt.Println("Server is running on port:", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Println("Server error:", err)
	}
}
func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to the Home Page")
}
