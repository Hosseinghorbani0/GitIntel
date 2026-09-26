package llm

import (
	"os"
	"strings"
)

// LLMCredentialManager owns the configured API keys and tracks whether they are usable.
type LLMCredentialManager struct {
	keyName string
	items   []Credential
	loaded  bool
}

func NewCredentialManager(keyName string) *LLMCredentialManager {
	return &LLMCredentialManager{keyName: keyName}
}

func (m *LLMCredentialManager) Load() []Credential {
	if m.loaded {
		return m.items
	}
	m.loaded = true
	raw := strings.TrimSpace(os.Getenv(m.keyName))
	if raw == "" {
		switch m.keyName {
		case "HF_API_KEYS":
			raw = strings.TrimSpace(os.Getenv("HF_API_KEY"))
		case "LLM_API_KEYS":
			raw = strings.TrimSpace(os.Getenv("LLM_API_KEY"))
		}
	}
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	creds := make([]Credential, 0, len(parts))
	for _, part := range parts {
		key := strings.TrimSpace(part)
		if key == "" {
			continue
		}
		creds = append(creds, Credential{Value: key, Valid: true})
	}
	m.items = creds
	return creds
}

func (m *LLMCredentialManager) HasCredentials() bool {
	return len(m.Load()) > 0
}

func (m *LLMCredentialManager) SelectAvailable() (Credential, bool) {
	for _, cred := range m.items {
		if cred.Disabled || !cred.Valid || cred.Value == "" {
			continue
		}
		return cred, true
	}
	return Credential{}, false
}

func (m *LLMCredentialManager) MarkInvalid(value string, reason string) {
	for i := range m.items {
		if m.items[i].Value == value {
			m.items[i].Disabled = true
			m.items[i].Valid = false
			m.items[i].Reason = reason
			return
		}
	}
}

func (m *LLMCredentialManager) MaskedConfig() string {
	creds := m.Load()
	if len(creds) == 0 {
		return "not configured"
	}
	masked := make([]string, 0, len(creds))
	for _, c := range creds {
		masked = append(masked, MaskSecret(c.Value))
	}
	return strings.Join(masked, ", ")
}
