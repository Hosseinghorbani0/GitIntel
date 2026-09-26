package llm

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func LoadConfigFromEnv() (Config, error) {
	provider := strings.TrimSpace(os.Getenv("LLM_PROVIDER"))
	if provider == "" {
		provider = "huggingface"
	}
	baseURL := strings.TrimSpace(os.Getenv("LLM_BASE_URL"))
	if baseURL == "" {
		baseURL = strings.TrimSpace(os.Getenv("HF_BASE_URL"))
	}
	if baseURL == "" {
		if strings.EqualFold(provider, "huggingface") {
			baseURL = "https://router.huggingface.co/v1"
		}
	}
	model := strings.TrimSpace(os.Getenv("LLM_MODEL"))
	if model == "" {
		model = strings.TrimSpace(os.Getenv("HF_MODEL"))
	}
	if model == "" {
		if strings.EqualFold(provider, "huggingface") {
			model = "Qwen/Qwen3.8-2.4T-A95B:novita"
		}
	}
	credential := strings.TrimSpace(os.Getenv("LLM_API_KEY"))
	if credential == "" {
		credential = strings.TrimSpace(os.Getenv("HF_API_KEY"))
	}
	tempValue := strings.TrimSpace(os.Getenv("LLM_TEMPERATURE"))
	temperature := 0.2
	if tempValue != "" {
		parsed, err := strconv.ParseFloat(tempValue, 64)
		if err == nil {
			temperature = parsed
		}
	}
	maxTokens := 1200
	if raw := strings.TrimSpace(os.Getenv("LLM_MAX_TOKENS")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err == nil && parsed > 0 {
			maxTokens = parsed
		}
	}
	timeout := 30
	if raw := strings.TrimSpace(os.Getenv("LLM_TIMEOUT")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err == nil && parsed > 0 {
			timeout = parsed
		}
	}
	return Config{
		Provider:    provider,
		Model:       model,
		BaseURL:     baseURL,
		Credential:  credential,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Timeout:     timeout,
	}, nil
}

func (c Config) Validate() error {
	if c.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if c.BaseURL == "" {
		return fmt.Errorf("base_url is required")
	}
	if c.Model == "" {
		return fmt.Errorf("model is required")
	}
	return nil
}
