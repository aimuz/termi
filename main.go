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
	if len(os.Args) < 2 {
		fmt.Println("请在命令后输入自然语言，例如：\n  termi 我想对 baidu.com 发起 ping")
		os.Exit(1)
	}

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		fmt.Println("\n请设置以下环境变量之一：")
		fmt.Println("  OPENAI_API_KEY - 使用 OpenAI")
		fmt.Println("  AZURE_OPENAI_API_KEY - 使用 Azure OpenAI")
		fmt.Println("  GEMINI_API_KEY - 使用 Google Gemini")
		fmt.Println("  ANTHROPIC_API_KEY - 使用 Anthropic Claude")
		fmt.Println("  LLAMA_CPP_BASE_URL - 使用 Llama.cpp 服务")
		fmt.Println("\n或创建配置文件: ~/.config/termi/config.json")
		os.Exit(1)
	}

	// 初始化 LLM 提供商
	if err := llm.Initialize(cfg); err != nil {
		fmt.Printf("初始化 LLM 提供商失败: %v\n", err)
		os.Exit(1)
	}

	query := strings.Join(os.Args[1:], " ")

	if err := ui.RunApp(query); err != nil {
		fmt.Printf("应用出错: %v\n", err)
		os.Exit(1)
	}
}
