package slack_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/its-the-vibe/vibe-notify/internal/slack"
)

func TestPostMessage_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", ct)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := slack.NewClientWithHTTP(srv.URL, srv.Client())
	err := client.PostMessage(context.Background(), slack.Message{Text: "hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPostMessage_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer srv.Close()

	client := slack.NewClientWithHTTP(srv.URL, srv.Client())
	err := client.PostMessage(context.Background(), slack.Message{Text: "hello"})
	if err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}

func TestPostMessage_NetworkError(t *testing.T) {
	client := slack.NewClient("http://127.0.0.1:0/webhook") // nothing listening
	err := client.PostMessage(context.Background(), slack.Message{Text: "hello"})
	if err == nil {
		t.Fatal("expected error for network failure, got nil")
	}
}
