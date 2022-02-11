var data = {
    ws: null,
    serverUrl: "ws://localhost:8080/ws?name=anhnt650",
    roomInput: null,// sử dụng để join các room mới
    rooms: [],// theo dõi các room đã tham gia
    user: { // dành cho user data, như name
        name: ""
    },
    users: []
};

$(".msg_send_btn").on("click", () => {
    var msg = $("#write_msg").val();

    // ws send
    if (data.newMessage !== "") {
        data.ws.send(JSON.stringify({
            action: 'send-message', message: msg, target: {
                id: '123456', name: 'test'
            }
        }));
    }
    $(".msg_history").append(`<div class="outgoing_msg">
            <div class="sent_msg">
                <p>${msg}</p>
                <span class="time_date">${new Date().toLocaleString()}</span> 
            </div>
        </div>`)
});

// handle socket message

function handleNewMessage(event) {
    // data = data.split(/\r?\n/);
    console.log(event)
    var dat = JSON.parse(event.data)
    console.log(`received: ${dat.message}`)
    $(".msg_history").append(`<div class="incoming_msg">
        <div class="incoming_msg_img"> <img src="https://ptetutorials.com/images/user-profile.png" alt="sunil"> </div>
            <div class="received_msg">
                <div class="received_withd_msg">
                    <p>${dat.message}</p>
                    <span class="time_date">${new Date().toLocaleString()}</span>
                </div>    
            </div>
        </div>`)
}

(function connectToWebsocket() {
    data.ws = new WebSocket(data.serverUrl);
    data.ws.addEventListener('open', (event) => {
        console.log("connected to WS!");
        data.ws.send(JSON.stringify({action: 'join-room', message: 'hello', target: {
                id: '123456', name: 'test'
            }}));
        //data.ws.send(JSON.stringify({ action: 'join-room-private', message: "123456" }));
    });

    data.ws.addEventListener('message', (event) => {
        handleNewMessage(event)
    });

    data.ws.addEventListener('error', function (event) {
        console.log('WebSocket error: ', event);
    });
})();
