package main

import (
	"strings"
	"testing"
)

func TestMCPInitialize(t *testing.T) {
	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	}

	// Capture output via response handler test
	var capturedResp JSONRPCResponse
	sendResponse = func(id interface{}, result interface{}) {
		capturedResp = JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result:  result,
		}
	}

	handleRequest(req)

	if capturedResp.ID != 1 {
		t.Fatalf("expected ID 1, got %v", capturedResp.ID)
	}

	resMap, ok := capturedResp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected result to be map, got %T", capturedResp.Result)
	}

	if resMap["protocolVersion"] != "2024-11-05" {
		t.Errorf("expected protocolVersion 2024-11-05, got %v", resMap["protocolVersion"])
	}
}

func TestMCPToolsList(t *testing.T) {
	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      "req-2",
		Method:  "tools/list",
	}

	var capturedResp JSONRPCResponse
	sendResponse = func(id interface{}, result interface{}) {
		capturedResp = JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result:  result,
		}
	}

	handleRequest(req)

	resMap, ok := capturedResp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected result to be map, got %T", capturedResp.Result)
	}

	tools, ok := resMap["tools"].([]map[string]interface{})
	if !ok || len(tools) != 4 {
		t.Fatalf("expected 4 registered tools, got %d", len(tools))
	}

	toolNames := make([]string, 0, len(tools))
	for _, tool := range tools {
		toolNames = append(toolNames, tool["name"].(string))
	}

	expected := []string{
		"gitintel_analyze_profile",
		"gitintel_get_signals",
		"gitintel_generate_profile_readme",
		"gitintel_get_ai_prompt",
	}

	for _, name := range expected {
		found := false
		for _, tn := range toolNames {
			if tn == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected tool %s to be listed", name)
		}
	}
}

func TestMCPUnknownMethodReturnsError(t *testing.T) {
	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      99,
		Method:  "non_existent_method",
	}

	var capturedErr *RPCError
	sendError = func(id interface{}, code int, message string) {
		capturedErr = &RPCError{
			Code:    code,
			Message: message,
		}
	}

	handleRequest(req)

	if capturedErr == nil {
		t.Fatal("expected RPC error for unknown method")
	}
	if capturedErr.Code != -32601 {
		t.Errorf("expected code -32601, got %d", capturedErr.Code)
	}
	if !strings.Contains(capturedErr.Message, "Method not found") {
		t.Errorf("unexpected error message: %s", capturedErr.Message)
	}
}
