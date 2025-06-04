package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

var client *openai.Client

func init() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return
	}
	client = openai.NewClient(apiKey)
}

// Enabled 返回是否已正确配置 API KEY
func Enabled() bool { return client != nil }

// Ask 将用户需求 prompt 给 OpenAI，并返回模型生成的答案
func Ask(prompt string) (string, error) {
	if client == nil {
		return "", fmt.Errorf("OpenAI API KEY 未配置 (环境变量 OPENAI_API_KEY)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: "你是一个 Linux 命令行专家，请仅输出可直接粘贴执行的 bash 命令，不要添加其他解释。"},
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature: 0.2,
	})
	if err != nil {
		return "", err
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	// 去掉可能的代码块标记
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	return strings.TrimSpace(content), nil
}

// AskCommand 使用 OpenAI 的 response_format=json_object 返回 {"command": "..."}
func AskCommand(prompt string) (string, error) {
	if client == nil {
		return "", fmt.Errorf("OpenAI API KEY 未配置 (环境变量 OPENAI_API_KEY)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: `你是 Linux 命令行专家，请根据用户需求输出 JSON，格式: {"command":"..."}，不需要其他字段。`},
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature:    0.1,
		ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
	})
	if err != nil {
		return "", err
	}

	var out struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &out); err != nil {
		// fallback: 返回原始文本
		return strings.TrimSpace(resp.Choices[0].Message.Content), nil
	}
	return strings.TrimSpace(out.Command), nil
}

// AskSmart 根据用户 query 返回 command 或 ask
// 如果需要更多信息，则 ask 字段非空
func AskSmart(prompt string) (command string, ask string, err error) {
	if client == nil {
		return "", "", fmt.Errorf("OpenAI API KEY 未配置")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: `你是 Linux 命令行专家。如果已足够信息，请返回 JSON {"command":"..."}，其中 command 是可直接执行的 Bash 命令；如果信息不足，请返回 JSON {"ask":"..."}，ask 用中文向用户提出你需要的补充问题，不需要其他字段。`},
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature:    0.2,
		ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
	})
	if err != nil {
		return "", "", err
	}
	var out struct {
		Command string `json:"command"`
		Ask     string `json:"ask"`
	}
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &out); err != nil {
		return "", "", err
	}
	return strings.TrimSpace(out.Command), strings.TrimSpace(out.Ask), nil
}
