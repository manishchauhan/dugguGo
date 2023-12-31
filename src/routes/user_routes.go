package routes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/manishchauhan/dugguGo/models/userModel"
	"github.com/manishchauhan/dugguGo/servers/errorhandler"
	"github.com/manishchauhan/dugguGo/servers/jsonResponse"

	"github.com/manishchauhan/dugguGo/util/auth/jwtAuth"
	"github.com/manishchauhan/dugguGo/util/mysqlDbManager"
	"github.com/manishchauhan/dugguGo/util/passwordutils"
)

func RegisterUserRoutes(router *mux.Router, dm *mysqlDbManager.DBManager) {

	subrouter := router.PathPrefix("/user").Subrouter()
	subrouter.HandleFunc("/login", handleUserLogin(dm)).Methods("POST")
	subrouter.HandleFunc("/logout", handleUserLogout).Methods("GET")
	subrouter.HandleFunc("/favorite-games", fetchFavoriteGames).Methods("GET")
	subrouter.HandleFunc("/rooms", fetchRoomList).Methods("GET")
	//subrouter.HandleFunc("/userlist", fetchAllUsers(dm)).Methods("GET")
	subrouter.HandleFunc("/register", registerUser(dm)).Methods("POST")
	subrouter.HandleFunc("/addnewscore", setHighScore(dm)).Methods("POST")
	subrouter.HandleFunc("/fetchscores", getScores(dm)).Methods("GET")
	subrouter.HandleFunc("/userexists", userExists(dm)).Methods("GET")
	subrouter.Handle("/home", jwtAuth.AuthMiddleware(http.HandlerFunc(getHome(dm)))).Methods("GET")
	subrouter.Handle("/home/chat", jwtAuth.AuthMiddleware(http.HandlerFunc(getChat(dm)))).Methods("GET")
}
func getChat(dm *mysqlDbManager.DBManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var User userModel.IFUser
		User.Userid = r.Context().Value("userid").(int)
		User.Username = r.Context().Value("username").(string)
		User.Email = r.Context().Value("email").(string)
		//Everything is great create a websocket connection with clinet================
		successResponse := jsonResponse.ResponseMessage{
			LoginStatus: true,
			Message:     "You have successfully got game payload",
			User:        User,
		}
		jsonResponse.WriteJSONResponse(w, http.StatusOK, successResponse)
	}
}
func userExists(dm *mysqlDbManager.DBManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract the user ID from the request, assuming it's provided in the query parameters
		userID := r.URL.Query().Get("game_user")
		if userID == "" {
			errorhandler.SendErrorResponse(w, http.StatusBadRequest, "User ID is required")
			return
		}

		// Check if the user already exists in the database
		checkUserQuery := fmt.Sprintf("SELECT COUNT(*) FROM scores WHERE game_user = ?")
		var userCount int

		// Execute the query and scan the result into userCount
		row, err := dm.QueryRow(checkUserQuery, userID)
		if err != nil {
			errorhandler.SendErrorResponse(w, http.StatusBadRequest, "Error querying the database"+err.Error())
			return
		}
		err = row.Scan(&userCount)
		if err != nil {
			errorhandler.SendErrorResponse(w, http.StatusInternalServerError, "Database error")
			return
		}

		if userCount > 0 {
			errorhandler.SendErrorResponse(w, http.StatusConflict, "User already exists")
			return
		}
	}
}
func getHome(dm *mysqlDbManager.DBManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var User userModel.IFUser
		User.Userid = r.Context().Value("userid").(int)
		User.Username = r.Context().Value("username").(string)
		User.Email = r.Context().Value("email").(string)
		successResponse := jsonResponse.ResponseMessage{
			LoginStatus: true,
			Message:     "You have successfully got game payload",
			User:        User,
		}
		jsonResponse.WriteJSONResponse(w, http.StatusOK, successResponse)
	}
}
func getScores(dm *mysqlDbManager.DBManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		query := fmt.Sprintf("SELECT game_user, coins, distance, legend FROM scores")

		rows, err := dm.Query(query)
		if err != nil {
			errorhandler.SendErrorResponse(w, http.StatusMethodNotAllowed, jsonResponse.Methodnotallowed)
			return
		}
		defer rows.Close()

		var allScores []userModel.IFScore

		for rows.Next() {
			var newScore userModel.IFScore
			if scanErr := rows.Scan(&newScore.Game_user, &newScore.Coins, &newScore.Distance, &newScore.Legend); scanErr != nil {
				errorhandler.SendErrorResponse(w, http.StatusMethodNotAllowed, "Something went wrong")
				return
			}
			allScores = append(allScores, newScore)
		}

		if len(allScores) == 0 {
			errorhandler.SendErrorResponse(w, http.StatusMethodNotAllowed, "Something went wrong")
			return
		}

		if jsonErr := jsonResponse.WriteJSONResponse(w, http.StatusOK, allScores); jsonErr != nil {
			errorhandler.SendErrorResponse(w, http.StatusMethodNotAllowed, "Json response is bad")
			return
		}
	}
}

// set high score
func setHighScore(dm *mysqlDbManager.DBManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorhandler.SendErrorResponse(w, http.StatusLengthRequired, jsonResponse.Methodnotallowed)
			return
		}

		var scoreDetails userModel.IFScore
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&scoreDetails); err != nil {
			errorhandler.SendErrorResponse(w, http.StatusBadRequest, jsonResponse.Errordecoding+err.Error())
			return
		}
		defer r.Body.Close()
		// Specify the table name
		tableName := "scores"
		// Insert the data into the table
		uniqueKeyColumns := []string{"game_user"}
		_, err := dm.ExecuteInsertOrUpdate(tableName, &scoreDetails, uniqueKeyColumns)
		if err != nil {
			errorhandler.SendErrorResponse(w, http.StatusInternalServerError, jsonResponse.DBinsertError+err.Error())
			return
		}
		// send back response if everything was successful
		jsonResponse.SendJSONResponse(w, http.StatusOK, jsonResponse.UserRegistered)
	}
}

// register user
func registerUser(dm *mysqlDbManager.DBManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorhandler.SendErrorResponse(w, http.StatusMethodNotAllowed, jsonResponse.Methodnotallowed)
			return
		}

		var user userModel.IFUser

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&user); err != nil {
			errorhandler.SendErrorResponse(w, http.StatusBadRequest, jsonResponse.Errordecoding+err.Error())
			return
		}
		defer r.Body.Close()
		// Specify the table name
		tableName := "user"
		// Insert the data into the table
		hashedPassword, hashError := passwordutils.HashPassword(user.Password)
		if hashError != nil {
			errorhandler.SendErrorResponse(w, http.StatusBadRequest, "Failed to hash password:"+hashError.Error())
			return
		}
		user.Password = string(hashedPassword)
		fmt.Println(user)
		_, err := dm.ExecuteInsert(tableName, &user)
		if err != nil {
			errorhandler.SendErrorResponse(w, http.StatusInternalServerError, jsonResponse.DBinsertError+err.Error())
			return
		}
		// send back response if everything was successful
		jsonResponse.SendJSONResponse(w, http.StatusOK, jsonResponse.UserRegistered)
	}

}
func handleUserLogin(dm *mysqlDbManager.DBManager) http.HandlerFunc {
	// Authentication logic
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorhandler.SendErrorResponse(w, http.StatusMethodNotAllowed, jsonResponse.Methodnotallowed)
			return
		}
		var user userModel.IFUser

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&user); err != nil {
			errorhandler.SendErrorResponse(w, http.StatusBadRequest, jsonResponse.Errordecoding+err.Error())
			return
		}
		defer r.Body.Close()
		//check password
		// Prepare and execute the SQL query using QueryRow from DBManager
		query := "SELECT * FROM user WHERE username = ?"

		row, err := dm.QueryRow(query, user.Username)

		if err != nil {
			errorhandler.SendErrorResponse(w, http.StatusBadRequest, "Error querying the database"+err.Error())
			return
		}

		var foundUser userModel.IFUser
		err = row.Scan(&foundUser.Userid, &foundUser.Username, &foundUser.Password, &foundUser.Email)
		if err == sql.ErrNoRows {
			errorhandler.SendErrorResponse(w, http.StatusUnauthorized, "User not found")
			return
		} else if err != nil {
			errorhandler.SendErrorResponse(w, http.StatusInternalServerError, "Error scanning database row")
			return
		}

		// Verify the password using bcrypt
		err = passwordutils.ComparePasswords([]byte(foundUser.Password), user.Password)
		if err != nil {

			errorhandler.SendErrorResponse(w, http.StatusInternalServerError, "Invalid password")
			return
		}
		// Everything is going good now create a jwt token and validate
		accessToken, err := jwtAuth.CreateAccessToken(foundUser.Userid, foundUser.Username, foundUser.Email)
		if err != nil {
			http.Error(w, "Error creating access token", http.StatusInternalServerError)
			errorhandler.SendErrorResponse(w, http.StatusInternalServerError, "Error creating access token")
			return
		}

		refreshToken, err := jwtAuth.CreateRefreshToken(foundUser.Userid, foundUser.Username, foundUser.Email)
		if err != nil {
			errorhandler.SendErrorResponse(w, http.StatusInternalServerError, "Error creating access token")
			return
		}
		jwtAuth.SetCookies(w, accessToken, refreshToken)
		//we need to use https for better security
		var jsonUserObject userModel.IFUser
		jsonUserObject.Email = foundUser.Email
		jsonUserObject.Userid = foundUser.Userid
		jsonUserObject.Username = foundUser.Username
		jsonResponse.WriteJSONResponse(w, http.StatusOK, jsonUserObject)
	}
}

/*
func fetchAllUsers(dm *mysqlDbManager.DBManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := dm.Query("SELECT id,username,email FROM register")
		if err != nil {

			return
		}
		defer rows.Close()

		var registers []userModel.IFUser

		for rows.Next() {
			var register userModel.IFUser
			if scanErr := rows.Scan(&register.ID, &register.Username, &register.Email); scanErr != nil {

				return
			}
			registers = append(registers, register)
		}

		if len(registers) == 0 {
			errorhandler.SendErrorResponse(w, http.StatusNotFound, "No data found")
			return
		}
		if jsonErr := jsonResponse.WriteJSONResponse(w, http.StatusOK, registers); jsonErr != nil {
			// Handle the JSON response error

		}

	}
}*/

func handleUserLogout(w http.ResponseWriter, r *http.Request) {
	// Logout logic
	jwtAuth.ClearCookies(w)
	jsonResponse.SendJSONResponse(w, http.StatusOK, "You have been logged out")
}

func fetchFavoriteGames(w http.ResponseWriter, r *http.Request) {
	// Logic to fetch favorite games
	fmt.Fprintln(w, "Fetch favorite games")
}

func fetchRoomList(w http.ResponseWriter, r *http.Request) {
	// Logic to fetch room list
	fmt.Fprintln(w, "Fetch room list")
}
