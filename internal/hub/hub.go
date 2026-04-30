package hub

import (
	"encoding/json"
	"log"
)

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
			log.Println("🟢 Registering client:", client.UserID, client.Email)

			h.Clients[client] = true

			h.NameToClient[client.UserID] = append(h.NameToClient[client.UserID], client)

			log.Println("📊 Total connected clients:", len(h.Clients))
			log.Println("📊 NameToClient map:", h.NameToClient)

			msg := Outgoing{
				Type:    "info",
				Message: client.Email + " joined",
			}
			data, _ := json.Marshal(msg)

			for c := range h.Clients {
				select {
				case c.Send <- data:
				default:
					log.Println("⚠️ Failed sending join message, removing client")
					close(c.Send)
					delete(h.Clients, c)
				}
			}

		// ❌ Remove client
		case client := <-h.Unregister:
			log.Println("🔴 Unregistering client:", client.UserID, client.Email)

			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)

				clients := h.NameToClient[client.UserID]
				newList := []*Client{}

				for _, c := range clients {
					if c != client {
						newList = append(newList, c)
					}
				}

				if len(newList) == 0 {
					delete(h.NameToClient,  client.UserID)
				} else {
					h.NameToClient[client.UserID] = newList
				}

				log.Println("📊 Remaining clients:", len(h.Clients))
				close(client.Send)
			}

		// 📢 Broadcast message
		case msg := <-h.Broadcast:
			log.Println("📢 Broadcasting message from:", msg.From)

			data, _ := json.Marshal(msg)

			for client := range h.Clients {
				select {
				case client.Send <- data:
				default:
					log.Println("⚠️ Failed broadcast send, removing client")
					close(client.Send)
					delete(h.Clients, client)
				}
			}

		// 🔒 Private message
		case pm := <-h.Private:
			log.Println("🔁 Routing private message to:", pm.to)

			clients, ok := h.NameToClient[pm.to]

			if !ok {
				log.Println("❌ Receiver NOT CONNECTED:", pm.to)
				log.Println("📊 Available keys:", getKeys(h.NameToClient))
				continue
			}

			log.Println("✅ Receiver FOUND:", pm.to, "Connections:", len(clients))

			data, _ := json.Marshal(pm.out)

			for _, c := range clients {
				select {
				case c.Send <- data:
					log.Println("📤 Message delivered to:", pm.to)
				default:
					log.Println("⚠️ Failed sending to one connection")
				}
			}
		}
	}
}

// 🔧 helper to debug map keys
func getKeys(m map[string][]*Client) []string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}