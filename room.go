package main

type Room struct {
	name       string
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
}

// hàm tạo ra một room mới
func NewRoom(name string) *Room {
	return &Room{
		name:       name,
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

func (room *Room) registerCLientInRoom(client *Client) {
	room.notifyClientJoin(client)
	room.clients[client] = true
}

func (room *Room) notifyClientJoin(client *Client) {

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
