package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LLMProvider 定义支持的 LLM 提供商类型
type LLMProvider string

const (
	ProviderOpenAI      LLMProvider = "openai"
	ProviderAzureOpenAI LLMProvider = "azure-openai"
	ProviderGemini      LLMProvider = "gemini"
	ProviderClaude      LLMProvider = "claude"
	ProviderLlamaCPP    LLMProvider = "llama-cpp"
)

// LLMConfig LLM 配置结构
type LLMConfig struct {
	Provider LLMProvider `json:"provider"`

	// OpenAI 配置
	OpenAI *OpenAIConfig `json:"openai,omitempty"`

	// Azure OpenAI 配置
	AzureOpenAI *AzureOpenAIConfig `json:"azure_openai,omitempty"`

	// Gemini 配置
	Gemini *GeminiConfig `json:"gemini,omitempty"`

	// Claude 配置
	Claude *ClaudeConfig `json:"claude,omitempty"`

	// Llama-cpp 配置
	LlamaCPP *LlamaCPPConfig `json:"llama_cpp,omitempty"`
}

// OpenAIConfig OpenAI 配置
type OpenAIConfig struct {
	APIKey  string `json:"api_key"`
	Model   string `json:"model"`
	BaseURL string `json:"base_url,omitempty"`
	OrgID   string `json:"org_id,omitempty"`
	Timeout int    `json:"timeout,omitempty"` // 秒
}

// AzureOpenAIConfig Azure OpenAI 配置
type AzureOpenAIConfig struct {
	APIKey       string `json:"api_key"`
	BaseURL      string `json:"base_url"`
	DeploymentID string `json:"deployment_id"`
	APIVersion   string `json:"api_version"`
	Timeout      int    `json:"timeout,omitempty"` // 秒
}

// GeminiConfig Gemini 配置
type GeminiConfig struct {
	APIKey  string `json:"api_key"`
	Model   string `json:"model"`
	BaseURL string `json:"base_url,omitempty"`
	Timeout int    `json:"timeout,omitempty"` // 秒
}

// ClaudeConfig Claude 配置
type ClaudeConfig struct {
	APIKey  string `json:"api_key"`
	Model   string `json:"model"`
	BaseURL string `json:"base_url,omitempty"`
	Timeout int    `json:"timeout,omitempty"` // 秒
}

// LlamaCPPConfig Llama-cpp 配置
type LlamaCPPConfig struct {
	BaseURL string `json:"base_url"`
	Model   string `json:"model,omitempty"`
	Timeout int    `json:"timeout,omitempty"` // 秒
}

// Config 应用配置
type Config struct {
	LLM LLMConfig `json:"llm"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		LLM: LLMConfig{
			Provider: ProviderOpenAI,
			OpenAI: &OpenAIConfig{
				Model:   "gpt-3.5-turbo",
				Timeout: 30,
			},
		},
	}
}

// LoadConfig 从文件加载配置，如果文件不存在则从环境变量加载
func LoadConfig() (*Config, error) {
	// 首先尝试从配置文件加载
	configPath := getConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		return loadFromFile(configPath)
	}

	// 如果配置文件不存在，从环境变量加载
	return loadFromEnv()
}

// SaveConfig 保存配置到文件
func (c *Config) SaveConfig() error {
	configPath := getConfigPath()

	// 确保配置目录存在
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// getConfigPath 获取配置文件路径
func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./termi.json"
	}
	return filepath.Join(homeDir, ".config", "termi", "config.json")
}

// loadFromFile 从文件加载配置
func loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

// loadFromEnv 从环境变量加载配置
func loadFromEnv() (*Config, error) {
	config := DefaultConfig()

	// 检查 OpenAI 配置
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		config.LLM.Provider = ProviderOpenAI
		config.LLM.OpenAI.APIKey = apiKey
		if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" {
			config.LLM.OpenAI.BaseURL = baseURL
		}
		if orgID := os.Getenv("OPENAI_ORG_ID"); orgID != "" {
			config.LLM.OpenAI.OrgID = orgID
		}
		return config, nil
	}

	// 检查 Azure OpenAI 配置
	if apiKey := os.Getenv("AZURE_OPENAI_API_KEY"); apiKey != "" {
		config.LLM.Provider = ProviderAzureOpenAI
		config.LLM.AzureOpenAI = &AzureOpenAIConfig{
			APIKey:       apiKey,
			BaseURL:      os.Getenv("AZURE_OPENAI_BASE_URL"),
			DeploymentID: os.Getenv("AZURE_OPENAI_DEPLOYMENT_ID"),
			APIVersion:   getEnvOrDefault("AZURE_OPENAI_API_VERSION", "2023-12-01-preview"),
			Timeout:      30,
		}
		return config, nil
	}

	// 检查 Gemini 配置
	if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
		config.LLM.Provider = ProviderGemini
		config.LLM.Gemini = &GeminiConfig{
			APIKey:  apiKey,
			Model:   getEnvOrDefault("GEMINI_MODEL", "gemini-pro"),
			BaseURL: os.Getenv("GEMINI_BASE_URL"),
			Timeout: 30,
		}
		return config, nil
	}

	// 检查 Claude 配置
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.LLM.Provider = ProviderClaude
		config.LLM.Claude = &ClaudeConfig{
			APIKey:  apiKey,
			Model:   getEnvOrDefault("CLAUDE_MODEL", "claude-3-haiku-20240307"),
			BaseURL: os.Getenv("ANTHROPIC_BASE_URL"),
			Timeout: 30,
		}
		return config, nil
	}

	// 检查 Llama-cpp 配置
	if baseURL := os.Getenv("LLAMA_CPP_BASE_URL"); baseURL != "" {
		config.LLM.Provider = ProviderLlamaCPP
		config.LLM.LlamaCPP = &LlamaCPPConfig{
			BaseURL: baseURL,
			Model:   os.Getenv("LLAMA_CPP_MODEL"),
			Timeout: 30,
		}
		return config, nil
	}

	return nil, fmt.Errorf("未找到任何 LLM 提供商配置")
}

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
