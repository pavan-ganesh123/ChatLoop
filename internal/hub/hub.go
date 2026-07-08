package hub

import (
	"encoding/json"
	"log"
)

type Outgoing struct {
	Type             string   `json:"type"`
	MessageID        string   `json:"messageId,omitempty"`
	From             string   `json:"from,omitempty"`
	FromUserID       string   `json:"fromUserId,omitempty"`
	To               string   `json:"to,omitempty"`
	Message          string   `json:"message,omitempty"`
	PostID             string   `json:"postId,omitempty"`
	PostTitle          string   `json:"postTitle,omitempty"`
	PostImage          string   `json:"postImage,omitempty"`
	ReplyToMessageID string   `json:"replyToMessageId,omitempty"`
	CreatedAt        string   `json:"createdAt,omitempty"`
	Users            []string `json:"users,omitempty"`
	DeletedForEveryone bool `json:"deletedForEveryone,omitempty"`
}

type privateMsg struct {
	to   string
	from string
	out  Outgoing
}

type Hub struct {
	Clients      map[*Client]bool
	NameToClient map[string][]*Client
	Register     chan *Client
	Unregister   chan *Client
	Broadcast    chan Outgoing
	Private      chan privateMsg
	Delete chan privateMsg
}

func NewHub() *Hub {

	return &Hub{
		Clients:      make(map[*Client]bool),
		NameToClient: make(map[string][]*Client),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		Broadcast:    make(chan Outgoing),
		Private:      make(chan privateMsg),
		Delete: make(chan privateMsg),
	}
}

func (h *Hub) Run() {

	for {

		select {

		// =====================================================
		// REGISTER
		// =====================================================

		case client := <-h.Register:


			h.Clients[client] = true

			h.NameToClient[client.UserID] =
				append(
					h.NameToClient[client.UserID],
					client,
				)


			joinMsg := Outgoing{
				Type:    "info",
				Message: client.Email + " joined",
			}

			data, _ := json.Marshal(joinMsg)

			for c := range h.Clients {

				select {

				case c.Send <- data:

				default:

					log.Println(
						"⚠️ Failed sending join message",
					)

					close(c.Send)

					delete(h.Clients, c)
				}
			}

		// =====================================================
		// UNREGISTER
		// =====================================================

		case client := <-h.Unregister:


			if _, ok := h.Clients[client]; ok {

				delete(h.Clients, client)

				clients :=
					h.NameToClient[client.UserID]

				newList := []*Client{}

				for _, c := range clients {

					if c != client {
						newList = append(newList, c)
					}
				}

				if len(newList) == 0 {

					delete(
						h.NameToClient,
						client.UserID,
					)

				} else {

					h.NameToClient[client.UserID] =
						newList
				}

				close(client.Send)

			}

		// =====================================================
		// BROADCAST
		// =====================================================

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

		// =====================================================
		// PRIVATE MESSAGE
		// =====================================================

		case pm := <-h.Private:

			data, _ := json.Marshal(pm.out)

			// =========================================
			// SEND TO RECEIVER
			// =========================================

			receiverClients, ok :=
				h.NameToClient[pm.to]

			if ok {

				for _, c := range receiverClients {

					select {

					case c.Send <- data:


					default:

						log.Println(
							"⚠️ Failed receiver delivery",
						)
					}
				}
			}

			// =========================================
			// SEND TO SENDER
			// =========================================

			senderClients, ok :=
				h.NameToClient[pm.from]

			if ok {

				for _, c := range senderClients {

					select {

					case c.Send <- data:


					default:

						log.Println(
							"⚠️ Failed sender delivery",
						)
					}
				}
			}
		case dm := <-h.Delete:

			data, _ := json.Marshal(dm.out)

			receiverClients, ok :=
				h.NameToClient[dm.to]

			if ok {

				for _, c := range receiverClients {

					select {

					case c.Send <- data:

					default:

					}
				}
			}

			senderClients, ok :=
				h.NameToClient[dm.from]

			if ok {

				for _, c := range senderClients {

					select {

					case c.Send <- data:

					default:

					}
				}
			}
		}
		
	}
}

// =====================================================
// DEBUG HELPER
// =====================================================

func getKeys(
	m map[string][]*Client,
) []string {

	keys := []string{}

	for k := range m {
		keys = append(keys, k)
	}

	return keys
}