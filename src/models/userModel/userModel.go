package userModel

type IFUser struct {
	Userid   int    `json:"userid"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
type IFScore struct {
	Game_user string `json:"game_user"`
	Coins     int    `json:"coins"`
	Distance  int    `json:"distance"`
	Legend    bool   `json:"legend"`
}
