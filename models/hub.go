package models

type Hub struct {
	Clients map[*Client]bool

	Broadcast chan []byte

	Register chan *Client

	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Broadcast: make(chan []byte),
		Register: make(chan *Client),
		Unregister: make(chan *Client),
		Clients: make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
		  go h.BroadcastConnection(client)
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				go h.BroadcastDisconnection(client)
				delete(h.Clients, client)
				close(client.Send)
			}
		case message := <-h.Broadcast:
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					delete(h.Clients, client)
					close(client.Send)
				}
			}
		}
	}
}

func (h *Hub) BroadcastConnection(client *Client) {
	h.Broadcast <- []byte("A user has connected")
}

func (h *Hub) BroadcastDisconnection(client *Client) {
	h.Broadcast <- []byte("A user has disconnected")
}