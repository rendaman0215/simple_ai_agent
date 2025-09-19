package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/rendaman0215/simple_ai_agent/internal/domain/entity"
	"github.com/rendaman0215/simple_ai_agent/internal/domain/repository"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

// GeminiClient はGemini APIクライアントの実装
type GeminiClient struct {
	client *genai.Client
	model  *genai.GenerativeModel
	logger *logrus.Logger
}

// NewGeminiClient は新しいGeminiClientを作成する
func NewGeminiClient(apiKey string, logger *logrus.Logger) (repository.AIRepository, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Gemini 2.5 Flash モデルを使用
	model := client.GenerativeModel("gemini-2.5-flash")

	// 麻雀AIとしての設定を追加
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{
			genai.Text("あなたは麻雀の専門家です。麻雀に関する質問に対して、正確で分かりやすい回答を日本語で提供してください。戦術、ルール、確率計算など、麻雀に関するあらゆる側面について回答できます。"),
		},
	}

	return &GeminiClient{
		client: client,
		model:  model,
		logger: logger,
	}, nil
}

// AskAI はGemini APIにプロンプトを送信してレスポンスを取得する
func (g *GeminiClient) AskAI(ctx context.Context, request *entity.AIRequest) (*entity.AIResponse, error) {
	startTime := time.Now()

	g.logger.WithFields(logrus.Fields{
		"prompt":      request.Prompt,
		"max_tokens":  request.MaxTokens,
		"temperature": request.Temperature,
		"context":     request.Context,
	}).Debug("Sending request to Gemini API")

	// モデルの設定を更新
	g.model.SetTemperature(request.Temperature)
	g.model.SetMaxOutputTokens(request.MaxTokens)

	// コンテキストがある場合は追加
	var parts []genai.Part
	for _, ctx := range request.Context {
		parts = append(parts, genai.Text(ctx))
	}
	parts = append(parts, genai.Text(request.Prompt))

	// Gemini APIにリクエストを送信
	resp, err := g.model.GenerateContent(ctx, parts...)
	if err != nil {
		g.logger.WithError(err).Error("Failed to generate content with Gemini API")
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	processingTime := time.Since(startTime).Milliseconds()

	// レスポンスが空でないことを確認
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		g.logger.Error("Received empty response from Gemini API")
		return nil, entity.ErrAIServiceUnavailable
	}

	// レスポンステキストを取得
	responseText := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			responseText += string(textPart)
		}
	}

	if responseText == "" {
		g.logger.Error("No text content in Gemini API response")
		return nil, entity.ErrAIServiceUnavailable
	}

	// メトリクスを含むレスポンスを作成
	tokensUsed := int32(0)
	if resp.UsageMetadata != nil {
		tokensUsed = int32(resp.UsageMetadata.TotalTokenCount)
	}

	confidence := float32(0.8) // Geminiは信頼度スコアを提供しないため、デフォルト値を使用

	g.logger.WithFields(logrus.Fields{
		"response_length": len(responseText),
		"tokens_used":     tokensUsed,
		"processing_time": processingTime,
	}).Debug("Received response from Gemini API")

	return entity.NewAIResponseWithMetrics(responseText, tokensUsed, confidence, processingTime), nil
}

// AskAIStream はGemini APIにプロンプトを送信してストリーミングレスポンスを取得する
func (g *GeminiClient) AskAIStream(ctx context.Context, request *entity.AIRequest) (<-chan *entity.AIResponse, <-chan error) {
	responseChan := make(chan *entity.AIResponse)
	errorChan := make(chan error, 1)

	go func() {
		defer close(responseChan)
		defer close(errorChan)

		startTime := time.Now()

		g.logger.WithFields(logrus.Fields{
			"prompt":      request.Prompt,
			"max_tokens":  request.MaxTokens,
			"temperature": request.Temperature,
			"context":     request.Context,
		}).Debug("Sending streaming request to Gemini API")

		// モデルの設定を更新
		g.model.SetTemperature(request.Temperature)
		g.model.SetMaxOutputTokens(request.MaxTokens)

		// コンテキストがある場合は追加
		var parts []genai.Part
		for _, ctx := range request.Context {
			parts = append(parts, genai.Text(ctx))
		}
		parts = append(parts, genai.Text(request.Prompt))

		// ストリーミングリクエストを送信
		iter := g.model.GenerateContentStream(ctx, parts...)

		fullResponse := ""
		for {
			resp, err := iter.Next()
			if err != nil {
				if err.Error() == "iterator is done" {
					break
				}
				errorChan <- fmt.Errorf("failed to get stream response: %w", err)
				return
			}

			// レスポンスチャンクを処理
			for _, candidate := range resp.Candidates {
				for _, part := range candidate.Content.Parts {
					if textPart, ok := part.(genai.Text); ok {
						chunkText := string(textPart)
						fullResponse += chunkText

						// チャンクレスポンスを送信
						responseChan <- entity.NewAIResponse(chunkText)
					}
				}
			}
		}

		processingTime := time.Since(startTime).Milliseconds()

		// 最終レスポンスのメタデータを送信
		tokensUsed := int32(len(fullResponse) / 4) // 概算
		confidence := float32(0.8)

		finalResponse := entity.NewAIResponseWithMetrics("", tokensUsed, confidence, processingTime)
		responseChan <- finalResponse

		g.logger.WithFields(logrus.Fields{
			"total_response_length": len(fullResponse),
			"processing_time":       processingTime,
		}).Debug("Completed streaming response from Gemini API")
	}()

	return responseChan, errorChan
}

// HealthCheck はGemini APIの健康状態を確認する
func (g *GeminiClient) HealthCheck(ctx context.Context) error {
	// 簡単なテストクエリを送信
	testRequest := entity.NewAIRequest("Hello")
	_, err := g.AskAI(ctx, testRequest)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	return nil
}

// Close はクライアントを閉じる
func (g *GeminiClient) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}
