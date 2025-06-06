package config

import (
	"cmp"
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

// Validate 验证配置是否有效
func (c *Config) Validate() error {
	return c.LLM.Validate()
}

// Validate 验证 LLM 配置
func (lc *LLMConfig) Validate() error {
	switch lc.Provider {
	case ProviderOpenAI:
		if lc.OpenAI == nil {
			return fmt.Errorf("OpenAI 配置缺失")
		}
		return lc.OpenAI.Validate()
	case ProviderAzureOpenAI:
		if lc.AzureOpenAI == nil {
			return fmt.Errorf("Azure OpenAI 配置缺失")
		}
		return lc.AzureOpenAI.Validate()
	case ProviderGemini:
		if lc.Gemini == nil {
			return fmt.Errorf("Gemini 配置缺失")
		}
		return lc.Gemini.Validate()
	case ProviderClaude:
		if lc.Claude == nil {
			return fmt.Errorf("Claude 配置缺失")
		}
		return lc.Claude.Validate()
	case ProviderLlamaCPP:
		if lc.LlamaCPP == nil {
			return fmt.Errorf("Llama-cpp 配置缺失")
		}
		return lc.LlamaCPP.Validate()
	default:
		return fmt.Errorf("不支持的 LLM 提供商: %s", lc.Provider)
	}
}

// Validate 验证 OpenAI 配置
func (oc *OpenAIConfig) Validate() error {
	if oc.APIKey == "" {
		return fmt.Errorf("OpenAI API Key 不能为空")
	}
	if oc.Model == "" {
		return fmt.Errorf("OpenAI Model 不能为空")
	}
	return nil
}

// Validate 验证 Azure OpenAI 配置
func (ac *AzureOpenAIConfig) Validate() error {
	if ac.APIKey == "" {
		return fmt.Errorf("Azure OpenAI API Key 不能为空")
	}
	if ac.BaseURL == "" {
		return fmt.Errorf("Azure OpenAI Base URL 不能为空")
	}
	if ac.DeploymentID == "" {
		return fmt.Errorf("Azure OpenAI Deployment ID 不能为空")
	}
	return nil
}

// Validate 验证 Gemini 配置
func (gc *GeminiConfig) Validate() error {
	if gc.APIKey == "" {
		return fmt.Errorf("Gemini API Key 不能为空")
	}
	if gc.Model == "" {
		return fmt.Errorf("Gemini Model 不能为空")
	}
	return nil
}

// Validate 验证 Claude 配置
func (cc *ClaudeConfig) Validate() error {
	if cc.APIKey == "" {
		return fmt.Errorf("Claude API Key 不能为空")
	}
	if cc.Model == "" {
		return fmt.Errorf("Claude Model 不能为空")
	}
	return nil
}

// Validate 验证 Llama-cpp 配置
func (lc *LlamaCPPConfig) Validate() error {
	if lc.BaseURL == "" {
		return fmt.Errorf("Llama-cpp Base URL 不能为空")
	}
	return nil
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
	providers := []struct {
		name      LLMProvider
		envKey    string
		configure func(*Config, string) error
	}{
		{ProviderOpenAI, "OPENAI_API_KEY", configureOpenAI},
		{ProviderAzureOpenAI, "AZURE_OPENAI_API_KEY", configureAzureOpenAI},
		{ProviderGemini, "GEMINI_API_KEY", configureGemini},
		{ProviderClaude, "ANTHROPIC_API_KEY", configureClaude},
		{ProviderLlamaCPP, "LLAMA_CPP_BASE_URL", configureLlamaCPP},
	}

	config := DefaultConfig()

	for _, provider := range providers {
		if value := os.Getenv(provider.envKey); value != "" {
			config.LLM.Provider = provider.name
			if err := provider.configure(config, value); err != nil {
				return nil, fmt.Errorf("配置 %s 失败: %w", provider.name, err)
			}
			return config, nil
		}
	}

	return nil, fmt.Errorf("未找到任何 LLM 提供商配置")
}

func configureOpenAI(config *Config, apiKey string) error {
	config.LLM.OpenAI.APIKey = apiKey
	if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" {
		config.LLM.OpenAI.BaseURL = baseURL
	}
	if orgID := os.Getenv("OPENAI_ORG_ID"); orgID != "" {
		config.LLM.OpenAI.OrgID = orgID
	}
	return nil
}

func configureAzureOpenAI(config *Config, apiKey string) error {
	config.LLM.AzureOpenAI = &AzureOpenAIConfig{
		APIKey:       apiKey,
		BaseURL:      os.Getenv("AZURE_OPENAI_BASE_URL"),
		DeploymentID: os.Getenv("AZURE_OPENAI_DEPLOYMENT_ID"),
		APIVersion:   getEnvOrDefault("AZURE_OPENAI_API_VERSION", "2023-12-01-preview"),
		Timeout:      30,
	}
	return nil
}

func configureGemini(config *Config, apiKey string) error {
	config.LLM.Gemini = &GeminiConfig{
		APIKey:  apiKey,
		Model:   getEnvOrDefault("GEMINI_MODEL", "gemini-pro"),
		BaseURL: os.Getenv("GEMINI_BASE_URL"),
		Timeout: 30,
	}
	return nil
}

func configureClaude(config *Config, apiKey string) error {
	config.LLM.Claude = &ClaudeConfig{
		APIKey:  apiKey,
		Model:   getEnvOrDefault("CLAUDE_MODEL", "claude-3-haiku-20240307"),
		BaseURL: os.Getenv("ANTHROPIC_BASE_URL"),
		Timeout: 30,
	}
	return nil
}

func configureLlamaCPP(config *Config, baseURL string) error {
	config.LLM.LlamaCPP = &LlamaCPPConfig{
		BaseURL: baseURL,
		Model:   os.Getenv("LLAMA_CPP_MODEL"),
		Timeout: 30,
	}
	return nil
}

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	return cmp.Or(os.Getenv(key), defaultValue)
}
