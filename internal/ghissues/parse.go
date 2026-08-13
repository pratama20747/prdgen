package ghissues

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Parse mengambil raw output dari LLM dan mem-parsingnya jadi slice Issue.
// LLM kadang membungkus JSON dalam markdown code fence (```json ... ```)
// walau prompt sudah eksplisit melarang itu -- fungsi ini membersihkan fence
// itu dulu sebelum parse, supaya tidak gagal cuma gara-gara noise formatting.
func Parse(raw string) ([]Issue, error) {
	cleaned := stripCodeFence(raw)

	var issues []Issue
	if err := json.Unmarshal([]byte(cleaned), &issues); err != nil {
		return nil, fmt.Errorf(
			"ghissues: gagal parse JSON (%w). Cek isi file ISSUES.json secara manual -- "+
				"kemungkinan model menyisipkan teks di luar JSON, atau JSON-nya terpotong "+
				"karena kehabisan token", err,
		)
	}
	if len(issues) == 0 {
		return nil, fmt.Errorf("ghissues: JSON valid tapi tidak berisi issue apapun")
	}
	for i, iss := range issues {
		if strings.TrimSpace(iss.Title) == "" {
			return nil, fmt.Errorf("ghissues: issue ke-%d tidak punya title", i+1)
		}
	}
	return issues, nil
}

func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	if idx := strings.Index(s, "\n"); idx != -1 {
		s = s[idx+1:]
	}
	s = strings.TrimSuffix(strings.TrimSpace(s), "```")
	return strings.TrimSpace(s)
}

// ParseSingle mem-parsing SATU object Issue (bukan array), dipakai untuk
// hasil revisi single-issue dari RunReviseIssue.
func ParseSingle(raw string) (Issue, error) {
	cleaned := stripCodeFence(raw)
	var issue Issue
	if err := json.Unmarshal([]byte(cleaned), &issue); err != nil {
		return Issue{}, fmt.Errorf("ghissues: gagal parse hasil revisi (%w)", err)
	}
	if strings.TrimSpace(issue.Title) == "" {
		return Issue{}, fmt.Errorf("ghissues: hasil revisi tidak punya title")
	}
	return issue, nil
}
