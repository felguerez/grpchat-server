package chat

import (
	"context"
	"github.com/felguerez/grpchat/proto"
)

type Server struct {
	chat.UnimplementedChatServiceServer
}

func (s *Server) SendMessage(ctx context.Context, req *chat.MessageRequest) (*chat.MessageResponse, error) {
	// Handle sending a message
	return &chat.MessageResponse{Status: "Success"}, nil
}

func (s *Server) JoinChat(req *chat.StreamingRequest, stream chat.ChatService_JoinChatServer) error {
	// Handle joining the chat
	return nil
}
