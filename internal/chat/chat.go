package chat

import (
	"context"
	"fmt"
	"github.com/felguerez/grpchat/internal/db"
	"github.com/felguerez/grpchat/proto"
	"github.com/google/uuid"
	"time"
)

type Server struct {
	chat.UnimplementedChatServiceServer
}

func (s *Server) SendMessage(ctx context.Context, req *chat.SendMessageRequest) (*chat.SendMessageResponse, error) {
	message := db.Message{ // Replace with the actual struct definition
		UserID:         req.UserId,
		Content:        req.Content,
		ConversationID: req.ConversationId,
		Timestamp:      time.Now().Unix(),
	}
	fmt.Println("Received Message:")
	fmt.Println(message)
	err := db.PutMessage(message)
	if err != nil {
		fmt.Println("damn")
		fmt.Sprintf("Uh oh an error when putting message: %s", err.Error())
		return nil, err
	}

	return &chat.SendMessageResponse{Status: "Success"}, nil
}

func (s *Server) JoinConversation(req *chat.JoinConversationRequest, stream chat.ChatService_JoinConversationServer) error {
	return nil
}

func (s *Server) CreateConversation(ctx context.Context, req *chat.CreateConversationRequest) (*chat.CreateConversationResponse, error) {
	conversation := db.Conversation{
		ID:        uuid.New().String(),
		Name:      req.Name,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
		CreatedBy: req.OwnerId,
	}
	fmt.Println("Received Conversation:")
	fmt.Println(conversation)
	err := db.PutConversation(conversation)
	if err != nil {
		fmt.Println("Error occurred:")
		fmt.Sprintf("Error when putting conversation: %s", err.Error())
		return nil, err
	}
	return &chat.CreateConversationResponse{ConversationId: conversation.ID}, nil
}
func (s *Server) GetConversations(ctx context.Context, req *chat.GetConversationsRequest) (*chat.GetConversationsResponse, error) {
	// Validate request (e.g., check that limit is positive, sort_by is a valid field, etc.)

	// Retrieve conversations from the database.
	// This is a placeholder; you'll need to implement db.GetConversations based on your database schema.
	limit := int64(req.Limit)
	conversations, err := db.GetConversations(req.UserId, limit, req.SortBy)
	if err != nil {
		return nil, err
	}

	// Convert conversations to the format expected by the proto message.
	// This is a placeholder; you'll need to implement the conversion based on your actual data structures.
	protoConversations := convertToProtoConversations(conversations)

	return &chat.GetConversationsResponse{
		Conversations: protoConversations,
	}, nil
}
