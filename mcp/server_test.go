package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

// jsonRPCRequest 构建 JSON-RPC 请求体
func jsonRPCRequest(id int, method string, params map[string]interface{}) string {
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	}
	b, _ := json.Marshal(req)
	return string(b)
}

// postMCP 发送 POST 请求到 MCP 端点，返回 response 和 headers
func postMCP(url, body, sessionID string) (*http.Response, string, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if sessionID != "" {
		req.Header.Set("Mcp-Session-Id", sessionID)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: nil, // 跳过代理
		},
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, "", err
	}
	return resp, string(b), nil
}

// getFreePort 获取一个可用端口
func getFreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func TestServeHTTP_Initialize(t *testing.T) {
	port, err := getFreePort()
	if err != nil {
		t.Fatalf("获取空闲端口失败: %v", err)
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	url := fmt.Sprintf("http://%s/mcp", addr)

	// 使用临时空目录作为数据目录
	dataDir := t.TempDir()
	s := NewServer(dataDir, "test")

	// 在 goroutine 中启动 HTTP 服务
	errCh := make(chan error, 1)
	go func() {
		errCh <- ServeHTTP(s, addr)
	}()

	// 等待服务器启动
	time.Sleep(500 * time.Millisecond)

	// 测试 initialize
	body := jsonRPCRequest(1, "initialize", map[string]interface{}{
		"protocolVersion": "2025-03-26",
		"capabilities":    map[string]interface{}{},
		"clientInfo":      map[string]interface{}{"name": "test", "version": "1.0"},
	})

	resp, respBody, err := postMCP(url, body, "")
	if err != nil {
		t.Fatalf("initialize 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d, body: %s", resp.StatusCode, respBody)
	}

	// 检查响应包含 serverInfo
	if !strings.Contains(respBody, "AfdianToMarkdown") {
		t.Errorf("响应中应包含 serverInfo.name='AfdianToMarkdown', 实际: %s", respBody)
	}

	// 检查 session ID header
	sessionID := resp.Header.Get("Mcp-Session-Id")
	if sessionID == "" {
		t.Error("响应中应包含 Mcp-Session-Id header")
	}
}

func TestServeHTTP_ToolCall(t *testing.T) {
	port, err := getFreePort()
	if err != nil {
		t.Fatalf("获取空闲端口失败: %v", err)
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	url := fmt.Sprintf("http://%s/mcp", addr)

	dataDir := t.TempDir()
	s := NewServer(dataDir, "test")

	errCh := make(chan error, 1)
	go func() {
		errCh <- ServeHTTP(s, addr)
	}()
	time.Sleep(500 * time.Millisecond)

	// Step 1: Initialize 并获取 session ID
	initBody := jsonRPCRequest(1, "initialize", map[string]interface{}{
		"protocolVersion": "2025-03-26",
		"capabilities":    map[string]interface{}{},
		"clientInfo":      map[string]interface{}{"name": "test", "version": "1.0"},
	})

	initResp, _, err := postMCP(url, initBody, "")
	if err != nil {
		t.Fatalf("initialize 失败: %v", err)
	}
	sessionID := initResp.Header.Get("Mcp-Session-Id")

	// Step 2: 调用 list_authors 工具
	callBody := jsonRPCRequest(2, "tools/call", map[string]interface{}{
		"name":      "list_authors",
		"arguments": map[string]interface{}{},
	})

	resp, respBody, err := postMCP(url, callBody, sessionID)
	if err != nil {
		t.Fatalf("tools/call 请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("期望状态码 200，实际 %d, body: %s", resp.StatusCode, respBody)
	}

	// 空目录应该返回 "暂无已下载的作者" 或类似文本
	if !strings.Contains(respBody, "result") {
		t.Errorf("响应中应包含 result 字段, 实际: %s", respBody)
	}
}
