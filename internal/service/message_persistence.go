package service

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

func SaveMessage(
	messageID string,
	senderID string,
	receiverID string,
	content string,
	replyToMessageID string,
) {

	query := `
	mutation SaveMessage(
		$messageId: String!,
		$senderId: ID!,
		$receiverId: ID!,
		$content: String!,
		$replyToMessageId: String
	) {
		saveMessage(
			messageId: $messageId,
			senderId: $senderId,
			receiverId: $receiverId,
			content: $content,
			replyToMessageId: $replyToMessageId
		) {
			id
			messageId
		}
	}`

	body := GraphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"messageId": messageID,
			"senderId": senderID,
			"receiverId": receiverID,
			"content": content,
			"replyToMessageId": replyToMessageID,
		},
	}

	jsonBody, _ := json.Marshal(body)

	resp, err := http.Post(
		"http://localhost:8080/graphql",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)

	if err != nil {

		log.Println(
			"❌ Failed to save message:",
			err,
		)

		return
	}

	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)

	log.Println(
		"📦 GraphQL Response:",
		string(responseBody),
	)
}