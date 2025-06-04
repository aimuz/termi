package suggest

// Suggestion 表示一条候选命令
type Suggestion struct {
	Text   string // 真实命令
	Source string // 例如 llm
}
