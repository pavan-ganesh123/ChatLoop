package service

import (
	"database/sql"
	"log"
)

var DB *sql.DB

// SET DATABASE CONNECTION
func SetDB(db *sql.DB) {
	DB = db
}

// CHECK IF EITHER USER BLOCKED THE OTHER
func CheckIfBlocked(
	senderID string,
	receiverID string,
) bool {

	// SAFETY CHECK
	if DB == nil {
		log.Println("❌ DB connection is nil")
		return false
	}

	query := `
	SELECT COUNT(*)
	FROM friends
	WHERE
	(
		user_id = $1
		AND friend_id = $2
		AND status = 'BLOCKED'
	)
	OR
	(
		user_id = $2
		AND friend_id = $1
		AND status = 'BLOCKED'
	)
	`

	var count int

	err := DB.QueryRow(
		query,
		senderID,
		receiverID,
	).Scan(&count)

	if err != nil {
		log.Println("❌ CheckIfBlocked query failed:", err)
		return false
	}

	if count > 0 {
		log.Println(
			"🚫 Block relation found between",
			senderID,
			"and",
			receiverID,
		)
		return true
	}

	return false
}