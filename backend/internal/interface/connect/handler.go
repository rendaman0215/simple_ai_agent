package connecthandler

import (
	"context"
	"time"

	connect "connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/rendaman0215/simple_ai_agent/internal/usecase"
	aiv1 "github.com/rendaman0215/simple_ai_agent/proto/gen/go/mahjong/ai/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MahjongAIConnectHandler はConnect用サービス実装
type MahjongAIConnectHandler struct {
	aiUsecase *usecase.AIUsecase
	logger    *logrus.Logger
}

// NewMahjongAIConnectHandler は新しいハンドラを作成
func NewMahjongAIConnectHandler(aiUsecase *usecase.AIUsecase, logger *logrus.Logger) *MahjongAIConnectHandler {
	return &MahjongAIConnectHandler{aiUsecase: aiUsecase, logger: logger}
}

// AskMahjongAI は同期API
func (h *MahjongAIConnectHandler) AskMahjongAI(ctx context.Context, req *connect.Request[aiv1.AskMahjongAIRequest]) (*connect.Response[aiv1.AskMahjongAIResponse], error) {
	startTime := time.Now()

	requestID := uuid.New().String()
	if req.Msg.GetMetadata() != nil && req.Msg.GetMetadata().GetRequestId() != "" {
		requestID = req.Msg.GetMetadata().GetRequestId()
	}

	h.logger.WithField("request_id", requestID).Info("[connect] AskMahjongAI called")

	if req.Msg.GetPrompt() == "" {
		res := &aiv1.AskMahjongAIResponse{
			Result: &aiv1.AskMahjongAIResponse_Error{Error: &aiv1.ErrorInfo{
				Code:    "INVALID_ARGUMENT",
				Message: "prompt cannot be empty",
				Details: "The prompt field is required and cannot be empty",
			}},
			Metadata: &aiv1.ResponseMetadata{
				RequestId:        requestID,
				Timestamp:        timestamppb.New(time.Now()),
				ProcessingTimeMs: time.Since(startTime).Milliseconds(),
				ServerVersion:    "1.0.0",
			},
		}
		return connect.NewResponse(res), nil
	}

	maxTokens := req.Msg.GetMaxTokens()
	if maxTokens <= 0 {
		maxTokens = 1000
	}
	temperature := req.Msg.GetTemperature()
	if temperature <= 0 {
		temperature = 0.7
	}

	response, err := h.aiUsecase.AskMahjongAI(ctx, req.Msg.GetPrompt(), maxTokens, temperature, req.Msg.GetContext())
	if err != nil {
		h.logger.WithError(err).Error("[connect] Failed to process AI request")
		res := &aiv1.AskMahjongAIResponse{
			Result: &aiv1.AskMahjongAIResponse_Error{Error: &aiv1.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
				Details: "Failed to process AI request",
			}},
			Metadata: &aiv1.ResponseMetadata{
				RequestId:        requestID,
				Timestamp:        timestamppb.New(time.Now()),
				ProcessingTimeMs: time.Since(startTime).Milliseconds(),
				ServerVersion:    "1.0.0",
			},
		}
		return connect.NewResponse(res), nil
	}

	res := &aiv1.AskMahjongAIResponse{
		Result: &aiv1.AskMahjongAIResponse_Response{Response: response.Response},
		Metadata: &aiv1.ResponseMetadata{
			RequestId:        requestID,
			Timestamp:        timestamppb.New(time.Now()),
			ProcessingTimeMs: response.ProcessingMs,
			ServerVersion:    "1.0.0",
		},
		TokensUsed: response.TokensUsed,
		Confidence: response.Confidence,
	}
	return connect.NewResponse(res), nil
}

// AskMahjongAIStream はサーバーストリームAPI
func (h *MahjongAIConnectHandler) AskMahjongAIStream(ctx context.Context, req *connect.Request[aiv1.AskMahjongAIRequest], stream *connect.ServerStream[aiv1.AskMahjongAIStreamResponse]) error {
	requestID := uuid.New().String()
	if req.Msg.GetMetadata() != nil && req.Msg.GetMetadata().GetRequestId() != "" {
		requestID = req.Msg.GetMetadata().GetRequestId()
	}

	h.logger.WithField("request_id", requestID).Info("[connect] AskMahjongAIStream called")

	if req.Msg.GetPrompt() == "" {
		return stream.Send(&aiv1.AskMahjongAIStreamResponse{
			Chunk: &aiv1.AskMahjongAIStreamResponse_Error{Error: &aiv1.ErrorInfo{
				Code:    "INVALID_ARGUMENT",
				Message: "prompt cannot be empty",
				Details: "The prompt field is required and cannot be empty",
			}},
			IsFinal: true,
		})
	}

	maxTokens := req.Msg.GetMaxTokens()
	if maxTokens <= 0 {
		maxTokens = 1000
	}
	temperature := req.Msg.GetTemperature()
	if temperature <= 0 {
		temperature = 0.7
	}

	respChan, errChan := h.aiUsecase.AskMahjongAIStream(ctx, req.Msg.GetPrompt(), maxTokens, temperature, req.Msg.GetContext())
	for {
		select {
		case r, ok := <-respChan:
			if !ok {
				return stream.Send(&aiv1.AskMahjongAIStreamResponse{
					Chunk: &aiv1.AskMahjongAIStreamResponse_Metadata{Metadata: &aiv1.ResponseMetadata{
						RequestId:        requestID,
						Timestamp:        timestamppb.New(time.Now()),
						ProcessingTimeMs: 0,
						ServerVersion:    "1.0.0",
					}},
					IsFinal: true,
				})
			}
			if r.Response != "" {
				if err := stream.Send(&aiv1.AskMahjongAIStreamResponse{Chunk: &aiv1.AskMahjongAIStreamResponse_TextChunk{TextChunk: r.Response}, IsFinal: false}); err != nil {
					return err
				}
			}
		case err := <-errChan:
			if err != nil {
				h.logger.WithError(err).Error("[connect] Failed to process streaming AI request")
				return stream.Send(&aiv1.AskMahjongAIStreamResponse{
					Chunk: &aiv1.AskMahjongAIStreamResponse_Error{Error: &aiv1.ErrorInfo{
						Code:    "INTERNAL_ERROR",
						Message: err.Error(),
						Details: "Failed to process streaming AI request",
					}},
					IsFinal: true,
				})
			}
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// HealthCheck はヘルスチェックAPI
func (h *MahjongAIConnectHandler) HealthCheck(ctx context.Context, req *connect.Request[aiv1.HealthCheckRequest]) (*connect.Response[aiv1.HealthCheckResponse], error) {
	h.logger.Info("[connect] HealthCheck called")
	if err := h.aiUsecase.HealthCheck(ctx); err != nil {
		res := &aiv1.HealthCheckResponse{Status: aiv1.HealthCheckResponse_NOT_SERVING, Message: err.Error(), Timestamp: timestamppb.New(time.Now())}
		return connect.NewResponse(res), nil
	}
	res := &aiv1.HealthCheckResponse{Status: aiv1.HealthCheckResponse_SERVING, Message: "Service is healthy", Timestamp: timestamppb.New(time.Now())}
	return connect.NewResponse(res), nil
}
