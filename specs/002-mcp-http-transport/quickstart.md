# Quickstart: MCP HTTP Transport

## 开发环境

- Go 1.24+
- 已有的 AfdianToMarkdown 项目代码
- mcp-go v0.43.2（已在 go.mod 中）

## 构建与运行

```bash
# 构建
go build -o AfdianToMarkdown .

# 以 HTTP 模式启动 MCP 服务
./AfdianToMarkdown --dir ./data mcp --http

# 指定监听地址和端口
./AfdianToMarkdown --dir ./data mcp --http --addr 0.0.0.0:9090

# 以 stdio 模式启动（与之前一致）
./AfdianToMarkdown --dir ./data mcp
```

## 验证

使用 curl 测试 MCP 端点：

```bash
# 发送 initialize 请求
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}'

# 调用 list_authors 工具
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_authors","arguments":{}}}'
```

## NAS 部署示例

### Cron 更新 + 服务重启脚本

```bash
#!/bin/bash
# update-and-serve.sh

APP=/path/to/AfdianToMarkdown
DATA_DIR=/path/to/data
COOKIE=/path/to/cookies.json
PID_FILE=/tmp/afdian-mcp.pid

# 停止现有服务
if [ -f "$PID_FILE" ]; then
    kill "$(cat $PID_FILE)" 2>/dev/null
    sleep 2
fi

# 更新内容
$APP --dir "$DATA_DIR" --cookie "$COOKIE" update

# 重新启动 HTTP MCP 服务
nohup $APP --dir "$DATA_DIR" mcp --http > /tmp/afdian-mcp.log 2>&1 &
echo $! > "$PID_FILE"
```

### Systemd 服务文件示例

```ini
[Unit]
Description=AfdianToMarkdown MCP Server
After=network.target

[Service]
ExecStart=/path/to/AfdianToMarkdown --dir /path/to/data mcp --http
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

## 修改范围

| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `main.go` | 修改 | mcp 子命令添加 `--http`、`--addr` 标志 |
| `mcp/server.go` | 修改 | 添加 `ServeHTTP()` 函数 |

工具处理器（tools.go）、搜索逻辑（search.go）、存储层（storage/）无需修改。
