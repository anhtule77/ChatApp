package main

import "fmt"

const welcomeMessage = "%s đã tham gia phòng"

type Room struct {
	Name       string `json:"name"`
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
}

// hàm tạo ra một room mới
func NewRoom(name string) *Room {
	return &Room{
		Name:       name,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message),
	}
}

// hàm Runroom chạy các room, chấp nhận các yêu cầu khác nhau
func (room *Room) RunRoom() {
	for {
		select {
		case client := <-room.register:
			room.registerCLientInRoom(client)
		case client := <-room.unregister:
			room.unregisterClientInRoom(client)
		case message := <-room.broadcast:
			room.broadcastToClientsInRoom(message.encode())
		}
	}
}

//gọi method notify khi người dùng registe
func (room *Room) registerCLientInRoom(client *Client) {
	// bằng cách gửi tin nhắn trước,người dùng mới sẽ kh thấy tn của chính họ
	room.notifyClientJoin(client)
	room.clients[client] = true
}

//gửi thông báo người dùng mới tham gia, để những người trong room biết
func (room *Room) notifyClientJoin(client *Client) {
	message := &Message{
		Action:  SendMessageAction,
		Target:  room.Name,
		Message: fmt.Sprintf(welcomeMessage, client.GetName()),
	}
	room.broadcastToClientsInRoom(message.encode())
}

func (room *Room) unregisterClientInRoom(client *Client) {
	if _, ok := room.clients[client]; ok {
		delete(room.clients, client)
	}
}

func (room *Room) broadcastToClientsInRoom(message []byte) {
	for client := range room.clients {
		client.send <- message
	}
}

func (room *Room) GetName() string {
	return room.Name
}
