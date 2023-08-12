package routes

import (
	"github.com/gorilla/mux"
	"github.com/manishchauhan/dugguGo/util/mysqlDbManager"
)

func RegisterRoutes(r *mux.Router, db *mysqlDbManager.DBManager) {
	RegisterUserRoutes(r, db)
	RegisterAdminRoutes(r, db)
}
