const { createApp } = Vue




createApp({
    el: '#app',
    data() {
        return {
            message: 'Hello Vue!',
            ws: null,
            serverUrl: "ws://localhost:8080/ws",
            roomInput: null,
            rooms: [],
            user: {
                name: "",
                username: "",
                password: "",
                token: ""
            },
            users: [],
            initialReconnectDelay: 1000,
            currentReconnectDelay: 0,
            maxReconnectDelay: 16000,
            loginError: ""
        }
    },
    mounted: function () {
    },
    methods: {
        connect: function () {
            this.connectToWebsocket();
        },
        async login() {
            try {
                const result = await axios.post("http://" + location.host + '/api/login', this.user);
                if (result.data.status !== "undefined" && result.data.status == "error") {
                    this.loginError = "Login failed";
                } else {
                    this.user.token = result.data;
                    this.connectToWebsocket();
                }
            } catch (e) {
                this.loginError = "Login failed";
                console.log(e);
            }
        },
        connectToWebsocket: function () {
            if (this.user.token != "") {
                this.ws = new WebSocket(this.serverUrl + "?bearer=" + this.user.token);
            } else {
                this.ws = new WebSocket(this.serverUrl + "?name=" + this.user.name);
            }
            this.ws.addEventListener('open', (event) => { this.onWebsocketOpen(event) });
            this.ws.addEventListener('message', (event) => { this.handleNewMessage(event) });
            this.ws.addEventListener('close', (event) => { this.onWebsocketClose(event) });
        },
        onWebsocketOpen: function () {
            console.log("connected to WS!");
            this.currentReconnectDelay = 1000;
        },

        onWebsocketClose: function () {
            this.ws = null;

            setTimeout(() => {
                this.reconnectToWebsocket();
            }, this.currentReconnectDelay);

        },

        reconnectToWebsocket: function () {
            if (this.currentReconnectDelay < this.maxReconnectDelay) {
                this.currentReconnectDelay *= 2;
            }
            this.connectToWebsocket();
        },

        handleNewMessage: function (event) {
            let data = event.data;
            data = data.split(/\r?\n/);

            for (let i = 0; i < data.length; i++) {
                let msg = JSON.parse(data[i]);
                switch (msg.action) {
                    case "send-message":
                        this.handleChatMessage(msg);
                        break;
                    case "user-join":
                        this.handleUserJoined(msg);
                        break;
                    case "user-left":
                        this.handleUserLeft(msg);
                        break;
                    case "room-joined":
                        this.handleRoomJoined(msg);
                        break;
                    default:
                        break;
                }

            }
        },
        handleChatMessage: function (msg) {
            const room = this.findRoom(msg.target.id);
            if (typeof room !== "undefined") {
                room.messages.push(msg);
            }
        },
        handleUserJoined: function (msg) {
            if (!this.userExists(msg.sender)) {
                this.users.push(msg.sender);
            }
        },
        handleUserLeft: function (msg) {
            for (let i = 0; i < this.users.length; i++) {
                if (this.users[i].id == msg.sender.id) {
                    this.users.splice(i, 1);
                    return;
                }
            }
        },
        handleRoomJoined: function (msg) {
            room = msg.target;
            room.name = room.private ? msg.sender.name : room.name;
            room["messages"] = [];
            this.rooms.push(room);
        },
        sendMessage: function (room) {
            if (room.newMessage !== "") {
                this.ws.send(JSON.stringify({
                    action: 'send-message',
                    message: room.newMessage,
                    target: {
                        id: room.id,
                        name: room.name
                    }
                }));
                room.newMessage = "";
            }
        },
        findRoom: function (roomId) {
            for (let i = 0; i < this.rooms.length; i++) {
                if (this.rooms[i].id === roomId) {
                    return this.rooms[i];
                }
            }
        },
        joinRoom: function () {
            this.ws.send(JSON.stringify({ action: 'join-room', message: this.roomInput }));
            this.roomInput = "";
        },
        leaveRoom: function (room) {
            this.ws.send(JSON.stringify({ action: 'leave-room', message: room.id }));

            for (let i = 0; i < this.rooms.length; i++) {
                if (this.rooms[i].id === room.id) {
                    this.rooms.splice(i, 1);
                    break;
                }
            }
        },
        joinPrivateRoom: function (room) {
            this.ws.send(JSON.stringify({ action: 'join-room-private', message: room.id }));
        },
        userExists: function (user) {
            for (let i = 0; i < this.users.length; i++) {
                if (this.users[i].id == user.id) {
                    return true;
                }
            }
            return false;
        }
    }
}).mount('#app')
