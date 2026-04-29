package hub

import "encoding/json"

type Outgoing struct {
	Type    string   `json:"type"`
	From    string   `json:"from,omitempty"`
	Message string   `json:"message,omitempty"`
	To      string   `json:"to,omitempty"`
	Users   []string `json:"users,omitempty"`
}

type privateMsg struct {
	to  string
	out Outgoing
}

type Hub struct {
	Clients      map[*Client]bool
	NameToClient map[string][]*Client
	Register     chan *Client
	Unregister   chan *Client
	Broadcast    chan Outgoing
	Private      chan privateMsg
}

// constructor
func NewHub() *Hub {
	return &Hub{
		Clients:      make(map[*Client]bool),
		NameToClient: make(map[string][]*Client),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		Broadcast:    make(chan Outgoing),
		Private:      make(chan privateMsg),
	}
}

func (h *Hub) Run() {
	for {
		select {

		// ✅ Register new client
		case client := <-h.Register:
			h.Clients[client] = true
			h.NameToClient[client.Email] = append(h.NameToClient[client.Email], client)

			msg := Outgoing{
				Type:    "info",
				Message: client.Email + " joined",
			}
			data, _ := json.Marshal(msg)

			for c := range h.Clients {
				select {
				case c.Send <- data:
				default:
					close(c.Send)
					delete(h.Clients, c)
				}
			}

		// ❌ Remove client
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)

				clients := h.NameToClient[client.Email]
				newList := []*Client{}

				for _, c := range clients {
					if c != client {
						newList = append(newList, c)
					}
				}

				if len(newList) == 0 {
					delete(h.NameToClient, client.Email)
				} else {
					h.NameToClient[client.Email] = newList
				}

				close(client.Send)
			}
		// 📢 Broadcast message
		case msg := <-h.Broadcast:
			data, _ := json.Marshal(msg)

			for client := range h.Clients {
				select {
				case client.Send <- data:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		// 🔒 Private message
		case pm := <-h.Private:
			if clients, ok := h.NameToClient[pm.to]; ok {
				for _, c := range clients {
					data, _ := json.Marshal(pm.out)
					c.Send <- data
				}
			}
		}
	}
}