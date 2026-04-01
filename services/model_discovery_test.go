package services

import (
	"context"
	"cursor2api-go/config"
	"cursor2api-go/utils"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/imroc/req/v3"
)

func TestDiscoverAvailableModelsParsesAllowedModelsFromUpstreamError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, `{"error":"Invalid model. Allowed models: google/gemini-3-flash, openai/gpt-5"}`)
	}))
	defer server.Close()

	originalURL := cursorAPIURL
	cursorAPIURL = server.URL
	defer func() {
		cursorAPIURL = originalURL
	}()

	service := &CursorService{
		config: &config.Config{
			Timeout:        5,
			MaxInputLength: 1000,
		},
		client:          req.C(),
		headerGenerator: utils.NewHeaderGenerator(),
	}

	got, err := service.DiscoverAvailableModels(context.Background())
	if err != nil {
		t.Fatalf("DiscoverAvailableModels() error = %v", err)
	}

	want := []string{"google/gemini-3-flash", "openai/gpt-5"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DiscoverAvailableModels() = %v, want %v", got, want)
	}
}

func TestDiscoverAvailableModelsReturnsErrorWhenUpstreamDoesNotExposeAllowedModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, `{"error":"Invalid model"}`)
	}))
	defer server.Close()

	originalURL := cursorAPIURL
	cursorAPIURL = server.URL
	defer func() {
		cursorAPIURL = originalURL
	}()

	service := &CursorService{
		config: &config.Config{
			Timeout:        5,
			MaxInputLength: 1000,
		},
		client:          req.C(),
		headerGenerator: utils.NewHeaderGenerator(),
	}

	if _, err := service.DiscoverAvailableModels(context.Background()); err == nil {
		t.Fatalf("DiscoverAvailableModels() error = nil, want parse failure")
	}
}
