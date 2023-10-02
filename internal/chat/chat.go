package chat

import (
	"context"
	"fmt"
	"github.com/felguerez/grpchat/internal/db"
	"github.com/felguerez/grpchat/proto"
	"time"
)

type Server struct {
	chat.UnimplementedChatServiceServer
}

func (s *Server) SendMessage(ctx context.Context, req *chat.SendMessageRequest) (*chat.SendMessageResponse, error) {
	message := db.Message{ // Replace with the actual struct definition
		UserID:         req.UserId,
		Content:        req.Content,
		ConversationID: 420,
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
