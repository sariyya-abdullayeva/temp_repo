const { createApp } = Vue
createApp({
  data() {
    return {
      ws: null,
      serverUrl: "ws://localhost:8080/ws",
      roomInput: null,
      rooms: [],
      user: {
        name: ""
      },
      users: []
    }
  },
  // mounted() {
  //   this.connectToWebsocket();
  // },
  methods: {

    connect() {
      this.connectToWebsocket();
    },
    connectToWebsocket() {
      this.ws = new WebSocket(this.serverUrl + "?name=" + this.user.name);
      this.ws.addEventListener('open', (event) => { this.onWebsocketOpen(event) });
      this.ws.addEventListener('message', (event) => { this.handleNewMessage(event) });
    },

    onWebsocketOpen() {
      console.log("connected to WS!");
      console.log(this.ws !== null);
    },

    handleNewMessage(event) {
      let data = event.data;
      console.log('message', data);
      data = data.split(/\r?\n/);

      for (let i = 0; i < data.length; i++) {
        let msg = JSON.parse(data[i]);
        console.log('message after parsing', msg);

        switch (msg.action) {
          case 'send-message':
            this.handleChatMessage(msg);
            break;
          case 'user-joined':
            this.handleUserJoined(msg);
            break;
          case 'user-left':
            this.handleUserLeft(msg);
            break;
          case 'room-joined':
            this.handleRoomJoined(msg);
            break;
          default:
            break;
        }
      }
    },
    handleRoomJoined(msg) {
      room = msg.target;
      room.name = room.private ? msg.sender.name : room.name;
      room['messages'] = [];
      this.rooms.push(room);
      console.log('rooms', this.rooms);
    },
    handleChatMessage(msg) {
      const room = this.findRoom(msg.target.id);
      if (typeof room !== "undefined") {
        room.messages.push(msg);
      }
    },

    // when a user joins a room
    handleUserJoined(msg) {
      this.users.push(msg.sender);
      console.log('users', this.users);
    },
    // when a user leaves a room
    handleUserLeft(msg) {
      for (let i = 0; i < this.users.length; i++) {
        if (this.users[i] === msg.sender.id) {
          this.users.splice(i, 1);
          return;
        }
      }
    }
    ,
    sendMessage(room) {
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
    findRoom(roomId) {
      for (let i = 0; i < this.rooms.length; i++) {
        if (this.rooms[i].id === roomId) {
          return this.rooms[i];
        }
      }
    },

    joinRoom() {
      this.ws.send(JSON.stringify({ action: 'join-room', message: this.roomInput }));
      this.roomInput = "";
    },

    joinPrivateRoom(room) {
      console.log('joindedprivateroom: room id', room.id);
      console.log('joindedprivateroom: room', room);
      this.ws.send(JSON.stringify({ action: 'join-room-private', message: room.id }));
    },
    leaveRoom(room) {
      this.ws.send(JSON.stringify({ action: 'leave-room', message: room.name }));

      for (let i = 0; i < this.rooms.length; i++) {
        if (this.rooms[i].name === room.name) {
          this.rooms.splice(i, 1);
          break;
        }
      }
    },


  }
}).mount('#app')