# cursor2api-go

[English](README_EN.md) | 简体中文

将 Cursor Web 转换为 OpenAI 兼容接口的 Go 服务，适合本地自建和二次集成。

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org)
[![License: PolyForm Noncommercial](https://img.shields.io/badge/License-PolyForm%20Noncommercial-orange.svg)](https://polyformproject.org/licenses/noncommercial/1.0.0/)

## 概览

`cursor2api-go` 提供以下能力：

- OpenAI 兼容的 `POST /v1/chat/completions`
- OpenAI 兼容的 `GET /v1/models`
- 启动时自动探测 Cursor Web 当前允许的基础模型
- 自动公开每个基础模型对应的 `-thinking` 版本
- 支持 `tools`、`tool_choice`、`tool_calls`
- 内置 Web 控制台，可直接查看状态、模型和示例请求
- 支持运行时热更新 API Key，并写回 `.env`

## 最新功能

- 启动即探测上游模型，不再依赖手工维护 `MODELS`
- 根路径 `/` 现在是可交互控制台，不是静态占位页
- 新增 `POST /v1/admin/api-key`，认证通过后可立即切换 API Key
- `Authorization` 校验改为读取运行时配置，热更新后即时生效
- 针对 Kilo Code 一类“必须用工具”的编排器增加兼容模式
- 非流式场景下，若本轮必须调用工具却没有产出 `tool_calls`，会自动补救重试 1 次

## 接口一览

| 路径 | 方法 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| `/` | `GET` | 否 | Web 控制台，动态显示健康状态、模型列表和示例命令 |
| `/health` | `GET` | 否 | 健康检查 |
| `/v1/models` | `GET` | 否 | 返回当前启动时探测到的模型列表 |
| `/v1/chat/completions` | `POST` | 是 | OpenAI 兼容聊天接口，支持流式、非流式、tools |
| `/v1/admin/api-key` | `POST` | 是 | 用当前 API Key 认证后，更新并持久化新的 API Key |

## 模型策略

- 服务启动时会主动向上游发起探测，请以 `GET /v1/models` 结果为准
- 每个基础模型都会自动公开一个 `-thinking` 版本
- `-thinking` 是公开模型别名，实际仍映射回对应基础模型请求上游
- thinking 只在内部桥接协议中使用，不会作为独立 reasoning 字段暴露给客户端

示例：

- `google/gemini-3-flash`
- `google/gemini-3-flash-thinking`

上游允许模型随时可能变化，README 中的模型名只能作为示例，不能当成固定清单。

## 快速开始

### 环境要求

- Go 1.24+
- Node.js 18+

### 本地运行

```bash
git clone https://github.com/<your-username>/cursor2api-go.git
cd cursor2api-go
cp .env.example .env
```

至少修改以下配置：

```dotenv
API_KEY=replace-with-your-secret
DEBUG=false
```

启动方式任选其一：

```bash
go run .
```

```bash
go build -o cursor2api-go
./cursor2api-go
```

```bash
chmod +x start.sh
./start.sh
```

Windows 可直接使用：

- `start-go.bat`
- `start-go-utf8.bat`

服务默认监听 `http://localhost:8002`。

### Docker Compose

```bash
docker compose up -d --build
```

查看日志：

```bash
docker compose logs -f
```

### Docker

```bash
docker build -t cursor2api-go .
docker run -d \
  --name cursor2api-go \
  --restart unless-stopped \
  -p 8002:8002 \
  -e API_KEY=replace-with-your-secret \
  cursor2api-go
```

## 配置说明

### 关键环境变量

| 变量名 | 默认值 | 说明 |
| --- | --- | --- |
| `PORT` | `8002` | 服务端口 |
| `DEBUG` | `false` | 调试模式，开启后输出更详细日志 |
| `API_KEY` | `0000` | API 认证密钥，生产环境必须修改 |
| `SYSTEM_PROMPT_INJECT` | 空 | 追加注入到系统提示词 |
| `TIMEOUT` | `60` | 上游请求超时时间，单位秒 |
| `MAX_INPUT_LENGTH` | `200000` | 输入消息总长度上限，超出后会裁剪旧消息 |
| `KILO_TOOL_STRICT` | `false` | 当提供 `tools` 且 `tool_choice=auto` 时，强制按“必须用工具”处理 |
| `USER_AGENT` | 内置默认值 | 浏览器指纹字段，可按需覆盖 |
| `UNMASKED_VENDOR_WEBGL` | 内置默认值 | 浏览器指纹字段，可按需覆盖 |
| `UNMASKED_RENDERER_WEBGL` | 内置默认值 | 浏览器指纹字段，可按需覆盖 |
| `SCRIPT_URL` | 内置默认值 | 浏览器环境仿真脚本占位配置，通常保持默认即可 |

说明：

- `MODELS` 已移除，模型列表由服务启动时自动探测
- `.env.example` 是推荐起点，运行时 API Key 更新也会回写到 `.env`

## 使用示例

### 1. 获取模型列表

```bash
curl http://localhost:8002/v1/models
```

### 2. 非流式聊天

```bash
curl -X POST http://localhost:8002/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 0000" \
  -d '{
    "model": "google/gemini-3-flash",
    "messages": [
      {"role": "user", "content": "reply with exactly OK"}
    ],
    "stream": false
  }'
```

### 3. 流式聊天

```bash
curl -X POST http://localhost:8002/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 0000" \
  -d '{
    "model": "google/gemini-3-flash",
    "messages": [
      {"role": "user", "content": "write a haiku about Go"}
    ],
    "stream": true
  }'
```

### 4. tools 调用

```bash
curl -X POST http://localhost:8002/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 0000" \
  -d '{
    "model": "google/gemini-3-flash",
    "stream": false,
    "messages": [
      {"role": "user", "content": "帮我查询北京天气"}
    ],
    "tools": [
      {
        "type": "function",
        "function": {
          "name": "get_weather",
          "description": "获取天气",
          "parameters": {
            "type": "object",
            "properties": {
              "city": {"type": "string"}
            },
            "required": ["city"]
          }
        }
      }
    ]
  }'
```

非流式返回会兼容 OpenAI 的：

- `message.tool_calls`
- `finish_reason = "tool_calls"`

流式返回会兼容：

- `delta.tool_calls`
- 末尾 `finish_reason = "tool_calls"`

### 5. `-thinking` 模型

```bash
curl -X POST http://localhost:8002/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 0000" \
  -d '{
    "model": "google/gemini-3-flash-thinking",
    "stream": true,
    "messages": [
      {"role": "user", "content": "先思考，再决定是否需要工具"}
    ]
  }'
```

### 6. 热更新 API Key

```bash
curl -X POST http://localhost:8002/v1/admin/api-key \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 0000" \
  -d '{
    "api_key": "new-secret-key"
  }'
```

更新成功后：

- 当前进程立即切换到新的 API Key
- `.env` 中的 `API_KEY` 会被同步更新
- 后续请求必须使用新 key 认证

## Web 控制台

访问根路径 `/` 可直接使用内置控制台：

- 查看服务健康状态
- 查看自动探测到的模型
- 复制示例 `curl`
- 在页面里直接更新 API Key

控制台中的示例代码会自动使用当前页面访问地址和已输入的 API Key。

## 第三方应用接入

在支持自定义 OpenAI API 的客户端中，按以下方式配置：

1. Base URL: `http://localhost:8002`
2. API Key: 你当前配置的 `API_KEY`
3. Model: 先调用 `GET /v1/models`，再从返回列表中选择

如果你的上层编排器要求“提供了 tools 就必须真正发起工具调用”，建议启用：

```dotenv
KILO_TOOL_STRICT=true
```

## 行为说明

- tools 支持通过内部 prompt/parse bridge 实现，不是 Cursor 原生工具调用
- `tool_choice` 支持 `auto`、`none`、`required` 和指定函数对象
- 非流式模式下，如果请求被判定为“必须用工具”但首次没有工具调用，会自动重试 1 次
- 为提高稳定性，服务会动态生成浏览器指纹；遇到 403 时会刷新指纹并重试
- 启动横幅会打印当前端口、文档地址、健康检查和模型列表

## 当前不支持

- Anthropic `/v1/messages`
- MCP 编排
- 原生上游 tool execution
- 对外暴露独立 reasoning/thinking 字段
- 通过 API 直接执行本地文件系统或系统命令

## 故障排除

常见问题见：

- [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
- [docs/API_CAPABILITIES.md](docs/API_CAPABILITIES.md)
- [docs/DYNAMIC_HEADERS.md](docs/DYNAMIC_HEADERS.md)

## 开发

```bash
go test ./...
```

```bash
go build ./...
```

## 许可证

本项目采用 [PolyForm Noncommercial 1.0.0](https://polyformproject.org/licenses/noncommercial/1.0.0/)。

- 允许非商业使用
- 不允许商业使用

详情见 [LICENSE](LICENSE)。

## 免责声明

请自行评估并遵守 Cursor 及相关上游服务的使用条款与风险。
