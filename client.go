package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/gorilla/websocket"
)

const (
	// Max wait time when writing message to peer
	writeWait = 10 * time.Second

	// Max time till next pong from peer
	pongWait = 60 * time.Second

	// Send ping interval, must be less then pong wait time
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 10000
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	// CHANGE LATER: This is insecure, but for now it's ok
	CheckOrigin: func(r *http.Request) bool { return true },
}

type ServiceID struct {
	UUID uuid.UUID
}

// Client represents the websocket client at the server
type Client struct {
	// The actual websocket connection.
	conn     *websocket.Conn
	wsServer *WsServer
	send     chan []byte
	// ID       uuid.UUID `json:"id"`
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	rooms map[*Room]bool
}

func newClient(conn *websocket.Conn, wsServer *WsServer, name string) *Client {
	// func newClient(conn *websocket.Conn, wsServer *WsServer, name string, ID string) *Client {
	fmt.Println("client.go newClient")
	client := &Client{
		Name:     name,
		conn:     conn,
		wsServer: wsServer,
		send:     make(chan []byte, 256),
		// ID:       uuid.NewV4(),
		ID:    uuid.NewV4(),
		rooms: make(map[*Room]bool),
	}

	return client
}

func (client *Client) readPump() {
	fmt.Println("client.go readPump")
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

		client.handleNewMessage(jsonMessage)
	}

}

func (client *Client) writePump() {
	fmt.Println("client.go writePump")
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()
	for {
		select {
		case message, ok := <-client.send:
			fmt.Println("client.go writePump client.send")
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The WsServer closed the channel.
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Attach queued chat messages to the current websocket message.
			n := len(client.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-client.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			fmt.Println("client.go writePump ticker.C")
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (client *Client) disconnect() {
	fmt.Println("client.go disconnect")
	client.wsServer.unregister <- client
	for room := range client.rooms {
		room.unregister <- client
	}
	close(client.send)
	client.conn.Close()
}

// ServeWs handles websocket requests from clients requests.
func ServeWs(wsServer *WsServer, w http.ResponseWriter, r *http.Request) {
	fmt.Println("client.go ServeWs")
	// userCtxValue := r.Context().Value(auth.UserContextKey)
	// if userCtxValue == nil {
	// 	log.Println("Not authenticated")
	// 	return
	// }

	// user := userCtxValue.(models.User)

	name, ok := r.URL.Query()["name"]

	if !ok || len(name[0]) < 1 {
		log.Println("Url Param 'name' is missing")
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

	fmt.Println("New Client joined the hub!")
	fmt.Println(client)
	wsServer.register <- client
}

func (client *Client) handleNewMessage(jsonMessage []byte) {
	fmt.Println("client.go handleNewMessage")
	var message Message
	if err := json.Unmarshal(jsonMessage, &message); err != nil {
		log.Printf("Error on unmarshal JSON message %s", err)
		return
	}

	// Attach the client object as the sender of the messsage.
	message.Sender = client

	switch message.Action {
	case SendMessageAction:
		fmt.Println("client.go handleNewMessage SendMessageAction")
		// The send-message action, this will send messages to a specific room now.
		// Which room wil depend on the message Target
		roomID := message.Target.GetId()
		// roomID := message.Target.GetId()
		if room := client.wsServer.findRoomByID(roomID); room != nil {
			room.broadcast <- &message
		}

		// We delegate the join and leave actions.
	case JoinRoomAction:
		fmt.Println("client.go handleNewMessage JoinRoomAction")
		client.handleJoinRoomMessage(message)

	case LeaveRoomAction:
		fmt.Println("client.go handleNewMessage LeaveRoomAction")
		client.handleLeaveRoomMessage(message)

	case JoinRoomPrivateAction:
		client.handleJoinRoomPrivateMessage(message)
	}

}

func (client *Client) handleJoinRoomMessage(message Message) {
	fmt.Println("client.go handleJoinRoomMessage")
	// client.joinRoom(roomName, nil)

	roomName := message.Message

	// room := client.wsServer.findRoomByName(roomName)
	// if room == nil {
	// 	room = client.wsServer.createRoom(roomName)
	// }

	// client.rooms[room] = true

	// room.register <- client
	client.joinRoom(roomName, nil)
}

func (client *Client) handleLeaveRoomMessage(message Message) {
	fmt.Println("client.go handleLeaveRoomMessage")
	room := client.wsServer.findRoomByID(message.Message)
	if room == nil {
		return
	}

	if _, ok := client.rooms[room]; ok {
		delete(client.rooms, room)
	}

	room.unregister <- client
}

func (client *Client) handleJoinRoomPrivateMessage(message Message) {
	fmt.Println("client.go handleJoinRoomPrivateMessage")

	target := client.wsServer.findClientByID(message.Message)

	if target == nil {
		return
	}

	// create unique room name combined to the two IDs
	roomName := message.Message + client.ID.String()

	client.joinRoom(roomName, target)
	target.joinRoom(roomName, client)

}

// Joining a room both for public and private roooms
// When joiing a private room a sender is passed as the opposing party

func (client *Client) joinRoom(roomName string, sender *Client) {
	fmt.Println("client.go joinRoom")

	room := client.wsServer.findRoomByName(roomName)
	if room == nil {
		room = client.wsServer.createRoom(roomName, sender != nil)
	}

	// Don't allow to join private rooms through public room message
	if sender == nil && room.Private {
		return
	}
	if !client.isInRoom(room) {
		client.rooms[room] = true
		room.register <- client
		client.notifyRoomJoined(room, sender)
	}

}

// Check if the client is not yet in the room
func (client *Client) isInRoom(room *Room) bool {
	fmt.Println("client.go isInRoom")
	if _, ok := client.rooms[room]; ok {
		return true
	}

	return false
}

func (client *Client) GetName() string {
	fmt.Println("client.go GetName")
	return client.Name
}

// Notify the client of the new room he/she joined
func (client *Client) notifyRoomJoined(room *Room, sender *Client) {
	message := Message{
		Action: RoomJoinedAction,
		Target: room,
		Sender: sender,
	}

	client.send <- message.encode()
}
