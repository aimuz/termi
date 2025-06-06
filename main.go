package main

import (
	"fmt"
	"os"
	"strings"

	"termi.sh/termi/internal/config"
	"termi.sh/termi/internal/llm"
	"termi.sh/termi/internal/ui"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return showUsage()
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		showConfigHelp(err)
		return err
	}

	if err := llm.Initialize(cfg); err != nil {
		return fmt.Errorf("初始化 LLM 提供商失败: %w", err)
	}

	query := strings.Join(os.Args[1:], " ")
	return ui.RunApp(query)
}

func showUsage() error {
	fmt.Println("请在命令后输入自然语言，例如：\n  termi 我想对 baidu.com 发起 ping")
	return nil
}

func showConfigHelp(err error) {
	fmt.Printf("加载配置失败: %v\n", err)
	fmt.Println("\n请设置以下环境变量之一：")
	fmt.Println("  OPENAI_API_KEY - 使用 OpenAI")
	fmt.Println("  AZURE_OPENAI_API_KEY - 使用 Azure OpenAI")
	fmt.Println("  GEMINI_API_KEY - 使用 Google Gemini")
	fmt.Println("  ANTHROPIC_API_KEY - 使用 Anthropic Claude")
	fmt.Println("  LLAMA_CPP_BASE_URL - 使用 Llama.cpp 服务")
	fmt.Println("\n或创建配置文件: ~/.config/termi/config.json")
}
