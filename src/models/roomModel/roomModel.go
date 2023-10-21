package roomModel

type IFroomModel struct {
	Chatroom_id        int    `json:"chatroom_id"`        //room_id
	Chatroom_name      string `json:"chatroom_name"`      //name
	Created_by_user_id int    `json:"created_by_user_id"` //user_id
	Chatroom_details   string `json:"chatroom_details"`   //details
}
