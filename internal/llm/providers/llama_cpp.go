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

// LlamaCPPProvider Llama-cpp 提供商实现
type LlamaCPPProvider struct {
	httpClient *http.Client
	config     *config.LlamaCPPConfig
}

// NewLlamaCPPProvider 创建 Llama-cpp 提供商
func NewLlamaCPPProvider(cfg *config.LlamaCPPConfig) (*LlamaCPPProvider, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("Llama-cpp Base URL 未配置")
	}
	
	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	
	return &LlamaCPPProvider{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		config: cfg,
	}, nil
}

// Name 返回提供商名称
func (p *LlamaCPPProvider) Name() string {
	return "Llama-cpp"
}

// Enabled 返回是否已正确配置
func (p *LlamaCPPProvider) Enabled() bool {
	return p.httpClient != nil && p.config.BaseURL != ""
}

// AskSmart 根据用户 query 返回 command 或 ask
func (p *LlamaCPPProvider) AskSmart(ctx context.Context, prompt string) (command string, ask string, err error) {
	timeout := time.Duration(p.config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	// 构建请求
	url := fmt.Sprintf("%s/completion", strings.TrimSuffix(p.config.BaseURL, "/"))
	
	fullPrompt := fmt.Sprintf(`你是 Linux 命令行专家。根据用户需求和对话历史，生成合适的 Bash 命令。

如果信息充足，返回 JSON {"command":"..."}，其中 command 是可直接执行的 Bash 命令。
如果需要更多信息，返回 JSON {"ask":"..."}，ask 用中文向用户提出具体的补充问题。

注意：
- 仔细理解用户的完整意图和上下文
- 如果之前的对话中已经提供了相关信息，请充分利用
- 生成的命令应该是安全、准确且可执行的

用户需求: %s

请直接返回JSON格式的响应：`, prompt)
	
	reqBody := map[string]interface{}{
		"prompt":      fullPrompt,
		"max_tokens":  1000,
		"temperature": 0.2,
		"top_p":       0.8,
		"stop":        []string{"<|im_end|>", "\n\n"},
		"stream":      false,
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
		return "", "", fmt.Errorf("Llama-cpp API 调用失败: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("Llama-cpp API 返回错误状态: %d", resp.StatusCode)
	}
	
	var llamaResp struct {
		Content string `json:"content"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&llamaResp); err != nil {
		return "", "", fmt.Errorf("解析 Llama-cpp 响应失败: %w", err)
	}
	
	responseText := strings.TrimSpace(llamaResp.Content)
	if responseText == "" {
		return "", "", fmt.Errorf("Llama-cpp API 返回空文本")
	}
	
	// 解析 JSON 响应
	var out struct {
		Command string `json:"command"`
		Ask     string `json:"ask"`
	}
	if err := json.Unmarshal([]byte(responseText), &out); err != nil {
		return "", "", fmt.Errorf("解析 Llama-cpp 响应失败: %w, 原始响应: %s", err, responseText)
	}
	
	return strings.TrimSpace(out.Command), strings.TrimSpace(out.Ask), nil
}