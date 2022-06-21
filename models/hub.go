package models

import "log"

type Hub struct {
	Rooms map[string]map[*Client]bool

	Broadcast chan Message

	Register chan *Subscription

	Unregister chan *Subscription
}

func NewHub() *Hub {
	return &Hub{
		Broadcast: make(chan Message),
		Register: make(chan *Subscription),
		Unregister: make(chan *Subscription),
		Rooms: make(map[string]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case subscription := <-h.Register:
			connections := h.Rooms[subscription.Room]
			if connections == nil {
				h.Rooms[subscription.Room] = make(map[*Client]bool)
			}
			h.Rooms[subscription.Room][subscription.Client] = true
			go h.BroadcastConnection(subscription.Room)
		case subscription := <-h.Unregister:
			connections := h.Rooms[subscription.Room]
			if connections != nil {
				if _, ok := connections[subscription.Client]; ok {
					go h.BroadcastDisconnection(subscription.Room)
					delete(connections, subscription.Client)
					close(subscription.Client.Send)
					if len(connections) == 0 {
						delete(h.Rooms, subscription.Room)
					}
				}
			}
			log.Printf("unregistered: %v", h.Rooms[subscription.Room])
		case message := <-h.Broadcast:
			log.Printf("message %v\n", message)
			connections := h.Rooms[message.Room]
			log.Printf("connections to %v: %v", message.Room, connections)
			for client := range connections {
				select {
				case client.Send <- message.Data:
				default:
					delete(connections, client)
					close(client.Send)
				}
			}
		}
	}
}

func (h *Hub) BroadcastConnection(room string) {
	msg := Message{
		Room: room,
		Data: []byte("A user has connected"),
	}
	h.Broadcast <- msg
}

func (h *Hub) BroadcastDisconnection(room string) {
	msg := Message{
		Room: room,
		Data: []byte("A user has disconnected"),
	}
	h.Broadcast <- msg
}