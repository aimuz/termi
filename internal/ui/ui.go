package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"termi.sh/termi/internal/llm"
	"termi.sh/termi/internal/suggest"
)

// RunWithLLM 启动 TUI，初始展示 list，并根据 query 异步向 LLM 请求额外建议。
// 返回用户选择的索引，取消返回 -1
func RunWithLLM(query string, list []suggest.Suggestion) (int, error) {
	loading := llm.Enabled()

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot

	m := &model{
		items:   list,
		loading: loading,
		spinner: sp,
		query:   query,
	}

	var opts []tea.ProgramOption
	p := tea.NewProgram(m, opts...)
	if _, err := p.Run(); err != nil {
		return -1, err
	}
	if m.canceled {
		return -1, nil
	}
	return m.cursor, nil
}

// RunSimple 仅根据已有 list 交互选择，不再触发 LLM。
func RunSimple(list []suggest.Suggestion) (int, error) {
	m := &model{items: list}
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return -1, err
	}
	if m.canceled {
		return -1, nil
	}
	return m.cursor, nil
}

type model struct {
	items    []suggest.Suggestion
	cursor   int
	canceled bool
	loading  bool
	spinner  spinner.Model
	query    string
}

func (m *model) Init() tea.Cmd {
	if m.loading {
		return tea.Batch(m.spinner.Tick, fetchLLMCmd(m.query))
	}
	return nil
}

type llmSuggestionMsg struct{ suggestion string }
type llmErrMsg struct{ err error }

func fetchLLMCmd(query string) tea.Cmd {
	return func() tea.Msg {
		if ans, err := llm.AskCommand(query); err == nil {
			return llmSuggestionMsg{ans}
		} else {
			return llmErrMsg{err}
		}
	}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.loading {
		// 交给 spinner 处理 Tick 消息
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		switch v := msg.(type) {
		case llmSuggestionMsg:
			if v.suggestion != "" {
				exists := false
				for _, it := range m.items {
					if it.Text == v.suggestion {
						exists = true
						break
					}
				}
				if !exists {
					m.items = append(m.items, suggest.Suggestion{Text: v.suggestion, Source: "llm"})
				}
			}
			m.loading = false
			return m, nil
		case llmErrMsg:
			m.loading = false
			return m, nil
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.canceled = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *model) View() string {
	if m.loading {
		return fmt.Sprintf("%s LLM 生成建议中... 按 q 取消\n", m.spinner.View())
	}

	if len(m.items) == 0 {
		return "没有可展示的候选\n"
	}
	s := "候选命令 (↑/↓ 选择, Enter 确定, q 退出):\n\n"
	for i, item := range m.items {
		cursor := " "
		if m.cursor == i {
			cursor = "➜"
		}
		s += fmt.Sprintf("%s %d. %s [%s]\n", cursor, i+1, item.Text, item.Source)
	}
	return s
}
