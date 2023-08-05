package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/manishchauhan/dugguGo/util/tree"
)

// Handler for the home route.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the Home Page!")
}

// Handler for the about route.
func aboutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "About Page")
}

// Handler for the contact route.
func contactHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Contact Page")
}
func resolveTree() {

	t := tree.Tree{}

	t.Insert(50)
	t.Insert(30)
	t.Insert(20)
	t.Insert(40)
	t.Insert(70)
	t.Insert(60)
	t.Insert(80)

	t.InOrderTraversal()
}

func main() {

	router := mux.NewRouter()

	// Define routes and their respective handlers.
	router.HandleFunc("/", homeHandler).Methods("GET")
	router.HandleFunc("/about", aboutHandler).Methods("GET")
	router.HandleFunc("/contact", contactHandler).Methods("GET")

	// Create a server and set the router as its handler.
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Start the server.
	log.Println("Server listening on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
