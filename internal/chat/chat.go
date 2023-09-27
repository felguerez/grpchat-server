package chat

import (
	"context"
	"fmt"
	"github.com/felguerez/grpchat/proto"
)

type Server struct {
	chat.UnimplementedChatServiceServer
}

func (s *Server) SendMessage(ctx context.Context, req *chat.MessageRequest) (*chat.MessageResponse, error) {
	fmt.Printf("Received username %s, content %s\n", req.Username, req.Content)
	return &chat.MessageResponse{Status: "Success"}, nil
}

func (s *Server) JoinChat(req *chat.StreamingRequest, stream chat.ChatService_JoinChatServer) error {
	return nil
}
