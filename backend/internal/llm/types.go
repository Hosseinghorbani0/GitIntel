package llm

import "time"

type ErrorCategory string

const (
	ErrorInvalidCredential   ErrorCategory = "InvalidCredential"
	ErrorUnauthorized        ErrorCategory = "Unauthorized"
	ErrorRateLimited         ErrorCategory = "RateLimited"
	ErrorQuotaExceeded       ErrorCategory = "QuotaExceeded"
	ErrorPaymentRequired     ErrorCategory = "PaymentRequired"
	ErrorProviderUnavailable ErrorCategory = "ProviderUnavailable"
	ErrorModelUnavailable    ErrorCategory = "ModelUnavailable"
	ErrorTimeout             ErrorCategory = "Timeout"
	ErrorNetwork             ErrorCategory = "NetworkError"
	ErrorInvalidResponse     ErrorCategory = "InvalidResponse"
	ErrorUnknownProvider     ErrorCategory = "UnknownProviderError"
)

type Config struct {
	Provider    string  `json:"provider"`
	Model       string  `json:"model"`
	BaseURL     string  `json:"base_url"`
	Credential  string  `json:"credential,omitempty"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
	Timeout     int     `json:"timeout"`
}

type Request struct {
	System   string      `json:"system"`
	User     string      `json:"user"`
	Input    interface{} `json:"input"`
	Language string      `json:"language"`
}

type Response struct {
	Text          string        `json:"text"`
	Provider      string        `json:"provider"`
	Model         string        `json:"model"`
	LatencyMS     int64         `json:"latency_ms"`
	TokensUsed    int           `json:"tokens_used,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	Status        string        `json:"status"`
	ErrorCategory ErrorCategory `json:"error_category,omitempty"`
}

type LLMProvider interface {
	Name() string
	Generate(req Request) (*Response, error)
	TestConnection() error
	ListModels() ([]string, error)
	Health() error
}

// Credential stores a secret without exposing the full value to logs or frontend responses.
type Credential struct {
	Value    string `json:"-"`
	Disabled bool   `json:"disabled"`
	Valid    bool   `json:"valid"`
	Reason   string `json:"reason,omitempty"`
}

func (c Credential) Masked() string {
	if c.Value == "" {
		return "not configured"
	}
	if len(c.Value) <= 8 {
		return "********"
	}
	if len(c.Value) <= 12 {
		return c.Value[:3] + "********"
	}
	return c.Value[:3] + "********" + c.Value[len(c.Value)-3:]
}
