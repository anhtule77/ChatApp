package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// thời gian tối đa khi viết mess
	writeWait = 10 * time.Second

	// thời gian tối đa pong from peer
	pongWait = 60 * time.Second

	//khoảng thời gian gửi ping, phải ít hơn thời gian chờ pong
	pingPeriod = (pongWait * 9) / 10

	// thời gian lớn nhất để gửi mess.
	maxMessageSize = 10000
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

// Client represents the websocket client at the server
type Client struct {
	// The actual websocket connection.
	conn     *websocket.Conn
	wsServer *WsServer
	send     chan []byte
	rooms    map[*Room]bool
	Name     string `json:"name"` // dùng để đặt tên cho client
}

type Target struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func newClient(conn *websocket.Conn, wsServer *WsServer, name string) *Client {
	return &Client{
		conn:     conn,
		wsServer: wsServer,
		send:     make(chan []byte, 256),
		rooms:    make(map[*Room]bool),
		Name:     name,
	}
}

func (client *Client) readPump() {
	defer func() {
		client.disconnect()
	}()

	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error { client.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// Start endless read loop, waiting for messages from client
	for {
		_, jsonMessage, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected close error: %v", err)
			}
			break
		}
		/*client.wsServer.broadcast <- jsonMessage*/
		//sử dụng các method mới
		client.handlerNewMessage(jsonMessage)
	}

}

func (client *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()
	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				//WsServer đã đóng chanel.
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Đính kèm tin nhắn trò chuyện đã xếp hàng đợi vào tin nhắn websocket hiện tại.
			n := len(client.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-client.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (client *Client) disconnect() {
	client.wsServer.unregister <- client
	close(client.send)
	client.conn.Close()
}

//ServeWs xử lý các yêu cầu websocket từ các yêu cầu của khách hàng.
func ServeWs(wsServer *WsServer, w http.ResponseWriter, r *http.Request) {

	name, ok := r.URL.Query()["name"]
	//check Query name
	if !ok || len(name[0]) < 1 {
		log.Printf("URL Param bị thiếu")
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := newClient(conn, wsServer, name[0])

	go client.writePump()
	go client.readPump()

	wsServer.register <- client
}

func (client *Client) disconect() {
	client.wsServer.unregister <- client
	for room := range client.rooms {
		room.unregister <- client
	}
}

//method giải mã Json message sau đó xử lý trực tiếp hoặc
// chuyển nó đến trình xử lý

func (client *Client) handlerNewMessage(jsonMessage []byte) {
	var message Message

	//Unmarshal phân tích cú pháp dữ liệu được mã hóa JSON và lưu trữ kết quả trong giá trị được trỏ tới bởi v.
	//Nếu v là nil hoặc không phải là một con trỏ,Unmarshal trả về lỗi InvalidUnmarshalError
	if err := json.Unmarshal(jsonMessage, &message); err != nil {
		log.Printf("Json error do not manager %s", err)
		log.Printf("%s", string(jsonMessage))
	}

	// gán đối tượng client là người gửi tin nhắn
	message.Sender = *client

	switch message.Action {
	case SendMessageAction:
		log.Printf("%s", message.Message)
		// action này sẽ gửi tin nhắn đến một phòng
		// room nào thì sẽ tùy thuộc vào message Target
		roomId := message.Target.Id
		//sử dụng method chatServer để tìm room, nếu tìm thấy,phát tin
		if room := client.wsServer.findRoomByID(roomId); room != nil {
			room.broadcast <- &message
			log.Printf("room: %s", room.Name)
		}
	case JoinRoomAction:
		client.handleJoinRoomMessage(message)
	case LeaveRoomAction:
		client.handlerLeaveRoomMessage(message)
	}
}

// với method này, cta sẽ trực tiếp gửi mess đến 1 room,
// vì hiện tại đang gửi các đối tượng Message thay vì các đối tượng []byte

func (client *Client) handleJoinRoomMessage(message Message) {
	log.Printf("join> %s", string(message.Message))
	roomName := message.Target.Name
	roomId := message.Target.Id
	room := client.wsServer.findRoomByID(roomId)
	if room == nil {
		room = client.wsServer.createRoom(roomId, roomName)
	}
	client.rooms[room] = true
	room.register <- client
}

func (client *Client) handlerLeaveRoomMessage(message Message) {
	room := client.wsServer.findRoomByName(message.Message)
	if _, ok := client.rooms[room]; ok {
		delete(client.rooms, room)
	}
	room.unregister <- client
}

func (client *Client) GetName() string {
	return client.Name
}
