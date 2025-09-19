package entity

// AIRequest は麻雀AIへのリクエストを表すエンティティ
type AIRequest struct {
	Prompt      string
	MaxTokens   int32
	Temperature float32
	Context     []string
}

// NewAIRequest は新しいAIRequestを作成する
func NewAIRequest(prompt string) *AIRequest {
	return &AIRequest{
		Prompt:      prompt,
		MaxTokens:   1000,    // デフォルト値
		Temperature: 0.7,     // デフォルト値
		Context:     []string{},
	}
}

// NewAIRequestWithOptions はオプション付きで新しいAIRequestを作成する
func NewAIRequestWithOptions(prompt string, maxTokens int32, temperature float32, context []string) *AIRequest {
	return &AIRequest{
		Prompt:      prompt,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		Context:     context,
	}
}

// Validate はリクエストの妥当性を検証する
func (r *AIRequest) Validate() error {
	if r.Prompt == "" {
		return ErrEmptyPrompt
	}
	if r.Temperature < 0.0 || r.Temperature > 2.0 {
		return ErrInvalidTemperature
	}
	if r.MaxTokens <= 0 {
		return ErrInvalidMaxTokens
	}
	return nil
}
