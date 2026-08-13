package llm

import "context"

type Message struct {
	Role    string
	Content string
}

type CompletionRequest struct {
	SystemPrompt string
	Messages     []Message
	Temperature  float64
	MaxTokens    int
}

type CompletionResponse struct {
	Content      string
	InputTokens  int
	OutputTokens int
	FinishReason string
}

type Provider interface {
	Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
	Name() string
}
