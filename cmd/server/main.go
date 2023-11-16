package main

import (
	"fmt"
	"github.com/felguerez/grpchat/internal/auth"
	"github.com/felguerez/grpchat/internal/chat"
	"github.com/felguerez/grpchat/internal/handlers"
	"github.com/felguerez/grpchat/internal/wschat"
	chatpb "github.com/felguerez/grpchat/proto"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"os"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger, _ := zap.NewProduction()
		defer logger.Sync()

		logger.Info("Incoming request",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
		)
		next.ServeHTTP(w, r)
	})
}

// loadEnvFile loads the appropriate .env files based on the APP_ENV variable and logs the files used.
func loadEnvFile(logger *zap.Logger) error {
	var envFiles []string

	if os.Getenv("APP_ENV") == "development" {
		envFiles = append(envFiles, ".env", ".env.development")
	} else {
		envFiles = append(envFiles, ".env")
	}

	// Load base .env file
	err := godotenv.Load(envFiles[0])
	if err != nil {
		logger.Error("Error loading .env file", zap.String("file", envFiles[0]), zap.Error(err))
		return err
	}
	logger.Info("Using env file", zap.String("file", envFiles[0]))

	// If in development, overload with .env.development
	if len(envFiles) > 1 {
		err = godotenv.Overload(envFiles[1])
		if err != nil {
			logger.Error("Error overloading .env file", zap.String("file", envFiles[1]), zap.Error(err))
			return err
		}
		logger.Info("Overloading env file", zap.String("file", envFiles[1]))
	}

	return nil
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	loadEnvFile(logger)

	auth.InitializeSpotifyOauthConfig(os.Getenv("SPOTIFY_CLIENT_ID"), os.Getenv("SPOTIFY_CLIENT_SECRET"), os.Getenv("SPOTIFY_REDIRECT_CALLBACK_URL"), []string{"user-read-email"})
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/messages", handlers.HandleGetAllMessages) // @TODO: implement grpc endpoint for this data

	http.HandleFunc("/", handlers.Root)
	http.Handle("/api/", LoggingMiddleware(http.StripPrefix("/api", handlers.RequireAuthorizationToken(apiMux))))
	http.Handle("/login", LoggingMiddleware(handlers.HandleLogin(logger)))
	http.Handle("/callback", LoggingMiddleware(handlers.HandleCallback(logger)))

	hub := wschat.NewHub(logger)
	http.HandleFunc("/api/conversations/", wschat.ServeWebSocketConnection(hub))

	httpPort := "8080"
	go func() {
		logger.Info(fmt.Sprintf("HTTP server is running on port :%s", httpPort))
		if err := http.ListenAndServe(":"+httpPort, nil); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	grpcServer := grpc.NewServer()
	chatpb.RegisterChatServiceServer(grpcServer, &chat.Server{})
	reflection.Register(grpcServer)

	grpcPort := "50051"
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	logger.Info(fmt.Sprintf("gRPC server is running on port :%s", grpcPort))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
