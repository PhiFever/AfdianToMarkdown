# MCP HTTP Endpoint Contract

## Transport Protocol

**协议**: MCP Streamable HTTP (MCP Specification 2025-03-26)
**端点**: `POST /mcp`（JSON-RPC over HTTP）
**内容类型**: `application/json`

## 端点行为

### POST /mcp
- **请求**: JSON-RPC 2.0 消息（单条或批量）
- **响应**:
  - `Content-Type: application/json` — 直接 JSON-RPC 响应
  - `Content-Type: text/event-stream` — SSE 流式响应（当需要流式传输时）
- **无认证**: 无 Authorization header，无 token

### GET /mcp
- **用途**: 打开 SSE 流接收服务器推送通知
- **响应**: `Content-Type: text/event-stream`

### DELETE /mcp
- **用途**: 关闭会话
- **响应**: 204 No Content

## CLI 接口

```bash
# 启动 HTTP 模式（默认地址 0.0.0.0:8080）
./AfdianToMarkdown mcp --http

# 指定监听地址
./AfdianToMarkdown mcp --http --addr 127.0.0.1:9090

# stdio 模式（默认，向后兼容）
./AfdianToMarkdown mcp
```

## MCP 客户端配置示例

```json
{
  "mcpServers": {
    "afdian": {
      "url": "http://<tailscale-ip>:8080/mcp"
    }
  }
}
```
