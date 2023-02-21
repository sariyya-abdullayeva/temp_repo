package main

import (
	"encoding/json"
	"fmt"
	"log"
)

const SendMessageAction = "send-message"
const JoinRoomAction = "join-room"
const LeaveRoomAction = "leave-room"
const UserJoinedAction = "user-joined"
const UserLeftAction = "user-left"
const JoinRoomPrivateAction = "join-room-private"
const RoomJoinedAction = "room-joined"

type Message struct {
	Action  string `json:"action"`
	Message string `json:"message"`
	// Target  string  `json:"target"`
	Target *Room   `json:"target"`
	Sender *Client `json:"sender"`
}

func (message *Message) encode() []byte {
	fmt.Println("message.go encode")
	json, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}

	return json
}

func (message *Message) UnmarshalJSON(data []byte) error {
	fmt.Println("message.go UnmarshalJSON")
	type Alias Message
	msg := &struct {
		Sender Client `json:"sender"`
		*Alias
	}{
		Alias: (*Alias)(message),
	}
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}
	message.Sender = &msg.Sender
	return nil
}
