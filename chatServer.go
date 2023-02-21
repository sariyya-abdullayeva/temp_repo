package main

import "fmt"

const PubSubGeneralChannel = "general"

type WsServer struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	rooms      map[*Room]bool
	broadcast  chan []byte
}

// NewWebsocketServer creates a new WsServer type
func NewWebsocketServer() *WsServer {
	fmt.Println("NewWebsocketServer")
	wsServer := &WsServer{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		rooms:      make(map[*Room]bool),
		broadcast:  make(chan []byte),
	}

	return wsServer
}

// Run our websocket server, accepting various requests
func (server *WsServer) Run() {
	fmt.Println("Run")
	for {
		select {
		case client := <-server.register:
			fmt.Println("chatServer.go: Run: case client := <-server.register:")
			server.registerClient(client)

		case client := <-server.unregister:
			fmt.Println("chatServer.go: Run: case client := <-server.unregister:")
			server.unregisterClient(client)

		case message := <-server.broadcast:
			fmt.Println("chatServer.go: Run: case message := <-server.broadcast:")
			server.broadcastToClients(message)
		}

	}
}

func (server *WsServer) registerClient(client *Client) {
	fmt.Println("registerClient")
	//notify all clients that a new user has joined
	server.notifyClientJoined(client)
	//list all online clients to the new client
	server.listOnlineClients(client)
	server.clients[client] = true
}

func (server *WsServer) unregisterClient(client *Client) {
	fmt.Println("unregisterClient")
	if _, ok := server.clients[client]; ok {
		//notify all clients that a user has left
		server.notifyClientLeft(client)
		delete(server.clients, client)
	}
}

func (server *WsServer) broadcastToClients(message []byte) {
	fmt.Println("broadcastToClients")
	for client := range server.clients {
		client.send <- message
	}
}

func (server *WsServer) findRoomByName(name string) *Room {
	fmt.Println("findRoomByName")
	var foundRoom *Room
	for room := range server.rooms {
		if room.GetName() == name {
			foundRoom = room
			break
		}
	}

	return foundRoom
}

func (server *WsServer) findRoomByID(ID string) *Room {
	var foundRoom *Room
	for room := range server.rooms {
		if room.GetId() == ID {
			foundRoom = room
			break
		}
	}

	return foundRoom
}

func (server *WsServer) createRoom(name string, private bool) *Room {
	fmt.Println("createRoom")
	room := NewRoom(name, private)
	go room.RunRoom()
	server.rooms[room] = true

	return room
}

func (server *WsServer) findClientByID(ID string) *Client {
	fmt.Println("findClientByID")
	var foundClient *Client
	for client := range server.clients {
		if client.ID.String() == ID {
			foundClient = client
			break
		}
	}

	return foundClient
}

func (server *WsServer) notifyClientJoined(client *Client) {
	fmt.Println("notifyClientJoined")
	message := &Message{
		Action: UserJoinedAction,
		Sender: client,
	}

	server.broadcastToClients(message.encode())
}

func (server *WsServer) notifyClientLeft(client *Client) {
	fmt.Println("notifyClientLeft")
	message := &Message{
		Action: UserLeftAction,
		Sender: client,
	}

	server.broadcastToClients(message.encode())
}

func (server *WsServer) listOnlineClients(client *Client) {
	fmt.Println("listOnlineClients")
	for existingClient := range server.clients {
		message := &Message{
			Action: UserJoinedAction,
			Sender: existingClient,
		}
		client.send <- message.encode()
	}
}
