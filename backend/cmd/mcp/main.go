package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"gitintel/backend/internal/analytics"
	gh "gitintel/backend/internal/github"
)

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ToolCallResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

func main() {
	// Set log to stderr so stdout is strictly preserved for JSON-RPC messages
	log.SetOutput(os.Stderr)
	log.SetPrefix("[GitIntel MCP] ")

	dec := json.NewDecoder(os.Stdin)
	for {
		var req JSONRPCRequest
		if err := dec.Decode(&req); err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Decode error: %v", err)
			sendError(nil, -32700, "Parse error")
			// Discard invalid token and continue
			continue
		}

		handleRequest(&req)
	}
}

func handleRequest(req *JSONRPCRequest) {
	switch req.Method {
	case "initialize":
		sendResponse(req.ID, map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "gitintel",
				"version": "1.0.0",
			},
		})

	case "notifications/initialized":
		// Client ACK - no response for notifications

	case "ping":
		sendResponse(req.ID, map[string]interface{}{})

	case "tools/list":
		tools := []map[string]interface{}{
			{
				"name":        "gitintel_analyze_profile",
				"description": "Performs deterministic, evidence-grounded GitHub engineering analysis for a developer. Returns structured metrics, verified signals, tech stack byte volume, and candidate projects without hallucination.",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"username": map[string]interface{}{
							"type":        "string",
							"description": "GitHub username to analyze",
						},
						"token": map[string]interface{}{
							"type":        "string",
							"description": "Optional GitHub Personal Access Token for higher rate-limit budget",
						},
					},
					"required": []string{"username"},
				},
			},
			{
				"name":        "gitintel_get_signals",
				"description": "Retrieves the 6 verified deterministic engineering signals (Code Evolution, Maintenance, Testing Evidence, Collaboration, Documentation, Release Delivery) with explicit limitation disclosures.",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"username": map[string]interface{}{
							"type":        "string",
							"description": "GitHub username",
						},
						"token": map[string]interface{}{
							"type":        "string",
							"description": "Optional GitHub token",
						},
					},
					"required": []string{"username"},
				},
			},
			{
				"name":        "gitintel_generate_profile_readme",
				"description": "Generates a complete, beautiful, ready-to-commit GitHub Profile README (username/username) in Markdown, showcasing verified code volume, testing, CI/CD, and highlighted repositories.",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"username": map[string]interface{}{
							"type":        "string",
							"description": "GitHub username",
						},
						"token": map[string]interface{}{
							"type":        "string",
							"description": "Optional GitHub token",
						},
					},
					"required": []string{"username"},
				},
			},
			{
				"name":        "gitintel_get_ai_prompt",
				"description": "Returns pre-engineered prompt context containing verified GitHub engineering telemetry, ready to feed into LLMs to write tailored resumes or cover letters without hallucination.",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"username": map[string]interface{}{
							"type":        "string",
							"description": "GitHub username",
						},
						"token": map[string]interface{}{
							"type":        "string",
							"description": "Optional GitHub token",
						},
					},
					"required": []string{"username"},
				},
			},
		}
		sendResponse(req.ID, map[string]interface{}{"tools": tools})

	case "tools/call":
		var params ToolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			sendError(req.ID, -32602, "Invalid params")
			return
		}
		handleToolCall(req.ID, params)

	default:
		if req.ID != nil {
			sendError(req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method))
		}
	}
}

func handleToolCall(id interface{}, params ToolCallParams) {
	username, _ := params.Arguments["username"].(string)
	username = strings.TrimSpace(username)
	if username == "" {
		sendToolResult(id, "Error: username is required", true)
		return
	}

	token, _ := params.Arguments["token"].(string)
	token = strings.TrimSpace(token)
	if token == "" {
		token = strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	}

	client := gh.NewClientFromToken(token)
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	profile, err := client.GetProfile(ctx, username)
	if err != nil {
		sendToolResult(id, fmt.Sprintf("Error fetching profile: %v", err), true)
		return
	}

	repos, err := client.GetRepositories(ctx, username)
	if err != nil {
		sendToolResult(id, fmt.Sprintf("Error fetching repositories: %v", err), true)
		return
	}

	// Bounded checks for high-signal candidates (up to 5 each)
	rateInfo, _ := client.RateLimit(ctx)
	if rateInfo.Remaining >= 10 {
		repos = client.VerifyCandidateReadmes(ctx, repos, 5)
		repos = client.VerifyCandidateReleases(ctx, repos, 5)
		if verified, err := client.VerifyTestingArtifacts(ctx, repos, 5); err == nil {
			repos = verified
		}
	}

	engine := analytics.NewEngine()
	analysis := engine.Analyze(*profile, repos)

	switch params.Name {
	case "gitintel_analyze_profile":
		payload, err := json.MarshalIndent(map[string]interface{}{
			"profile":               profile,
			"metrics":               analysis.Metrics,
			"summary":               analysis.Summary,
			"language_distribution": analysis.LanguageDistribution,
			"signals":               analysis.Signals,
			"featured_projects":     analysis.FeaturedProjects,
		}, "", "  ")
		if err != nil {
			sendToolResult(id, fmt.Sprintf("Error serializing analysis: %v", err), true)
			return
		}
		sendToolResult(id, string(payload), false)

	case "gitintel_get_signals":
		payload, err := json.MarshalIndent(map[string]interface{}{
			"username": profile.Username,
			"signals":  analysis.Signals,
		}, "", "  ")
		if err != nil {
			sendToolResult(id, fmt.Sprintf("Error serializing signals: %v", err), true)
			return
		}
		sendToolResult(id, string(payload), false)

	case "gitintel_generate_profile_readme":
		sendToolResult(id, analysis.ProfileReadmeMarkdown, false)

	case "gitintel_get_ai_prompt":
		sendToolResult(id, analysis.AIPromptMarkdown, false)

	default:
		sendError(id, -32601, fmt.Sprintf("Unknown tool: %s", params.Name))
	}
}

func sendToolResult(id interface{}, text string, isError bool) {
	sendResponse(id, ToolCallResult{
		Content: []ToolContent{
			{
				Type: "text",
				Text: text,
			},
		},
		IsError: isError,
	})
}

var sendResponse = func(id interface{}, result interface{}) {
	if id == nil {
		return
	}
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	bytes, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Marshal error: %v", err)
		return
	}
	fmt.Printf("%s\n", string(bytes))
}

var sendError = func(id interface{}, code int, message string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
	}
	bytes, _ := json.Marshal(resp)
	fmt.Printf("%s\n", string(bytes))
}
