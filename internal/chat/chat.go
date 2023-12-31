package chat

import (
	"context"
	"fmt"
	"github.com/felguerez/grpchat/internal/db"
	"github.com/felguerez/grpchat/internal/wschat"
	"github.com/felguerez/grpchat/proto"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"io"
	"time"
)

type Server struct {
	chat.UnimplementedChatServiceServer
	Logger *zap.Logger
	Hub    *wschat.Hub
}

func (s *Server) SendMessage(ctx context.Context, req *chat.SendMessageRequest) (*chat.SendMessageResponse, error) {
	message := db.Message{ // Replace with the actual struct definition
		UserID:         req.UserId,
		Content:        req.Content,
		ConversationID: req.ConversationId,
		Timestamp:      time.Now().Unix(),
	}
	s.Logger.Info("Received Message:", zap.Any("message", message))
	err := db.PutMessage(message)
	if err != nil {
		s.Logger.Info("damn")
		s.Logger.Error("Error putting message", zap.Error(err))
		return nil, err
	}

	return &chat.SendMessageResponse{Status: "Success"}, nil
}

func (s *Server) JoinConversation(ctx context.Context, req *chat.JoinConversationRequest) (*chat.JoinConversationResponse, error) {
	if req.UserId == "" || req.ConversationId == "" {
		return nil, fmt.Errorf("UserId and ConversationId cannot be empty")
	}

	err := db.AddMemberToConversation(req.ConversationId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &chat.JoinConversationResponse{Status: "Success"}, nil
}

func (s *Server) CreateConversation(ctx context.Context, req *chat.CreateConversationRequest) (*chat.CreateConversationResponse, error) {
	s.Logger.Info("Received CreateConversation request", zap.Any("request", req))
	currentTime := time.Now().Unix()
	conversation := db.Conversation{
		ID:        uuid.New().String(),
		Name:      req.Name,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
		CreatedBy: req.OwnerId,
		Members:   []string{req.OwnerId},
	}
	s.Logger.Info("Received Conversation:", zap.Any("conversation", conversation))
	err := db.PutConversation(conversation)
	if err != nil {
		s.Logger.Error("Error occurred:", zap.Error(err))
		return nil, err
	}
	return &chat.CreateConversationResponse{ConversationId: conversation.ID}, nil
}
func (s *Server) GetConversation(ctx context.Context, req *chat.GetConversationRequest) (*chat.GetConversationResponse, error) {
	if req.ConversationId == "" {
		return nil, fmt.Errorf("req.ConversationId cannot be empty")
	}

	conversation, messages, err := db.GetConversationWithMessages(req.ConversationId)
	if err != nil {
		return nil, err
	}

	protoConversation := convertToProtoConversation(conversation)
	protoMessages := convertToProtoMessages(messages)
	s.Logger.Info("Returning conversation", zap.Any("conversation", protoConversation))
	s.Logger.Info("Returning messages", zap.Any("messages", protoMessages))

	return &chat.GetConversationResponse{
		Conversation: protoConversation,
		Messages:     protoMessages,
	}, nil
}

// Convert db.Message to chat.Message
func convertToProtoMessages(messages []db.Message) []*chat.Message {
	var protoMessages []*chat.Message
	for _, msg := range messages {
		protoMsg := &chat.Message{
			UserId:         msg.UserID,
			Content:        msg.Content,
			ConversationId: msg.ConversationID,
		}
		protoMessages = append(protoMessages, protoMsg)
	}
	return protoMessages
}

func convertToProtoConversation(conversation db.Conversation) *chat.Conversation {
	return &chat.Conversation{
		Id:      conversation.ID,
		Name:    conversation.Name,
		OwnerId: conversation.CreatedBy,
		Members: conversation.Members,
	}
}

func (s *Server) GetConversations(ctx context.Context, req *chat.GetConversationsRequest) (*chat.GetConversationsResponse, error) {
	if req.Limit == 0 {
		req.Limit = 100
	}
	if req.Limit < 0 {
		return nil, fmt.Errorf("req.Limit should be a positive integer")
	}

	if req.UserId == "" {
		return nil, fmt.Errorf("req.UserID cannot be empty")
	}

	conversations, err := db.GetConversations(req.UserId, req.Limit, req.SortBy)
	if err != nil {
		return nil, err
	}

	protoConversations := convertToProtoConversations(conversations)

	return &chat.GetConversationsResponse{
		Conversations: protoConversations,
	}, nil
}
func convertToProtoConversations(conversations []db.Conversation) []*chat.Conversation {
	var protoConversations []*chat.Conversation
	for _, c := range conversations {
		protoConversations = append(protoConversations, &chat.Conversation{
			Id:      c.ID,
			Name:    c.Name,
			OwnerId: c.CreatedBy,
			Members: c.Members,
		})
	}
	return protoConversations
}

func (s *Server) ChatStream(stream chat.ChatService_ChatStreamServer) error {
	s.Logger.Info("<------ Received ChatStream request", zap.Time("timestamp", time.Now()),
		zap.String("method", "ChatStream"))
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// Client is done sending messages
			return nil
		}
		if err != nil {
			return err
		}
		message := db.Message{
			UserID:         req.UserId,
			Content:        req.Content,
			ConversationID: req.ConversationId,
			Timestamp:      time.Now().Unix(),
		}
		s.Logger.Info("Received Message from ChatStream:", zap.Any("message", message))
		err = db.PutMessage(message)
		if err != nil {
			s.Logger.Info("damn")
			s.Logger.Error("Error putting message", zap.Error(err))
			return err
		}
		s.Logger.Info("Time to broadcast", zap.Time("timestamp", time.Now()),
			zap.String("method", "ChatStream"))
		convertedMessage := convertMessageForWebSocket(message)
		s.Hub.Broadcast <- convertedMessage

		// Send a message back to the client
		if err := stream.Send(&chat.SendMessageRequest{
			UserId:         "server",
			Content:        "Acknowledgment for message received",
			ConversationId: "",
		}); err != nil {
			return err
		}
	}
}

func convertMessageForWebSocket(message db.Message) []byte {
	return []byte(fmt.Sprintf(`{"userId": "%s", "content": "%s", "conversationId": "%s"}`, message.UserID, message.Content, message.ConversationID))
}
