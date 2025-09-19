package repository

import (
	"context"

	"github.com/rendaman0215/simple_ai_agent/internal/domain/entity"
)

// AIRepository はAIサービスとのやり取りを抽象化するリポジトリインターフェース
type AIRepository interface {
	// AskAI はAIにプロンプトを送信してレスポンスを取得する
	AskAI(ctx context.Context, request *entity.AIRequest) (*entity.AIResponse, error)

	// AskAIStream はAIにプロンプトを送信してストリーミングレスポンスを取得する
	AskAIStream(ctx context.Context, request *entity.AIRequest) (<-chan *entity.AIResponse, <-chan error)

	// HealthCheck はAIサービスの健康状態を確認する
	HealthCheck(ctx context.Context) error
}
