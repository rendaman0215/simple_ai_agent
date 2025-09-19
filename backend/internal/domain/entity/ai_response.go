package entity

// AIResponse は麻雀AIからのレスポンスを表すエンティティ
type AIResponse struct {
	Response     string
	TokensUsed   int32
	Confidence   float32
	ProcessingMs int64
}

// NewAIResponse は新しいAIResponseを作成する
func NewAIResponse(response string) *AIResponse {
	return &AIResponse{
		Response:     response,
		TokensUsed:   0,
		Confidence:   0.0,
		ProcessingMs: 0,
	}
}

// NewAIResponseWithMetrics はメトリクス付きで新しいAIResponseを作成する
func NewAIResponseWithMetrics(response string, tokensUsed int32, confidence float32, processingMs int64) *AIResponse {
	return &AIResponse{
		Response:     response,
		TokensUsed:   tokensUsed,
		Confidence:   confidence,
		ProcessingMs: processingMs,
	}
}
