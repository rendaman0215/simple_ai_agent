package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rendaman0215/simple_ai_agent/internal/usecase"
	aiv1 "github.com/rendaman0215/simple_ai_agent/proto/gen/go/mahjong/ai/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MahjongAIHandler はgRPCサービスのハンドラー
type MahjongAIHandler struct {
	aiv1.UnimplementedMahjongAIServiceServer
	aiUsecase *usecase.AIUsecase
	logger    *logrus.Logger
}

// NewMahjongAIHandler は新しいMahjongAIHandlerを作成する
func NewMahjongAIHandler(aiUsecase *usecase.AIUsecase, logger *logrus.Logger) *MahjongAIHandler {
	return &MahjongAIHandler{
		aiUsecase: aiUsecase,
		logger:    logger,
	}
}

// AskMahjongAI は麻雀AIへの質問を処理する
func (h *MahjongAIHandler) AskMahjongAI(ctx context.Context, req *aiv1.AskMahjongAIRequest) (*aiv1.AskMahjongAIResponse, error) {
	startTime := time.Now()

	// リクエストIDを生成
	requestID := uuid.New().String()
	if req.Metadata != nil && req.Metadata.RequestId != "" {
		requestID = req.Metadata.RequestId
	}

	h.logger.WithField("request_id", requestID).Info("AskMahjongAI called")

	// リクエストの妥当性を確認
	if req.Prompt == "" {
		h.logger.Error("Empty prompt received")
		return &aiv1.AskMahjongAIResponse{
			Result: &aiv1.AskMahjongAIResponse_Error{
				Error: &aiv1.ErrorInfo{
					Code:    "INVALID_ARGUMENT",
					Message: "prompt cannot be empty",
					Details: "The prompt field is required and cannot be empty",
				},
			},
			Metadata: &aiv1.ResponseMetadata{
				RequestId:        requestID,
				Timestamp:        timestamppb.New(time.Now()),
				ProcessingTimeMs: time.Since(startTime).Milliseconds(),
				ServerVersion:    "1.0.0",
			},
		}, nil
	}

	// デフォルト値の設定
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 1000
	}

	temperature := req.Temperature
	if temperature <= 0 {
		temperature = 0.7
	}

	// ユースケースを呼び出し
	response, err := h.aiUsecase.AskMahjongAI(ctx, req.Prompt, maxTokens, temperature, req.Context)
	if err != nil {
		h.logger.WithError(err).Error("Failed to process AI request")

		return &aiv1.AskMahjongAIResponse{
			Result: &aiv1.AskMahjongAIResponse_Error{
				Error: &aiv1.ErrorInfo{
					Code:    "INTERNAL_ERROR",
					Message: err.Error(),
					Details: "Failed to process AI request",
				},
			},
			Metadata: &aiv1.ResponseMetadata{
				RequestId:        requestID,
				Timestamp:        timestamppb.New(time.Now()),
				ProcessingTimeMs: time.Since(startTime).Milliseconds(),
				ServerVersion:    "1.0.0",
			},
		}, nil
	}

	h.logger.Info("AI request processed successfully")

	return &aiv1.AskMahjongAIResponse{
		Result: &aiv1.AskMahjongAIResponse_Response{
			Response: response.Response,
		},
		Metadata: &aiv1.ResponseMetadata{
			RequestId:        requestID,
			Timestamp:        timestamppb.New(time.Now()),
			ProcessingTimeMs: response.ProcessingMs,
			ServerVersion:    "1.0.0",
		},
		TokensUsed: response.TokensUsed,
		Confidence: response.Confidence,
	}, nil
}

// AskMahjongAIStream は麻雀AIへのストリーミング質問を処理する
func (h *MahjongAIHandler) AskMahjongAIStream(req *aiv1.AskMahjongAIRequest, stream aiv1.MahjongAIService_AskMahjongAIStreamServer) error {
	// リクエストIDを生成
	requestID := uuid.New().String()
	if req.Metadata != nil && req.Metadata.RequestId != "" {
		requestID = req.Metadata.RequestId
	}

	h.logger.WithField("request_id", requestID).Info("AskMahjongAIStream called")

	// リクエストの妥当性を確認
	if req.Prompt == "" {
		return stream.Send(&aiv1.AskMahjongAIStreamResponse{
			Chunk: &aiv1.AskMahjongAIStreamResponse_Error{
				Error: &aiv1.ErrorInfo{
					Code:    "INVALID_ARGUMENT",
					Message: "prompt cannot be empty",
					Details: "The prompt field is required and cannot be empty",
				},
			},
			IsFinal: true,
		})
	}

	// デフォルト値の設定
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 1000
	}

	temperature := req.Temperature
	if temperature <= 0 {
		temperature = 0.7
	}

	// ストリーミングユースケースを呼び出し
	responseChan, errorChan := h.aiUsecase.AskMahjongAIStream(stream.Context(), req.Prompt, maxTokens, temperature, req.Context)

	for {
		select {
		case response, ok := <-responseChan:
			if !ok {
				// チャンネルが閉じられた場合、最終メッセージを送信
				return stream.Send(&aiv1.AskMahjongAIStreamResponse{
					Chunk: &aiv1.AskMahjongAIStreamResponse_Metadata{
						Metadata: &aiv1.ResponseMetadata{
							RequestId:        requestID,
							Timestamp:        timestamppb.New(time.Now()),
							ProcessingTimeMs: 0,
							ServerVersion:    "1.0.0",
						},
					},
					IsFinal: true,
				})
			}

			// レスポンスチャンクを送信
			if response.Response != "" {
				if err := stream.Send(&aiv1.AskMahjongAIStreamResponse{
					Chunk: &aiv1.AskMahjongAIStreamResponse_TextChunk{
						TextChunk: response.Response,
					},
					IsFinal: false,
				}); err != nil {
					return err
				}
			}

		case err := <-errorChan:
			if err != nil {
				h.logger.WithError(err).Error("Failed to process streaming AI request")
				return stream.Send(&aiv1.AskMahjongAIStreamResponse{
					Chunk: &aiv1.AskMahjongAIStreamResponse_Error{
						Error: &aiv1.ErrorInfo{
							Code:    "INTERNAL_ERROR",
							Message: err.Error(),
							Details: "Failed to process streaming AI request",
						},
					},
					IsFinal: true,
				})
			}
			return nil

		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}

// HealthCheck はサービスの健康状態を確認する
func (h *MahjongAIHandler) HealthCheck(ctx context.Context, req *aiv1.HealthCheckRequest) (*aiv1.HealthCheckResponse, error) {
	h.logger.Info("HealthCheck called")

	// AIサービスのヘルスチェック
	err := h.aiUsecase.HealthCheck(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Health check failed")
		return &aiv1.HealthCheckResponse{
			Status:    aiv1.HealthCheckResponse_NOT_SERVING,
			Message:   err.Error(),
			Timestamp: timestamppb.New(time.Now()),
		}, nil
	}

	h.logger.Info("Health check passed")
	return &aiv1.HealthCheckResponse{
		Status:    aiv1.HealthCheckResponse_SERVING,
		Message:   "Service is healthy",
		Timestamp: timestamppb.New(time.Now()),
	}, nil
}
