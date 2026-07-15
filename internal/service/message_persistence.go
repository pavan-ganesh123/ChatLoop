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
		"https://codecache-13ic.onrender.com/graphql",
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
	io.ReadAll(resp.Body)

}

func DeleteForEveryone(
    messageID string,
) {

    query := `
    mutation DeleteForEveryone(
        $messageId:String!
    ){
        deleteForEveryone(
            messageId:$messageId
        ){
            id
        }
    }`

    body := GraphQLRequest{
        Query: query,
        Variables: map[string]interface{}{
            "messageId": messageID,
        },
    }

    jsonBody, _ := json.Marshal(body)

    http.Post(
        "https://codecache-13ic.onrender.com/graphql",
        "application/json",
        bytes.NewBuffer(jsonBody),
    )
}

func SaveSharedPost(
	messageID string,
	senderID string,
	receiverID string,
	postID string,
) {

	query := `
	mutation SaveSharedPost(
		$messageId: String!,
		$senderId: ID!,
		$receiverId: ID!,
		$sharedPostId: ID!
	) {
		saveSharedPost(
			messageId: $messageId,
			senderId: $senderId,
			receiverId: $receiverId,
			sharedPostId: $sharedPostId
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
			"sharedPostId": postID,
		},
	}

	jsonBody, _ := json.Marshal(body)

	resp, err := http.Post(
		"https://codecache-13ic.onrender.com/graphql",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)

	if err != nil {
		log.Println(
			"❌ Failed to save shared post:",
			err,
		)
		return
	}

	defer resp.Body.Close()

	io.ReadAll(resp.Body)

	
}