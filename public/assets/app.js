var app = new Vue({
    el: '#app',
    data: {
        ws: null,
        serverUrl: "ws://localhost:8080/ws",
        roomInput: null, // sử dụng để join các room mới
        rooms: [], // theo dõi các room đã tham gia
        user: { // dành cho user data, như name
            name:""
        }
    },
    methods: {
        connect(){
            this.connectToWebsocket();
        },
        connectToWebsocket() {
            // truyền tham số name khi connect
            this.ws = new WebSocket( this.serverUrl + "?name=" + this.user.name );
            this.ws.addEventListener('open', (event) => { this.onWebsocketOpen(event) });
            this.ws.addEventListener('message', (event) => { this.handleNewMessage(event) });
        },
        onWebsocketOpen() {
            console.log("connected to WS!");
        },
        handleNewMessage(event) {
            let data = event.data;
            data = data.split(/\r?\n/);

            for (let i = 0; i < data.length; i++) {
                let msg = JSON.parse(data[i]);
                //hiển thị thông báo vào đúng room
                const room = this.findRoom(msg.target)
                if(typeof room != "undefined"){
                    room.message.push(msg);
                }

            }
        },
        sendMessage(room) {
            // gửi tin nhắn đến đúng room
            if(room.newMessage !== "") {
                room.ws.send(JSON.stringify({
                    action: room.sendMessage(),
                    message: room.newMessage,
                    target: room.name
                }));
                room.newMessage = "";
            }
        },

        findRoom(roomName) {
            for (let i = 0;i < this.rooms.length;i++){
                if (this.rooms[i].name = roomName){
                    return this.rooms[i];
                }
            }
            return undefined;
        },

    }
})