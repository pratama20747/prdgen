package ghissues

import (
	"context"
	"fmt"
)

// MockExecutor adalah implementasi Executor palsu untuk unit test.
type MockExecutor struct {
	Created []Issue
	Err     error
}

func (m *MockExecutor) CreateIssue(ctx context.Context, issue Issue) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	m.Created = append(m.Created, issue)
	return fmt.Sprintf("https://github.com/example/repo/issues/%d", len(m.Created)), nil
}
