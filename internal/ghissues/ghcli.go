package ghissues

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// GHCLIExecutor menjalankan `gh issue create` lewat GitHub CLI resmi.
type GHCLIExecutor struct {
	// Repo opsional, format "owner/repo". Kosong = gh menebak dari repo git
	// di direktori kerja saat ini (perilaku default gh).
	Repo string
}

// CreateIssue membuat satu issue di GitHub.
//
// PENTING soal keamanan: perintah dijalankan lewat exec.CommandContext
// dengan argumen sebagai slice terpisah -- BUKAN lewat shell (exec.Command
// ("sh", "-c", ...)). Ini krusial karena issue.Title dan issue.Body berasal
// dari output LLM yang bisa saja (sengaja atau tidak) berisi karakter
// bermakna khusus di shell seperti ; ` $() && |. Dengan argv terpisah,
// karakter-karakter itu SELALU diperlakukan sebagai bagian dari string data
// oleh proses `gh`, tidak pernah diinterpretasikan sebagai command tambahan
// oleh shell manapun -- karena memang tidak ada shell yang terlibat sama
// sekali di jalur eksekusi ini.
func (e *GHCLIExecutor) CreateIssue(ctx context.Context, issue Issue) (string, error) {
	args := []string{"issue", "create", "--title", issue.Title, "--body", issue.Body}
	if len(issue.Labels) > 0 {
		args = append(args, "--label", strings.Join(issue.Labels, ","))
	}
	if e.Repo != "" {
		args = append(args, "--repo", e.Repo)
	}

	cmd := exec.CommandContext(ctx, "gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gh issue create gagal: %w\nstderr: %s", err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

// EnsureLabels memastikan semua label di `labels` sudah ada di repo tujuan,
// dengan cara bikin yang belum ada. `gh issue create --label X` GAGAL kalau
// label X belum pernah dibuat di repo -- gh TIDAK auto-create label seperti
// beberapa tool lain. Draft issue dari LLM sering pakai label baru (mis.
// "phase-12") yang belum tentu ada di repo GitHub, jadi ini harus dijalankan
// sekali di awal, sebelum loop pembuatan issue, supaya tidak gagal di
// tengah jalan.
func (e *GHCLIExecutor) EnsureLabels(ctx context.Context, labels []string) error {
	existing, err := e.listLabels(ctx)
	if err != nil {
		return fmt.Errorf("gagal mengambil daftar label yang sudah ada: %w", err)
	}

	seen := map[string]bool{}
	for _, l := range existing {
		seen[l] = true
	}

	for _, label := range labels {
		label = strings.TrimSpace(label)
		if label == "" || seen[label] {
			continue
		}
		if err := e.createLabel(ctx, label); err != nil {
			return fmt.Errorf("gagal membuat label %q: %w", label, err)
		}
		seen[label] = true
	}
	return nil
}

func (e *GHCLIExecutor) listLabels(ctx context.Context) ([]string, error) {
	args := []string{"label", "list", "--json", "name", "--limit", "1000"}
	if e.Repo != "" {
		args = append(args, "--repo", e.Repo)
	}

	cmd := exec.CommandContext(ctx, "gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("gh label list gagal: %w\nstderr: %s", err, strings.TrimSpace(stderr.String()))
	}

	var rows []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &rows); err != nil {
		return nil, fmt.Errorf("gagal parse output gh label list: %w", err)
	}

	names := make([]string, 0, len(rows))
	for _, r := range rows {
		names = append(names, r.Name)
	}
	return names, nil
}

func (e *GHCLIExecutor) createLabel(ctx context.Context, name string) error {
	// --force supaya tidak error kalau ternyata label sudah dibuat di
	// antara pengecekan listLabels dan sini (race kecil, misalnya dibuat
	// manual oleh user lain di repo yang sama).
	args := []string{"label", "create", name, "--force"}
	if e.Repo != "" {
		args = append(args, "--repo", e.Repo)
	}

	cmd := exec.CommandContext(ctx, "gh", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh label create gagal: %w\nstderr: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

// CheckGHAvailable memastikan binary `gh` ada di PATH dan sudah login,
// dipanggil sekali sebelum loop pembuatan issue dimulai -- supaya gagal
// cepat dengan pesan jelas, daripada gagal di tengah setelah sebagian
// issue sudah terlanjur dibuat.
func CheckGHAvailable(ctx context.Context) error {
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("binary 'gh' (GitHub CLI) tidak ditemukan di PATH. Install dulu: https://cli.github.com")
	}
	cmd := exec.CommandContext(ctx, "gh", "auth", "status")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh belum login (jalankan 'gh auth login' dulu): %s", strings.TrimSpace(stderr.String()))
	}
	return nil
}
