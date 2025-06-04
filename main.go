package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"termi.sh/termi/internal/llm"
	"termi.sh/termi/internal/runner"
	"termi.sh/termi/internal/suggest"
	"termi.sh/termi/internal/ui"
)

// loadingModel 是一个简单的 loading 模型
type loadingModel struct {
	spinner  spinner.Model
	quitting bool
	done     bool
}

func (m loadingModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m loadingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.QuitMsg:
		m.done = true
		return m, nil
	}
	return m, nil
}

func (m loadingModel) View() string {
	if m.quitting || m.done {
		// 清理 loading 输出
		return "\r\033[K"
	}
	return fmt.Sprintf("\r %s Thinking...", m.spinner.View())
}

// runWithLoading 在后台运行函数时显示 loading
func runWithLoading(fn func()) {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))

	m := loadingModel{spinner: s}

	// 启动 loading UI
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))

	done := make(chan bool)
	go func() {
		fn()
		done <- true
	}()

	go func() {
		<-done
		p.Send(tea.Quit())
	}()

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
	// 确保清理输出
	fmt.Fprint(os.Stderr, "\r\033[K")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("请在命令后输入自然语言，例如：\n  termi 我想对 baidu.com 发起 ping")
		os.Exit(1)
	}

	query := strings.Join(os.Args[1:], " ")

	// 0. 如果启用 LLM，先让 LLM 判断是否需要补充信息
	var llmCmd string
	if llm.Enabled() {
		current := query
		for range 3 {
			var cmd, ask string
			var err error

			// 在 loading 状态下调用 LLM
			runWithLoading(func() {
				cmd, ask, err = llm.AskSmart(current)
			})

			if err != nil {
				break
			}
			if ask != "" {
				fmt.Print(ask + " ")
				reader := bufio.NewReader(os.Stdin)
				extra, _ := reader.ReadString('\n')
				extra = strings.TrimSpace(extra)
				current = current + " " + extra
				continue
			}
			llmCmd = cmd
			break
		}
	}

	// 1. 仅使用 LLM 候选
	var candidates []suggest.Suggestion
	if llmCmd != "" {
		candidates = append(candidates, suggest.Suggestion{Text: llmCmd, Source: "llm"})
	} else {
		fmt.Println("LLM 未能生成可执行命令。")
		os.Exit(1)
	}

	// 2. 交互式选择（如果候选只有一条，可直接执行，留给后续优化）
	idx, err := ui.RunSimple(candidates)
	if err != nil {
		fmt.Println("UI 发生错误:", err)
		os.Exit(1)
	}
	if idx == -1 {
		fmt.Println("已取消。")
		return
	}

	choice := candidates[idx]
	fmt.Printf("\n执行命令: %s\n\n", choice.Text)

	// 直接执行命令
	if err := runner.Run(choice.Text); err != nil {
		fmt.Println("执行出错:", err)
	}
}
