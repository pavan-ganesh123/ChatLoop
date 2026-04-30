package ws

import (
	"log"
	"net/http"
	"realtime-chat/internal/auth"
	"realtime-chat/internal/hub"

	"github.com/gorilla/websocket"
	"strconv"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ServeWS(h *hub.Hub, w http.ResponseWriter, r *http.Request) {

	// 🔐 Step 1: Validate JWT
	token := r.URL.Query().Get("token")
	log.Println("🔥 Incoming WS Token:", token)
	user, err := auth.ValidateToken(token)
	if err != nil {
		log.Println("❌ Token validation failed:", err) // ✅ ADD THIS
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
		
	}
	log.Println("✅ Authenticated user:", user.Email)
	// 	user := struct {
	// 	ID    string
	// 	Email string
	// }{
	// 	ID:    "1",
	// 	Email: "test@example.com",
	// }
	// 🔌 Step 2: Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	log.Println("Origin:", r.Header.Get("Origin"))

	client := &hub.Client{
	Hub:    h,
	Conn:   conn,
	Send:   make(chan []byte, 256),
	UserID: strconv.Itoa(int(user.UserID)),     // from JWT / auth
	Email:  user.Email,  // from JWT / auth
}

	h.Register <- client

	go client.WritePump()
	go client.ReadPump()
}