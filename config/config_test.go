// Copyright (c) 2025-2026 libaxuan
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary .env file for testing
	envContent := `PORT=9000
DEBUG=true
API_KEY=test-key
MODELS=claude-sonnet-4.6
SYSTEM_PROMPT_INJECT=Test prompt
TIMEOUT=60
MAX_INPUT_LENGTH=10000
USER_AGENT=Test Agent
SCRIPT_URL=https://test.com/script.js`

	// Write to temporary .env file
	err := os.WriteFile(".env", []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}
	defer os.Remove(".env") // Clean up

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Test loaded values
	if config.Port != 9000 {
		t.Errorf("Port = %v, want 9000", config.Port)
	}
	if !config.Debug {
		t.Errorf("Debug = %v, want true", config.Debug)
	}
	if config.APIKey != "test-key" {
		t.Errorf("APIKey = %v, want test-key", config.APIKey)
	}
	if config.SystemPromptInject != "Test prompt" {
		t.Errorf("SystemPromptInject = %v, want Test prompt", config.SystemPromptInject)
	}
	if config.Timeout != 60 {
		t.Errorf("Timeout = %v, want 60", config.Timeout)
	}
	if config.MaxInputLength != 10000 {
		t.Errorf("MaxInputLength = %v, want 10000", config.MaxInputLength)
	}
	if config.FP.UserAgent != "Test Agent" {
		t.Errorf("UserAgent = %v, want Test Agent", config.FP.UserAgent)
	}
	if config.ScriptURL != "https://test.com/script.js" {
		t.Errorf("ScriptURL = %v, want https://test.com/script.js", config.ScriptURL)
	}
}

func TestGetModels(t *testing.T) {
	config := &Config{
		Models: "claude-sonnet-4.6",
	}

	models := config.GetModels()
	expected := []string{
		"claude-sonnet-4.6",
		"claude-sonnet-4.6-thinking",
	}

	if len(models) != len(expected) {
		t.Errorf("GetModels() length = %v, want %v", len(models), len(expected))
	}

	for i, model := range models {
		if model != expected[i] {
			t.Errorf("GetModels()[%d] = %v, want %v", i, model, expected[i])
		}
	}
}

func TestSetBaseModels(t *testing.T) {
	config := &Config{}

	config.SetBaseModels([]string{
		" google/gemini-3-flash ",
		"",
		"google/gemini-3-flash",
		"openai/gpt-5",
	})

	expectedBaseModels := []string{
		"google/gemini-3-flash",
		"openai/gpt-5",
	}
	baseModels := config.GetBaseModels()
	if len(baseModels) != len(expectedBaseModels) {
		t.Fatalf("GetBaseModels() length = %v, want %v", len(baseModels), len(expectedBaseModels))
	}
	for i, model := range baseModels {
		if model != expectedBaseModels[i] {
			t.Fatalf("GetBaseModels()[%d] = %v, want %v", i, model, expectedBaseModels[i])
		}
	}

	expectedModels := []string{
		"google/gemini-3-flash",
		"google/gemini-3-flash-thinking",
		"openai/gpt-5",
		"openai/gpt-5-thinking",
	}
	models := config.GetModels()
	if len(models) != len(expectedModels) {
		t.Fatalf("GetModels() length = %v, want %v", len(models), len(expectedModels))
	}
	for i, model := range models {
		if model != expectedModels[i] {
			t.Fatalf("GetModels()[%d] = %v, want %v", i, model, expectedModels[i])
		}
	}
}

func TestIsValidModel(t *testing.T) {
	config := &Config{
		Models: "claude-sonnet-4.6",
	}

	tests := []struct {
		name     string
		model    string
		expected bool
	}{
		{"valid base model", "claude-sonnet-4.6", true},
		{"valid thinking model", "claude-sonnet-4.6-thinking", true},
		{"invalid model", "unknown-model", false},
		{"empty model", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.IsValidModel(tt.model)
			if result != tt.expected {
				t.Errorf("IsValidModel(%q) = %v, want %v", tt.model, result, tt.expected)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Port:           8000,
				APIKey:         "test-key",
				Timeout:        30,
				MaxInputLength: 1000,
			},
			wantErr: false,
		},
		{
			name: "invalid port - too low",
			config: &Config{
				Port:           0,
				APIKey:         "test-key",
				Timeout:        30,
				MaxInputLength: 1000,
			},
			wantErr: true,
		},
		{
			name: "invalid port - too high",
			config: &Config{
				Port:           70000,
				APIKey:         "test-key",
				Timeout:        30,
				MaxInputLength: 1000,
			},
			wantErr: true,
		},
		{
			name: "missing API key",
			config: &Config{
				Port:           8000,
				APIKey:         "",
				Timeout:        30,
				MaxInputLength: 1000,
			},
			wantErr: true,
		},
		{
			name: "invalid timeout",
			config: &Config{
				Port:           8000,
				APIKey:         "test-key",
				Timeout:        0,
				MaxInputLength: 1000,
			},
			wantErr: true,
		},
		{
			name: "invalid max input length",
			config: &Config{
				Port:           8000,
				APIKey:         "test-key",
				Timeout:        30,
				MaxInputLength: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateAPIKeyCreatesEnvFileAndUpdatesRuntimeValue(t *testing.T) {
	tempDir := t.TempDir()
	config := &Config{
		APIKey:      "old-key",
		EnvFilePath: tempDir + "/.env",
	}

	if err := config.UpdateAPIKey("new-key"); err != nil {
		t.Fatalf("UpdateAPIKey() error = %v", err)
	}

	if got := config.GetAPIKey(); got != "new-key" {
		t.Fatalf("GetAPIKey() = %v, want new-key", got)
	}

	data, err := os.ReadFile(config.EnvFilePath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", config.EnvFilePath, err)
	}
	if string(data) != "API_KEY=new-key\n" {
		t.Fatalf("env file = %q, want %q", string(data), "API_KEY=new-key\n")
	}
}

func TestUpdateAPIKeyReplacesExistingEnvValue(t *testing.T) {
	tempDir := t.TempDir()
	envPath := tempDir + "/.env"
	if err := os.WriteFile(envPath, []byte("PORT=8002\nAPI_KEY=old-key\nDEBUG=false\n"), 0644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", envPath, err)
	}

	config := &Config{
		APIKey:      "old-key",
		EnvFilePath: envPath,
	}

	if err := config.UpdateAPIKey("rotated-key"); err != nil {
		t.Fatalf("UpdateAPIKey() error = %v", err)
	}

	data, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", envPath, err)
	}
	expected := "PORT=8002\nAPI_KEY=rotated-key\nDEBUG=false\n"
	if string(data) != expected {
		t.Fatalf("env file = %q, want %q", string(data), expected)
	}
}
