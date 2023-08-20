package userModel

type IFUser struct {
	UserId   int    `json:"userid"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
