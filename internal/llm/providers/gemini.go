package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"termi.sh/termi/internal/config"
)

// GeminiProvider Gemini 提供商实现
type GeminiProvider struct {
	httpClient *http.Client
	config     *config.GeminiConfig
}

// NewGeminiProvider 创建 Gemini 提供商
func NewGeminiProvider(cfg *config.GeminiConfig) (*GeminiProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("Gemini API Key 未配置")
	}

	return &GeminiProvider{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: cfg,
	}, nil
}

// Name 返回提供商名称
func (p *GeminiProvider) Name() string {
	return "Gemini"
}

// Enabled 返回是否已正确配置
func (p *GeminiProvider) Enabled() bool {
	return p.httpClient != nil && p.config.APIKey != ""
}

// AskSmart 根据用户 query 返回 command 或 ask
func (p *GeminiProvider) AskSmart(ctx context.Context, prompt string) (command string, ask string, err error) {
	timeout := time.Duration(p.config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	model := p.config.Model
	if model == "" {
		model = "gemini-pro"
	}

	// 构建请求
	baseURL := "https://generativelanguage.googleapis.com"
	if p.config.BaseURL != "" {
		baseURL = p.config.BaseURL
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", baseURL, model, p.config.APIKey)

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"text": fmt.Sprintf(`你是 Linux 命令行专家。根据用户需求和对话历史，生成合适的 Bash 命令。

如果信息充足，返回 JSON {"command":"..."}，其中 command 是可直接执行的 Bash 命令。
如果需要更多信息，返回 JSON {"ask":"..."}，ask 用中文向用户提出具体的补充问题。

注意：
- 仔细理解用户的完整意图和上下文
- 如果之前的对话中已经提供了相关信息，请充分利用
- 生成的命令应该是安全、准确且可执行的

用户需求: %s`, prompt),
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.2,
			"topP":            0.8,
			"maxOutputTokens": 1000,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", fmt.Errorf("构建请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("Gemini API 调用失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("Gemini API 返回错误状态: %d", resp.StatusCode)
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", "", fmt.Errorf("解析 Gemini 响应失败: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", "", fmt.Errorf("Gemini API 返回空结果")
	}

	responseText := geminiResp.Candidates[0].Content.Parts[0].Text

	if responseText == "" {
		return "", "", fmt.Errorf("Gemini API 返回空文本")
	}

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
