package ui

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"termi.sh/termi/internal/llm"
	"termi.sh/termi/internal/runner"
	"termi.sh/termi/internal/suggest"
)

// AppState represents the different states of the application
type AppState int

const (
	StateInit AppState = iota
	StateAnalyzing
	StateAsking
	StateSelecting
	StateExecuting
	StateCompleted
	StateError
	StateCanceled
	StateCopied
)

// AppModel is the main application model that handles the entire flow
type AppModel struct {
	state         AppState
	query         string
	originalQuery string
	candidates    []suggest.Suggestion
	cursor        int
	spinner       spinner.Model
	err           error

	// For user input state
	inputPrompt string
	textInput   textinput.Model

	// Context for conversation with LLM
	contextHistory []string

	// Execution related
	selectedCommand string
	copiedCommand   string

	// Styles
	titleStyle    lipgloss.Style
	itemStyle     lipgloss.Style
	selectedStyle lipgloss.Style
	errorStyle    lipgloss.Style
	successStyle  lipgloss.Style
}

// NewAppModel creates a new application model
func NewAppModel(query string) *AppModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))

	// Initialize text input
	ti := textinput.New()

	return &AppModel{
		state:         StateInit,
		query:         query,
		originalQuery: query,
		spinner:       s,
		textInput:     ti,
		titleStyle:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99")),
		itemStyle:     lipgloss.NewStyle(),
		selectedStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true),
		errorStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		successStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("46")),
	}
}

// RunApp starts the main application flow
func RunApp(query string) error {
	m := NewAppModel(query)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("界面运行出错: %w", err)
	}

	// Check if we need to execute a command after TUI exit
	if appModel, ok := finalModel.(*AppModel); ok {
		switch appModel.state {
		case StateCompleted:
			if appModel.selectedCommand != "" {
				fmt.Printf("\n执行命令: %s\n\n", appModel.selectedCommand)
				if execErr := runner.Run(appModel.selectedCommand); execErr != nil {
					return fmt.Errorf("命令执行失败: %w", execErr)
				}
			}
		case StateCopied:
			if appModel.copiedCommand != "" {
				fmt.Printf("📋 已复制到剪贴板: \n  %s\n", appModel.copiedCommand)
			}
		case StateError:
			return fmt.Errorf("应用错误: %w", appModel.err)
		case StateCanceled:
			fmt.Println("操作已取消")
			return nil
		}
	}

	return nil
}

// Message types for AppModel
type llmAnalysisMsg struct {
	command string
	ask     string
	err     error
}

type copiedMsg struct {
	success bool
	err     error
}

// copyToClipboard copies text to the system clipboard
func copyToClipboard(text string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return fmt.Errorf("no clipboard utility found (install xclip or xsel)")
		}
	case "windows":
		cmd = exec.Command("clip")
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// Init initializes the AppModel
func (m *AppModel) Init() tea.Cmd {
	if !llm.Enabled() {
		m.state = StateError
		m.err = fmt.Errorf("LLM 未启用，请设置 OPENAI_API_KEY 环境变量")
		return nil
	}

	m.state = StateAnalyzing
	return tea.Batch(
		m.spinner.Tick,
		m.analyzeLLMCmd(),
	)
}

// Update handles messages and state transitions
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Update textinput when in asking state
	if m.state == StateAsking {
		m.textInput, cmd = m.textInput.Update(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case spinner.TickMsg:
		var spinnerCmd tea.Cmd
		m.spinner, spinnerCmd = m.spinner.Update(msg)
		return m, tea.Batch(cmd, spinnerCmd)
	case llmAnalysisMsg:
		return m.handleLLMAnalysis(msg)
	case copiedMsg:
		return m.handleCopied(msg)
	}
	return m, cmd
}

// View renders the current state
func (m *AppModel) View() string {
	switch m.state {
	case StateInit:
		return m.titleStyle.Render("🚀 Termi") + "\n\n" +
			m.spinner.View() + " 初始化中..."
	case StateAnalyzing:
		return m.titleStyle.Render("🧠 分析中") + "\n\n" +
			m.spinner.View() + " 正在分析您的需求: " +
			lipgloss.NewStyle().Italic(true).Render(m.query) + "\n\n" +
			lipgloss.NewStyle().Faint(true).Render("请稍候...")
	case StateAsking:
		return m.renderAskingView()
	case StateSelecting:
		return m.renderSelectingView()
	case StateExecuting:
		return m.titleStyle.Render("⚡ 执行中") + "\n\n" +
			m.spinner.View() + " 正在执行命令...\n\n" +
			lipgloss.NewStyle().Faint(true).Render("请稍候...")
	case StateCompleted:
		return m.successStyle.Render("✅ 准备执行命令")
	case StateError:
		return m.titleStyle.Render("❌ 错误") + "\n\n" +
			m.errorStyle.Render(fmt.Sprintf("发生错误: %v", m.err)) + "\n\n" +
			lipgloss.NewStyle().Faint(true).Render("按 q 退出")
	case StateCanceled:
		return m.titleStyle.Render("🚫 已取消") + "\n\n" +
			lipgloss.NewStyle().Faint(true).Render("操作已取消")
	default:
		return m.errorStyle.Render("未知状态")
	}
}

// Helper methods
func (m *AppModel) analyzeLLMCmd() tea.Cmd {
	return func() tea.Msg {
		// Build full context with history
		fullQuery := m.query
		if len(m.contextHistory) > 0 {
			fullQuery = strings.Join(m.contextHistory, " ") + " " + m.query
		}

		cmd, ask, err := llm.AskSmart(fullQuery)
		return llmAnalysisMsg{
			command: cmd,
			ask:     ask,
			err:     err,
		}
	}
}

func (m *AppModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case StateAsking:
		switch msg.Type {
		case tea.KeyEnter:
			input := strings.TrimSpace(m.textInput.Value())
			if input == "" {
				return m, nil
			}
			// Add question and answer to context history
			m.contextHistory = append(m.contextHistory, m.inputPrompt+" "+input)
			m.textInput.SetValue("")
			m.state = StateAnalyzing
			return m, tea.Batch(m.spinner.Tick, m.analyzeLLMCmd())
		case tea.KeyCtrlC, tea.KeyEsc:
			m.state = StateCanceled
			return m, tea.Quit
		}
	case StateSelecting:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.state = StateCanceled
			return m, tea.Quit
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor < len(m.candidates)-1 {
				m.cursor++
			}
		case tea.KeyEnter:
			return m.executeCommand()
		}
		// Additional vim-style navigation
		switch msg.String() {
		case "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "j":
			if m.cursor < len(m.candidates)-1 {
				m.cursor++
			}
		case "q":
			m.state = StateCanceled
			return m, tea.Quit
		case "c":
			return m.copyCommand()
		}
	default:
		if msg.Type == tea.KeyCtrlC || msg.String() == "q" {
			m.state = StateCanceled
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *AppModel) handleLLMAnalysis(msg llmAnalysisMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.state = StateError
		m.err = m.formatLLMError(msg.err)
		return m, nil
	}

	if msg.ask != "" {
		return m.transitionToAsking(msg.ask), nil
	}

	if msg.command != "" {
		return m.transitionToSelecting(msg.command), nil
	}

	m.state = StateError
	m.err = fmt.Errorf("LLM 未能生成可执行命令，请尝试提供更详细的描述")
	return m, nil
}

func (m *AppModel) formatLLMError(err error) error {
	var llmErr *llm.LLMError
	if errors.As(err, &llmErr) {
		switch llmErr.Type {
		case llm.ErrorTypeAuth:
			return fmt.Errorf("请设置对应的 API KEY 环境变量")
		case llm.ErrorTypeTimeout:
			return fmt.Errorf("网络请求超时，请检查网络连接")
		case llm.ErrorTypeQuota:
			return fmt.Errorf("API 配额已用完，请检查账户")
		case llm.ErrorTypeNetwork:
			return fmt.Errorf("网络连接失败，请检查连接")
		default:
			return fmt.Errorf("LLM 服务出错: %v", llmErr.Message)
		}
	}

	// 向后兼容，处理非 LLMError 类型
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "API KEY"):
		return fmt.Errorf("请设置对应的 API KEY 环境变量")
	case strings.Contains(errStr, "timeout"):
		return fmt.Errorf("网络请求超时，请检查网络连接")
	case strings.Contains(errStr, "quota"):
		return fmt.Errorf("API 配额已用完，请检查账户")
	default:
		return fmt.Errorf("LLM 服务出错: %v", err)
	}
}

func (m *AppModel) transitionToAsking(ask string) *AppModel {
	m.state = StateAsking
	m.inputPrompt = ask
	m.textInput.SetValue("")
	m.textInput.Focus()
	return m
}

func (m *AppModel) transitionToSelecting(command string) *AppModel {
	m.candidates = []suggest.Suggestion{{Text: command, Source: "llm"}}
	m.state = StateSelecting
	return m
}

func (m *AppModel) executeCommand() (tea.Model, tea.Cmd) {
	if m.cursor >= len(m.candidates) {
		return m, nil
	}

	choice := m.candidates[m.cursor]
	m.selectedCommand = choice.Text
	m.state = StateCompleted

	// Exit the TUI - command will be executed in RunApp
	return m, tea.Quit
}

func (m *AppModel) renderAskingView() string {
	var s strings.Builder

	// Show original query
	s.WriteString(m.titleStyle.Render("🎯 原始需求: "))
	s.WriteString(lipgloss.NewStyle().Italic(true).Render(m.originalQuery))
	s.WriteString("\n\n")

	// Show conversation history if any
	if len(m.contextHistory) > 0 {
		s.WriteString(lipgloss.NewStyle().Faint(true).Render("对话历史:"))
		s.WriteString("\n")
		for i, ctx := range m.contextHistory {
			s.WriteString(lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("%d. %s", i+1, ctx)))
			s.WriteString("\n")
		}
		s.WriteString("\n")
	}

	// Current question
	prompt := m.titleStyle.Render("❓ ") + m.inputPrompt
	s.WriteString(prompt)
	s.WriteString("\n\n")

	// Input line using textinput component
	s.WriteString(m.textInput.View())
	s.WriteString("\n\n")

	// Help text
	helpText := lipgloss.NewStyle().
		Faint(true).
		Render("Enter: 提交, Ctrl+C/Esc: 取消")
	s.WriteString(helpText)

	return s.String()
}

func (m *AppModel) renderSelectingView() string {
	if len(m.candidates) == 0 {
		return m.errorStyle.Render("❌ 没有可执行的候选命令。")
	}

	var s strings.Builder

	// Title
	title := m.titleStyle.Render("🚀 选择要执行的命令:")
	s.WriteString(title + "\n\n")

	// Command list
	for i, item := range m.candidates {
		var line string
		if m.cursor == i {
			// Selected item
			cursor := m.selectedStyle.Render("➜ ")
			cmdText := m.selectedStyle.Render(item.Text)
			source := lipgloss.NewStyle().
				Faint(true).
				Foreground(lipgloss.Color("8")).
				Render(fmt.Sprintf("[%s]", item.Source))
			line = cursor + cmdText + " " + source
		} else {
			// Unselected item
			cursor := "  "
			cmdText := m.itemStyle.Render(item.Text)
			source := lipgloss.NewStyle().
				Faint(true).
				Foreground(lipgloss.Color("8")).
				Render(fmt.Sprintf("[%s]", item.Source))
			line = cursor + cmdText + " " + source
		}
		s.WriteString(line + "\n")
	}

	// Help text
	helpText := lipgloss.NewStyle().
		Faint(true).
		Render("\n↑/↓ 或 k/j: 选择, Enter: 执行, c: 复制, q/Esc: 退出")
	s.WriteString(helpText)

	return s.String()
}

func (m *AppModel) copyCommand() (tea.Model, tea.Cmd) {
	if m.cursor >= len(m.candidates) {
		return m, nil
	}

	choice := m.candidates[m.cursor]
	m.copiedCommand = choice.Text

	return m, func() tea.Msg {
		err := copyToClipboard(choice.Text)
		return copiedMsg{
			success: err == nil,
			err:     err,
		}
	}
}

func (m *AppModel) handleCopied(msg copiedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.state = StateError
		m.err = fmt.Errorf("复制失败: %v", msg.err)
		return m, nil
	}

	// Copy successful, set state and quit
	m.state = StateCopied
	return m, tea.Quit
}
