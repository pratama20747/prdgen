package llm

import "context"

type MockProvider struct {
	Responses []string
	Err       error

	calls       int
	LastRequest CompletionRequest
}

func (m *MockProvider) Name() string { return "mock" }

func (m *MockProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	m.LastRequest = req
	if m.Err != nil {
		return CompletionResponse{}, m.Err
	}
	if len(m.Responses) == 0 {
		return CompletionResponse{Content: "mock response"}, nil
	}
	idx := m.calls
	if idx >= len(m.Responses) {
		idx = len(m.Responses) - 1
	}
	m.calls++
	return CompletionResponse{Content: m.Responses[idx]}, nil
}
