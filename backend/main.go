package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/rendaman0215/simple_ai_agent/internal/infrastructure"
	"github.com/rendaman0215/simple_ai_agent/internal/interface/config"
	grpcHandler "github.com/rendaman0215/simple_ai_agent/internal/interface/grpc"
	"github.com/rendaman0215/simple_ai_agent/internal/usecase"
	aiv1 "github.com/rendaman0215/simple_ai_agent/proto/gen/go/mahjong/ai/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 設定を読み込み
	cfg := config.LoadConfig()

	// ロガーを設定
	logger := logrus.New()
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetLevel(level)
	}

	logger.Info("Starting Mahjong AI gRPC Server...")

	// Gemini APIキーの確認
	if cfg.GeminiAPIKey == "" {
		logger.Fatal("GEMINI_API_KEY environment variable is required")
	}

	// 依存関係を構築
	// Infrastructure層
	geminiClient, err := infrastructure.NewGeminiClient(cfg.GeminiAPIKey, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create Gemini client")
	}
	defer func() {
		if closer, ok := geminiClient.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				logger.WithError(err).Error("Failed to close Gemini client")
			}
		}
	}()

	// Usecase層
	aiUsecase := usecase.NewAIUsecase(geminiClient, logger)

	// Interface層
	handler := grpcHandler.NewMahjongAIHandler(aiUsecase, logger)

	// gRPCサーバーを作成
	server := grpc.NewServer()
	aiv1.RegisterMahjongAIServiceServer(server, handler)

	// リフレクションを有効にする（開発用）
	reflection.Register(server)

	// リスナーを作成
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		logger.WithError(err).Fatal("Failed to listen")
	}

	// グレースフルシャットダウンの設定
	go func() {
		logger.WithField("port", cfg.GRPCPort).Info("gRPC server started")
		if err := server.Serve(lis); err != nil {
			logger.WithError(err).Fatal("Failed to serve gRPC server")
		}
	}()

	// シグナルを待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down gRPC server...")
	server.GracefulStop()
	logger.Info("Server stopped")
}
