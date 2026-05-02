package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"realtime-chat/internal/hub"
	S "realtime-chat/internal/service"
	ws "realtime-chat/internal/websocket"
)

func main() {

	// POSTGRES CONNECTION
	connStr := "host=ep-red-scene-anvjjntv-pooler.c-6.us-east-1.aws.neon.tech port=5432 user=neondb_owner password=npg_qBuxJ4OQiTe1 dbname=neondb sslmode=require"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("❌ DB Connection Error:", err)
	}

	// TEST CONNECTION
	err = db.Ping()
	if err != nil {
		log.Fatal("❌ DB Ping Failed:", err)
	}

	log.Println("✅ PostgreSQL Connected")

	// SET DB FOR SERVICES
	S.SetDB(db)

	// CREATE HUB
	h := hub.NewHub()

	go h.Run()

	// WEBSOCKET ROUTE
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWS(h, w, r)
	})

	log.Println("✅ WebSocket endpoint ready at /ws")
	log.Println("🚀 Server started on :8081")

	err = http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal("❌ Server Error:", err)
	}
}