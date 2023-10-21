package routes

import (
	"github.com/gorilla/mux"
	"github.com/manishchauhan/dugguGo/util/mysqlDbManager"
)

func RegisterRoutes(r *mux.Router, dm *mysqlDbManager.DBManager) {
	RegisterUserRoutes(r, dm)
	RegisterAdminRoutes(r, dm)
	RegisterRoomsRoutes(r, dm)
}
