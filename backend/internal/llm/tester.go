package llm

import (
	"context"
	"fmt"
	"log"
	"time"
)

type Status string

const (
	StatusReady           Status = "Ready"
	StatusTesting         Status = "Testing"
	StatusSuccess         Status = "Success"
	StatusPaymentRequired Status = "Payment Required"
	StatusRateLimited     Status = "Rate Limited"
	StatusInvalidCred     Status = "Invalid Credential"
	StatusError           Status = "Error"
)

type TestResult struct {
	Provider         string    `json:"provider"`
	Model            string    `json:"model"`
	Status           Status    `json:"status"`
	CredentialStatus string    `json:"credential_status"`
	Latency          int64     `json:"latency_ms"`
	Tokens           int       `json:"tokens,omitempty"`
	Response         string    `json:"response"`
	Error            string    `json:"error,omitempty"`
	Created          time.Time `json:"created_at"`
}

type ModelTester struct {
	provider LLMProvider
	config   Config
}

func NewModelTester(cfg Config) (*ModelTester, error) {
	provider, err := BuildProvider(cfg)
	if err != nil {
		return nil, err
	}
	return &ModelTester{provider: provider, config: cfg}, nil
}

func (m *ModelTester) Test(language string) (*TestResult, error) {
	return m.TestContext(context.Background(), language)
}

func (m *ModelTester) TestContext(ctx context.Context, language string) (*TestResult, error) {
	start := time.Now()
	result := &TestResult{
		Provider: m.config.Provider,
		Model:    m.config.Model,
		Status:   StatusTesting,
		Created:  time.Now().UTC(),
	}
	var resp *Response
	var err error
	if provider, ok := m.provider.(*HuggingFaceProvider); ok {
		resp, err = provider.GenerateContext(ctx, TestPrompt(language))
	} else {
		resp, err = m.provider.Generate(TestPrompt(language))
	}
	if err != nil {
		result.Status = StatusError
		if IsRateLimited(err) {
			result.Status = StatusRateLimited
		} else if IsPaymentRequired(err) {
			result.Status = StatusPaymentRequired
		} else if IsInvalidCredential(err) {
			result.Status = StatusInvalidCred
		}
		result.Error = err.Error()
		result.Latency = time.Since(start).Milliseconds()
		log.Printf("LLM request provider=%s model=%s status=%s latency_ms=%d error_category=%v", m.config.Provider, m.config.Model, result.Status, result.Latency, err)
		return result, err
	}
	result.Status = StatusSuccess
	result.Latency = resp.LatencyMS
	result.Tokens = resp.TokensUsed
	result.Response = resp.Text
	log.Printf("LLM request provider=%s model=%s status=success latency_ms=%d tokens=%d", m.config.Provider, m.config.Model, resp.LatencyMS, resp.TokensUsed)
	return result, nil
}

func (m *ModelTester) StatusString() string {
	return fmt.Sprintf("provider=%s model=%s", m.config.Provider, m.config.Model)
}
