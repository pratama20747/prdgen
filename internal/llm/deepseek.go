package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultDeepSeekBaseURL = "https://api.deepseek.com/chat/completions"

type DeepSeekProvider struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

type DeepSeekOption func(*DeepSeekProvider)

func WithBaseURL(url string) DeepSeekOption {
	return func(p *DeepSeekProvider) { p.baseURL = url }
}

func WithHTTPClient(c *http.Client) DeepSeekOption {
	return func(p *DeepSeekProvider) { p.httpClient = c }
}

func NewDeepSeekProvider(apiKey, model string, opts ...DeepSeekOption) *DeepSeekProvider {
	p := &DeepSeekProvider{
		apiKey:  apiKey,
		baseURL: defaultDeepSeekBaseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 1200 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *DeepSeekProvider) Name() string { return "deepseek:" + p.model }

type deepSeekChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepSeekChatRequest struct {
	Model       string                `json:"model"`
	Messages    []deepSeekChatMessage `json:"messages"`
	Temperature float64               `json:"temperature,omitempty"`
	MaxTokens   int                   `json:"max_tokens,omitempty"`
	Stream      bool                  `json:"stream"`
}

type deepSeekChatResponse struct {
	Choices []struct {
		Message struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (p *DeepSeekProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	msgs := make([]deepSeekChatMessage, 0, len(req.Messages)+1)
	if req.SystemPrompt != "" {
		msgs = append(msgs, deepSeekChatMessage{Role: "system", Content: req.SystemPrompt})
	}
	for _, m := range req.Messages {
		msgs = append(msgs, deepSeekChatMessage{Role: m.Role, Content: m.Content})
	}

	body := deepSeekChatRequest{
		Model:       p.model,
		Messages:    msgs,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      false,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("deepseek: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL, bytes.NewReader(payload))
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("deepseek: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("deepseek: request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("deepseek: read response: %w", err)
	}

	var parsed deepSeekChatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return CompletionResponse{}, fmt.Errorf("deepseek: parse response (status %d): %w, body=%s", resp.StatusCode, err, truncate(raw, 500))
	}

	if resp.StatusCode != http.StatusOK {
		if parsed.Error != nil {
			return CompletionResponse{}, fmt.Errorf("deepseek: api error (status %d): %s", resp.StatusCode, parsed.Error.Message)
		}
		return CompletionResponse{}, fmt.Errorf("deepseek: unexpected status %d: %s", resp.StatusCode, truncate(raw, 500))
	}

	if len(parsed.Choices) == 0 {
		return CompletionResponse{}, fmt.Errorf("deepseek: empty choices in response")
	}

	choice := parsed.Choices[0]

	if choice.Message.Content == "" {
		reasoningLen := len(choice.Message.ReasoningContent)
		return CompletionResponse{}, fmt.Errorf(
			"deepseek: model %q returned empty content (finish_reason=%q, reasoning_content_len=%d, completion_tokens=%d). "+
				"Kemungkinan MaxTokens terlalu kecil untuk model reasoning ini, atau model habis token sebelum sempat menjawab. "+
				"Coba naikkan MaxTokens atau pakai model non-reasoning (mis. deepseek-chat).",
			p.model, choice.FinishReason, reasoningLen, parsed.Usage.CompletionTokens,
		)
	}

	return CompletionResponse{
		Content:      choice.Message.Content,
		InputTokens:  parsed.Usage.PromptTokens,
		OutputTokens: parsed.Usage.CompletionTokens,
		FinishReason: choice.FinishReason,
	}, nil
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "...(truncated)"
}
