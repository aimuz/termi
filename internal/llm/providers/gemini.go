package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"google.golang.org/genai"

	"termi.sh/termi/internal/config"
)

// GeminiProvider Gemini 提供商实现
type GeminiProvider struct {
	client *genai.Client
	config *config.GeminiConfig
}

// NewGeminiProvider 创建 Gemini 提供商
func NewGeminiProvider(cfg *config.GeminiConfig) (*GeminiProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("Gemini API Key 未配置")
	}

	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  cfg.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 Gemini 客户端失败: %w", err)
	}

	return &GeminiProvider{
		client: client,
		config: cfg,
	}, nil
}

// Name 返回提供商名称
func (p *GeminiProvider) Name() string {
	return "Gemini"
}

// Enabled 返回是否已正确配置
func (p *GeminiProvider) Enabled() bool {
	return p.client != nil && p.config.APIKey != ""
}

// AskSmart 根据用户 query 返回 command 或 ask
func (p *GeminiProvider) AskSmart(ctx context.Context, prompt string) (command string, ask string, err error) {
	timeout := time.Duration(p.config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	chat, err := p.client.Chats.Create(ctx, p.config.Model, &genai.GenerateContentConfig{
		Temperature: genai.Ptr[float32](0.2),
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{Text: systemPrompt()},
			},
			Role: "system",
		}}, nil)
	if err != nil {
		return "", "", fmt.Errorf("创建 Gemini 聊天失败: %w", err)
	}

	result, err := chat.SendMessage(ctx, genai.Part{Text: prompt})
	if err != nil {
		return "", "", fmt.Errorf("Gemini API 调用失败: %w", err)
	}

	responseText := result.Text()
	// 解析 JSON 响应
	var out struct {
		Command string `json:"command"`
		Ask     string `json:"ask"`
	}
	if err := json.Unmarshal([]byte(responseText), &out); err != nil {
		return "", "", fmt.Errorf("解析 Gemini 响应失败: %w, 原始响应: %s", err, responseText)
	}

	return strings.TrimSpace(out.Command), strings.TrimSpace(out.Ask), nil
}
