package main

import (
	"database/sql"
	"log"
	"net/http"
	"fmt"
	_ "github.com/lib/pq"

	"realtime-chat/internal/hub"
	S "realtime-chat/internal/service"
	ws "realtime-chat/internal/websocket"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found")
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host,
		port,
		user,
		password,
		dbname,
		sslmode,
	)

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