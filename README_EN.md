# cursor2api-go

English | [简体中文](README.md)

A Go service that turns Cursor Web into an OpenAI-compatible API for local deployment and integration.

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org)
[![License: PolyForm Noncommercial](https://img.shields.io/badge/License-PolyForm%20Noncommercial-orange.svg)](https://polyformproject.org/licenses/noncommercial/1.0.0/)

## Overview

`cursor2api-go` provides:

- OpenAI-compatible `POST /v1/chat/completions`
- OpenAI-compatible `GET /v1/models`
- Startup-time discovery of the currently allowed Cursor Web base models
- Automatic public `-thinking` variants for each base model
- Support for `tools`, `tool_choice`, and `tool_calls`
- A built-in web console at `/`
- Runtime API key rotation with `.env` persistence

## What's New

- Model discovery now happens at startup, so `MODELS` is no longer manually configured
- `/` is now an interactive console instead of a placeholder page
- Added `POST /v1/admin/api-key` for authenticated API key rotation
- Auth now reads the runtime config, so a rotated key takes effect immediately
- Added better compatibility for tool-enforcing orchestrators such as Kilo Code
- Non-stream requests auto-retry once when a tool call is required but the first pass returns none

## Screenshots

![Home preview](docs/images/home.png)
![Tool calls preview 1](docs/images/play1.png)
![Tool calls preview 2](docs/images/play2.png)

## Endpoints

| Path | Method | Auth | Description |
| --- | --- | --- | --- |
| `/` | `GET` | No | Web console with health, models, and sample commands |
| `/health` | `GET` | No | Health check |
| `/v1/models` | `GET` | No | Returns the models discovered during startup |
| `/v1/chat/completions` | `POST` | Yes | OpenAI-compatible chat endpoint with stream/non-stream/tool support |
| `/v1/admin/api-key` | `POST` | Yes | Updates and persists a new API key after authenticating with the current one |

## Model Strategy

- The service probes upstream during startup; treat `GET /v1/models` as the source of truth
- Every base model automatically exposes a matching public `-thinking` model
- A `-thinking` model maps back to the same upstream base model
- Thinking is an internal bridge capability; it is not exposed as a separate reasoning field

Example model names:

- `google/gemini-3-flash`
- `google/gemini-3-flash-thinking`

Upstream availability can change at any time, so model names in the README are examples, not a fixed contract.

## Quick Start

### Requirements

- Go 1.24+
- Node.js 18+

### Local Run

```bash
git clone https://github.com/<your-username>/cursor2api-go.git
cd cursor2api-go
cp .env.example .env
```

At minimum, update:

```dotenv
API_KEY=replace-with-your-secret
DEBUG=false
```

Start the service with any of the following:

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

On Windows, you can use:

- `start-go.bat`
- `start-go-utf8.bat`

The service listens on `http://localhost:8002` by default.

### Docker Compose

```bash
docker compose up -d --build
```

Logs:

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

## Configuration

### Key Environment Variables

| Variable | Default | Description |
| --- | --- | --- |
| `PORT` | `8002` | Service port |
| `DEBUG` | `false` | Enables detailed logging |
| `API_KEY` | `0000` | API auth key; change this in production |
| `SYSTEM_PROMPT_INJECT` | empty | Extra system prompt text to inject |
| `TIMEOUT` | `60` | Upstream timeout in seconds |
| `MAX_INPUT_LENGTH` | `200000` | Max combined input length before old messages are trimmed |
| `KILO_TOOL_STRICT` | `false` | Treat `tools + tool_choice=auto` as "tool use required" |
| `USER_AGENT` | built-in default | Browser fingerprint override |
| `UNMASKED_VENDOR_WEBGL` | built-in default | Browser fingerprint override |
| `UNMASKED_RENDERER_WEBGL` | built-in default | Browser fingerprint override |
| `SCRIPT_URL` | built-in default | Browser environment script placeholder; usually keep default |

Notes:

- `MODELS` has been removed from configuration
- The runtime API key update endpoint also writes the new value back to `.env`

## Usage Examples

### 1. List models

```bash
curl http://localhost:8002/v1/models
```

### 2. Non-stream chat

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

### 3. Stream chat

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

### 4. Tool call request

```bash
curl -X POST http://localhost:8002/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 0000" \
  -d '{
    "model": "google/gemini-3-flash",
    "stream": false,
    "messages": [
      {"role": "user", "content": "Check the weather in Beijing"}
    ],
    "tools": [
      {
        "type": "function",
        "function": {
          "name": "get_weather",
          "description": "Get weather",
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

Non-stream responses expose:

- `message.tool_calls`
- `finish_reason = "tool_calls"`

Stream responses expose:

- `delta.tool_calls`
- a final `finish_reason = "tool_calls"`

### 5. `-thinking` model

```bash
curl -X POST http://localhost:8002/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 0000" \
  -d '{
    "model": "google/gemini-3-flash-thinking",
    "stream": true,
    "messages": [
      {"role": "user", "content": "Think first, then decide whether a tool is needed"}
    ]
  }'
```

### 6. Rotate the API key

```bash
curl -X POST http://localhost:8002/v1/admin/api-key \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer 0000" \
  -d '{
    "api_key": "new-secret-key"
  }'
```

After success:

- The running process immediately switches to the new key
- `.env` is updated
- All subsequent authenticated requests must use the new key

## Web Console

Visit `/` to use the built-in console:

- check service health
- inspect discovered models
- copy ready-to-run `curl` commands
- rotate the API key in the browser

The example commands automatically use the current origin and the API key entered on the page.

## Third-Party Client Setup

For any client that supports a custom OpenAI API:

1. Base URL: `http://localhost:8002`
2. API Key: your current `API_KEY`
3. Model: fetch `GET /v1/models` first, then choose from the returned list

If your orchestrator requires an actual tool call whenever tools are present, enable:

```dotenv
KILO_TOOL_STRICT=true
```

## Behavior Notes

- Tool support is implemented through an internal prompt/parser bridge, not native Cursor tool execution
- `tool_choice` supports `auto`, `none`, `required`, and function-object targeting
- In non-stream mode, requests that require a tool call get one retry if the first attempt returns no tool calls
- The service uses dynamic browser fingerprints and refreshes them on 403 retries
- Startup output prints the active port, docs URL, health URL, and detected models

## Not Supported

- Anthropic `/v1/messages`
- MCP orchestration
- Native upstream tool execution
- Exposed reasoning/thinking response fields
- Direct local filesystem or OS command execution through the API

## Troubleshooting

See:

- [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
- [docs/API_CAPABILITIES.md](docs/API_CAPABILITIES.md)
- [docs/DYNAMIC_HEADERS.md](docs/DYNAMIC_HEADERS.md)

## Development

```bash
go test ./...
```

```bash
go build ./...
```

## License

This project uses [PolyForm Noncommercial 1.0.0](https://polyformproject.org/licenses/noncommercial/1.0.0/).

- Noncommercial use is allowed
- Commercial use is not allowed

See [LICENSE](LICENSE) for details.

## Disclaimer

You are responsible for evaluating and complying with the terms and risks of Cursor and any upstream services you rely on.
