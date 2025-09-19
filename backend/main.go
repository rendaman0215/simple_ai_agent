package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	connect "connectrpc.com/connect"
	"github.com/rendaman0215/simple_ai_agent/internal/infrastructure"
	"github.com/rendaman0215/simple_ai_agent/internal/interface/config"
	connectHandler "github.com/rendaman0215/simple_ai_agent/internal/interface/connect"
	grpcHandler "github.com/rendaman0215/simple_ai_agent/internal/interface/grpc"
	"github.com/rendaman0215/simple_ai_agent/internal/usecase"
	aiv1 "github.com/rendaman0215/simple_ai_agent/proto/gen/go/mahjong/ai/v1"
	aiv1connect "github.com/rendaman0215/simple_ai_agent/proto/gen/go/mahjong/ai/v1/aiv1connect"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
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

	logger.Info("Starting Mahjong AI Server (gRPC + Connect)...")

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

	// リスナーを作成 (gRPC)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		logger.WithError(err).Fatal("Failed to listen gRPC")
	}

	// グレースフルシャットダウンの設定
	go func() {
		logger.WithField("port", cfg.GRPCPort).Info("gRPC server started")
		if err := server.Serve(lis); err != nil {
			logger.WithError(err).Fatal("Failed to serve gRPC server")
		}
	}()

	// Connect ハンドラを作成
	connectSvc := connectHandler.NewMahjongAIConnectHandler(aiUsecase, logger)
	path, connectHTTPHandler := aiv1connect.NewMahjongAIServiceHandler(connectSvc,
		connect.WithCompressMinBytes(1024),
		connect.WithReadMaxBytes(10*1024*1024),
	)

	// HTTPサーバ (h2c) を起動
	mux := http.NewServeMux()
	mux.Handle(path, connectHTTPHandler)
	// CORS: シンプルにワイルドカード対応（必要に応じて強化）
	cors := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", cfg.CORSAllowOrigins)
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Expose-Headers", "Connect-Content-Encoding, Connect-Accept-Encoding")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler: h2c.NewHandler(cors(mux), &http2.Server{}),
	}

	go func() {
		logger.WithFields(logrus.Fields{"port": cfg.HTTPPort, "path": path}).Info("Connect HTTP server started")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// シグナルを待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down servers...")
	server.GracefulStop()
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctxShutdown); err != nil {
		logger.WithError(err).Error("HTTP server shutdown error")
	}
	logger.Info("Servers stopped")
}
