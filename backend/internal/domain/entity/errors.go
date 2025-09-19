package entity

import "errors"

var (
	// ErrEmptyPrompt はプロンプトが空の場合のエラー
	ErrEmptyPrompt = errors.New("prompt cannot be empty")
	
	// ErrAIServiceUnavailable はAIサービスが利用できない場合のエラー
	ErrAIServiceUnavailable = errors.New("AI service is unavailable")
	
	// ErrInvalidRequest は無効なリクエストの場合のエラー
	ErrInvalidRequest = errors.New("invalid request")
	
	// ErrInvalidTemperature は無効な温度パラメータの場合のエラー
	ErrInvalidTemperature = errors.New("temperature must be between 0.0 and 2.0")
	
	// ErrInvalidMaxTokens は無効な最大トークン数の場合のエラー
	ErrInvalidMaxTokens = errors.New("max tokens must be greater than 0")
)
