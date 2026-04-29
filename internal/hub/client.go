package hub

import (
	"github.com/gorilla/websocket"
	"encoding/json"
)

type Client struct {
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan []byte
	UserID string
	Email  string
}

type Incoming struct {
	Type    string `json:"type"`             // "broadcast" | "private" | "list"
	Message string `json:"message,omitempty"`
	To      string `json:"to,omitempty"`
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var incoming Incoming
		if err := json.Unmarshal(msg, &incoming); err != nil {
			continue // ignore bad messages
		}

		switch incoming.Type {

		case "broadcast":
			c.Hub.Broadcast <- Outgoing{
				Type:    "broadcast",
				From:    c.Email,
				Message: incoming.Message,
			}

		case "private":
			if incoming.To == "" {
				continue
			}

			c.Hub.Private <- privateMsg{
				to: incoming.To,
				out: Outgoing{
					Type:    "private",
					From:    c.Email,
					Message: incoming.Message,
				},
			}

		case "list":
			// send user list only to this client
			users := make([]string, 0)

			for name := range c.Hub.NameToClient {
				users = append(users, name)
			}

			out := Outgoing{
				Type:  "userlist",
				Users: users,
			}

			data, _ := json.Marshal(out)
			c.Send <- data

		default:
			// ignore unknown types
		}
	}
}

func (c *Client) WritePump() {
	defer c.Conn.Close()

	for {
		msg, ok := <-c.Send
		if !ok {
			// channel closed
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		err := c.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}