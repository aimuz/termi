package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"termi.sh/termi/internal/llm"
	"termi.sh/termi/internal/runner"
	"termi.sh/termi/internal/suggest"
	"termi.sh/termi/internal/ui"
)

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
			cmd, ask, err := llm.AskSmart(current)
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
