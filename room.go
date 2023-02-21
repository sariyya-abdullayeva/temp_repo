package main

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
)

const welcomeMessage = "%s joined the room"

type Room struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Private    bool      `json:"private"`
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
}

// NewRoom creates a new Room
func NewRoom(name string, private bool) *Room {
	fmt.Println("room.go NewRoom")
	return &Room{
		ID:         uuid.NewV4(),
		Name:       name,
		Private:    private,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message),
	}
}

// RunRoom runs our room, accepting various requests
func (room *Room) RunRoom() {
	fmt.Println("room.go RunRoom")
	for {
		select {

		case client := <-room.register:
			fmt.Println("room.go RunRoom: case client := <-room.register:")
			room.registerClientInRoom(client)

		case client := <-room.unregister:
			fmt.Println("room.go RunRoom: case client := <-room.unregister:")
			room.unregisterClientInRoom(client)

		case message := <-room.broadcast:
			fmt.Println("room.go RunRoom: case message := <-room.broadcast:")
			room.broadcastToClientsInRoom(message.encode())
		}

	}
}

func (room *Room) registerClientInRoom(client *Client) {
	fmt.Println("room.go registerClientInRoom")
	// room.notifyClientJoined(client)
	if !room.Private {
		room.notifyClientJoined(client)
	}
	room.clients[client] = true
}

func (room *Room) unregisterClientInRoom(client *Client) {
	fmt.Println("room.go unregisterClientInRoom")
	if _, ok := room.clients[client]; ok {
		delete(room.clients, client)
	}
}

func (room *Room) broadcastToClientsInRoom(message []byte) {
	fmt.Println("room.go broadcastToClientsInRoom")
	for client := range room.clients {
		client.send <- message
	}
}

func (room *Room) notifyClientJoined(client *Client) {
	fmt.Println("room.go notifyClientJoined")
	message := &Message{
		Action:  SendMessageAction,
		Target:  room,
		Message: fmt.Sprintf(welcomeMessage, client.GetName()),
	}

	room.broadcastToClientsInRoom(message.encode())
}

func (room *Room) GetName() string {
	fmt.Println("room.go GetName")
	return room.Name
}

func (room *Room) GetId() string {
	fmt.Println("room.go GetID")
	return room.ID.String()
}
