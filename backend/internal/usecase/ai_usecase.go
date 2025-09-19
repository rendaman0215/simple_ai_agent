package usecase

import (
	"context"

	"github.com/rendaman0215/simple_ai_agent/internal/domain/entity"
	"github.com/rendaman0215/simple_ai_agent/internal/domain/repository"
	"github.com/sirupsen/logrus"
)

// AIUsecase は麻雀AIに関するビジネスロジックを管理する
type AIUsecase struct {
	aiRepo repository.AIRepository
	logger *logrus.Logger
}

// NewAIUsecase は新しいAIUsecaseを作成する
func NewAIUsecase(aiRepo repository.AIRepository, logger *logrus.Logger) *AIUsecase {
	return &AIUsecase{
		aiRepo: aiRepo,
		logger: logger,
	}
}

// AskMahjongAI は麻雀AIにプロンプトを送信してレスポンスを取得する
func (u *AIUsecase) AskMahjongAI(ctx context.Context, prompt string, maxTokens int32, temperature float32, context []string) (*entity.AIResponse, error) {
	// ログ出力
	u.logger.WithFields(logrus.Fields{
		"prompt_length": len(prompt),
		"max_tokens":    maxTokens,
		"temperature":   temperature,
		"context_count": len(context),
	}).Info("AI request received")

	// リクエストエンティティを作成
	var request *entity.AIRequest
	if maxTokens > 0 || temperature > 0 || len(context) > 0 {
		request = entity.NewAIRequestWithOptions(prompt, maxTokens, temperature, context)
	} else {
		request = entity.NewAIRequest(prompt)
	}
	
	// バリデーション
	if err := request.Validate(); err != nil {
		u.logger.WithError(err).Error("Request validation failed")
		return nil, err
	}

	// AIリポジトリを通してAIサービスにリクエストを送信
	response, err := u.aiRepo.AskAI(ctx, request)
	if err != nil {
		u.logger.WithError(err).Error("Failed to get AI response")
		return nil, err
	}

	u.logger.WithFields(logrus.Fields{
		"response_length": len(response.Response),
		"tokens_used":     response.TokensUsed,
		"confidence":      response.Confidence,
	}).Info("AI response received successfully")

	return response, nil
}

// AskMahjongAIStream は麻雀AIにプロンプトを送信してストリーミングレスポンスを取得する
func (u *AIUsecase) AskMahjongAIStream(ctx context.Context, prompt string, maxTokens int32, temperature float32, context []string) (<-chan *entity.AIResponse, <-chan error) {
	u.logger.WithFields(logrus.Fields{
		"prompt_length": len(prompt),
		"max_tokens":    maxTokens,
		"temperature":   temperature,
		"context_count": len(context),
	}).Info("AI stream request received")

	// リクエストエンティティを作成
	var request *entity.AIRequest
	if maxTokens > 0 || temperature > 0 || len(context) > 0 {
		request = entity.NewAIRequestWithOptions(prompt, maxTokens, temperature, context)
	} else {
		request = entity.NewAIRequest(prompt)
	}

	// バリデーション
	if err := request.Validate(); err != nil {
		u.logger.WithError(err).Error("Stream request validation failed")
		errChan := make(chan error, 1)
		errChan <- err
		close(errChan)
		return nil, errChan
	}

	// AIリポジトリを通してストリーミングリクエストを送信
	return u.aiRepo.AskAIStream(ctx, request)
}

// HealthCheck はAIサービスの健康状態を確認する
func (u *AIUsecase) HealthCheck(ctx context.Context) error {
	u.logger.Info("Health check request received")
	
	err := u.aiRepo.HealthCheck(ctx)
	if err != nil {
		u.logger.WithError(err).Error("Health check failed")
		return err
	}
	
	u.logger.Info("Health check passed")
	return nil
}
