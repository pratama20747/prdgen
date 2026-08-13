// Package ghissues mengubah daftar issue (data terstruktur hasil LLM) jadi
// GitHub issue sungguhan, lewat GitHub CLI (`gh`).
//
// PRINSIP KEAMANAN UTAMA: LLM tidak pernah menghasilkan shell command. LLM
// cuma menghasilkan JSON (title/body/labels) yang di-parse jadi struct Issue
// di sini. Yang benar-benar menjalankan `gh` adalah GHCLIExecutor lewat
// exec.Command dengan argumen terpisah (bukan lewat shell string) -- jadi
// apapun isi title/body (termasuk karakter shell seperti ; ` $() &&) selalu
// diperlakukan sebagai data murni oleh proses `gh`, tidak pernah oleh shell
// manapun. Ini menutup risiko command injection meski konten issue
// sepenuhnya dihasilkan oleh model yang tidak sepenuhnya bisa dipercaya.
package ghissues

// Issue merepresentasikan satu GitHub issue yang akan dibuat.
// Field ini murni data -- tidak ada satupun yang dieksekusi sebagai command.
type Issue struct {
	Title  string   `json:"title"`
	Body   string   `json:"body"`
	Labels []string `json:"labels"`
	Phase  string   `json:"phase"`
}
