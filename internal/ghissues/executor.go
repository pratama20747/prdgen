package ghissues

import "context"

// Executor adalah abstraksi untuk benar-benar membuat issue di GitHub.
// Sama seperti llm.Provider, ini di-interface-kan supaya testable tanpa
// perlu binary `gh` asli atau akses network -- lihat MockExecutor.
type Executor interface {
	CreateIssue(ctx context.Context, issue Issue) (output string, err error)
}
