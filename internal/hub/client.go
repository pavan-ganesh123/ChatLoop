package hub

import (
	"encoding/json"
	"log"
	"time"
	S "realtime-chat/internal/service"
	"github.com/google/uuid"
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
	Type             string `json:"type"`
	MessageID        string `json:"messageId,omitempty"`
	Message          string `json:"message,omitempty"`
	To               string `json:"to,omitempty"`
	ReplyToMessageID string `json:"replyToMessageId,omitempty"`
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

		var incoming Incoming

		if err := json.Unmarshal(msg, &incoming); err != nil {
			log.Println("❌ JSON parse error:", err)
			continue
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

				log.Println("⚠️ Private message missing receiver")
				continue
			}

			blocked := S.CheckIfBlocked(
				c.UserID,
				incoming.To,
			)

			if blocked {

				out := Outgoing{
					Type:    "error",
					Message: "You cannot message this user",
				}

				data, _ := json.Marshal(out)

				c.Send <- data

				continue
			}

			messageID := uuid.New().String()

			outgoing := Outgoing{
				Type:             "private",
				MessageID:        messageID,
				From:             c.Email,
				FromUserID:       c.UserID,
				To:               incoming.To,
				Message:          incoming.Message,
				ReplyToMessageID: incoming.ReplyToMessageID,
				CreatedAt:        time.Now().Format(time.RFC3339),
			}

			c.Hub.Private <- privateMsg{
				to:   incoming.To,
				from: c.UserID,
				out:  outgoing,
			}

			go S.SaveMessage(
				messageID,
				c.UserID,
				incoming.To,
				incoming.Message,
				incoming.ReplyToMessageID,
			)

		case "delete":

			if incoming.To == "" {

				log.Println("⚠️ Delete missing receiver")
				continue
			}

			go S.DeleteForEveryone(
				incoming.MessageID,
			)

			c.Hub.Delete <- privateMsg{
				to:   incoming.To,
				from: c.UserID,
				out: Outgoing{
					Type:      "delete",
					MessageID: incoming.MessageID,
					DeletedForEveryone:   true,
				},
			}

		case "list":

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

			log.Println(
				"⚠️ Unknown message type:",
				incoming.Type,
			)
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