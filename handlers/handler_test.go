package handlers

import (
	"bytes"
	"cursor2api-go/config"
	"cursor2api-go/middleware"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUpdateAPIKeyHotReloadsAuthenticationAndPersistsEnv(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()
	envPath := tempDir + "/.env"
	if err := os.WriteFile(envPath, []byte("API_KEY=old-key\n"), 0644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", envPath, err)
	}

	cfg := &config.Config{
		APIKey:             "old-key",
		EnvFilePath:        envPath,
		MaxInputLength:     1000,
		SystemPromptInject: "",
	}
	handler := &Handler{config: cfg}

	router := gin.New()
	v1 := router.Group("/v1")
	v1.POST("/admin/api-key", middleware.AuthRequired(cfg), handler.UpdateAPIKey)
	v1.POST("/protected", middleware.AuthRequired(cfg), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	updateBody, _ := json.Marshal(map[string]string{"api_key": "new-key"})
	updateReq := httptest.NewRequest(http.MethodPost, "/v1/admin/api-key", bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer old-key")
	updateResp := httptest.NewRecorder()
	router.ServeHTTP(updateResp, updateReq)

	if updateResp.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d, body=%s", updateResp.Code, http.StatusOK, updateResp.Body.String())
	}

	oldReq := httptest.NewRequest(http.MethodPost, "/v1/protected", nil)
	oldReq.Header.Set("Authorization", "Bearer old-key")
	oldResp := httptest.NewRecorder()
	router.ServeHTTP(oldResp, oldReq)
	if oldResp.Code != http.StatusUnauthorized {
		t.Fatalf("old key status = %d, want %d", oldResp.Code, http.StatusUnauthorized)
	}

	newReq := httptest.NewRequest(http.MethodPost, "/v1/protected", nil)
	newReq.Header.Set("Authorization", "Bearer new-key")
	newResp := httptest.NewRecorder()
	router.ServeHTTP(newResp, newReq)
	if newResp.Code != http.StatusOK {
		t.Fatalf("new key status = %d, want %d", newResp.Code, http.StatusOK)
	}

	data, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", envPath, err)
	}
	if string(data) != "API_KEY=new-key\n" {
		t.Fatalf("env file = %q, want %q", string(data), "API_KEY=new-key\n")
	}
}
