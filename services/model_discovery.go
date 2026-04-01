package services

import (
	"context"
	"cursor2api-go/middleware"
	"cursor2api-go/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const modelDiscoveryProbeName = "cursor2api-auto-discovery-probe"

type upstreamErrorPayload struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// DiscoverAvailableModels probes Cursor Web for the currently allowed upstream models.
// The upstream rejects the synthetic model name and returns the real allow-list in its error message.
func (s *CursorService) DiscoverAvailableModels(ctx context.Context) ([]string, error) {
	buildResult, err := s.buildCursorRequest(&models.ChatCompletionRequest{
		Model: modelDiscoveryProbeName,
		Messages: []models.Message{
			{Role: "user", Content: "ping"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("build discovery probe: %w", err)
	}

	jsonPayload, err := json.Marshal(buildResult.Payload)
	if err != nil {
		return nil, fmt.Errorf("marshal discovery probe payload: %w", err)
	}

	maxRetries := 2
	for attempt := 1; attempt <= maxRetries; attempt++ {
		xIsHuman, err := s.fetchXIsHuman(ctx)
		if err != nil {
			if attempt < maxRetries {
				time.Sleep(time.Second * time.Duration(attempt))
				continue
			}
			return nil, fmt.Errorf("fetch discovery x-is-human: %w", err)
		}

		resp, err := s.client.R().
			SetContext(ctx).
			SetHeaders(s.chatHeaders(xIsHuman)).
			SetBody(jsonPayload).
			DisableAutoReadResponse().
			Post(cursorAPIURL)
		if err != nil {
			if attempt < maxRetries {
				time.Sleep(time.Second * time.Duration(attempt))
				continue
			}
			return nil, fmt.Errorf("send discovery probe: %w", err)
		}

		body, readErr := io.ReadAll(resp.Response.Body)
		resp.Response.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("read discovery probe response: %w", readErr)
		}

		message := strings.TrimSpace(string(body))
		if resp.StatusCode == http.StatusOK {
			return nil, fmt.Errorf("discovery probe unexpectedly succeeded")
		}
		if resp.StatusCode == http.StatusForbidden && attempt < maxRetries {
			s.headerGenerator.Refresh()
			time.Sleep(time.Second * time.Duration(attempt))
			continue
		}

		allowedModels, parseErr := parseAllowedModels(message)
		if parseErr != nil {
			return nil, fmt.Errorf("parse allowed models from status %d: %w", resp.StatusCode, parseErr)
		}
		if len(allowedModels) == 0 {
			return nil, middleware.NewCursorWebError(resp.StatusCode, "upstream returned an empty allowed model list")
		}
		return allowedModels, nil
	}

	return nil, fmt.Errorf("discover available models failed after %d attempts", maxRetries)
}

func parseAllowedModels(message string) ([]string, error) {
	text := strings.TrimSpace(message)
	if text == "" {
		return nil, fmt.Errorf("empty upstream error message")
	}

	var payload upstreamErrorPayload
	if err := json.Unmarshal([]byte(text), &payload); err == nil {
		switch {
		case strings.TrimSpace(payload.Error) != "":
			text = strings.TrimSpace(payload.Error)
		case strings.TrimSpace(payload.Message) != "":
			text = strings.TrimSpace(payload.Message)
		}
	}

	const marker = "Allowed models:"
	idx := strings.Index(text, marker)
	if idx == -1 {
		return nil, fmt.Errorf("allowed model list not found in upstream message: %s", text)
	}

	rawModels := strings.TrimSpace(text[idx+len(marker):])
	rawModels = strings.Trim(rawModels, ".")
	if rawModels == "" {
		return nil, fmt.Errorf("allowed model list is empty")
	}

	parts := strings.Split(rawModels, ",")
	result := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		model := strings.TrimSpace(strings.Trim(part, `"'`))
		if model == "" {
			continue
		}
		if _, exists := seen[model]; exists {
			continue
		}
		seen[model] = struct{}{}
		result = append(result, model)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("allowed model list is empty")
	}
	return result, nil
}
