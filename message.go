package main

import (
	json2 "encoding/json"
	"log"
)

const SendMessageAction = "send-message"
const JoinRoomAction = "join-room"
const LeaveRoomAction = "leave-room"

type Message struct {
	Action  string  `json:"action"`
	Message string  `json:"message"`
	Target  string  `json:"target"`
	Sender  *Client `json:"sender"`
}

// {"action": "asdasd", "message": "asdasd",  "sender": {}}

//phương thức mã hóa có thể được gọi để tạo một đối tượng byte json [] đã sẵn sàng để gửi lại cho các máy khách
func (message *Message) encode() []byte {
	json, err := json2.Marshal(message)
	if err != nil {
		log.Println(err)
	}
	return json
}
