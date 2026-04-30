package hub

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
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
		log.Println("❌ Disconnecting:", c.UserID, c.Email)
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	log.Println("✅ ReadPump started for:", c.UserID, c.Email)

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("❌ Read error:", c.UserID, err)
			break
		}

		log.Println("📩 Raw message from", c.UserID, ":", string(msg))

		var incoming Incoming
		if err := json.Unmarshal(msg, &incoming); err != nil {
			log.Println("❌ JSON parse error:", err)
			continue
		}

		log.Println("📨 Parsed message:", incoming.Type, "From:", c.UserID)

		switch incoming.Type {

		case "broadcast":
			log.Println("📢 Broadcast from:", c.UserID)

			c.Hub.Broadcast <- Outgoing{
				Type:    "broadcast",
				From:    c.Email,
				Message: incoming.Message,
			}

		case "private":
			if incoming.To == "" {
				log.Println("⚠️ Private message missing receiver")
				continue
			}

			log.Println("📤 Private message from", c.UserID, "to", incoming.To)

			c.Hub.Private <- privateMsg{
				to: incoming.To,
				out: Outgoing{
					Type:    "private",
					From:    c.Email,
					Message: incoming.Message,
				},
			}

		case "list":
			log.Println("📋 User list requested by:", c.UserID)

			users := make([]string, 0)

			for name := range c.Hub.NameToClient {
				users = append(users, name)
			}

			out := Outgoing{
				Type:  "userlist",
				Users: users,
			}

			data, _ := json.Marshal(out)

			log.Println("📤 Sending user list to:", c.UserID, users)

			c.Send <- data

		default:
			log.Println("⚠️ Unknown message type:", incoming.Type)
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		log.Println("❌ WritePump closing for:", c.UserID)
		c.Conn.Close()
	}()

	log.Println("✅ WritePump started for:", c.UserID)

	for {
		msg, ok := <-c.Send
		if !ok {
			log.Println("❌ Send channel closed for:", c.UserID)
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		log.Println("📤 Sending message to", c.UserID, ":", string(msg))

		err := c.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("❌ Write error for", c.UserID, ":", err)
			return
		}
	}
}