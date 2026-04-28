package slack_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/its-the-vibe/vibe-notify/internal/slack"
)

func TestPostMessage_Success(t *testing.T) {
	want := slack.Response{Channel: "C123", Ts: "1234567890.123456"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", ct)
		}
		if !strings.HasSuffix(r.URL.Path, "/message") {
			t.Errorf("expected path ending in /message, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	client := slack.NewClientWithHTTP(srv.URL, srv.Client())
	got, err := client.PostMessage(context.Background(), slack.Message{Text: "hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestPostMessage_WithMetadata(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding request body: %v", err)
		}
		if _, ok := body["metadata"]; !ok {
			t.Error("expected metadata field in request body")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(slack.Response{Channel: "C1", Ts: "1.0"})
	}))
	defer srv.Close()

	client := slack.NewClientWithHTTP(srv.URL, srv.Client())
	msg := slack.Message{
		Text: "hello",
		Metadata: map[string]interface{}{
			"event_type": "issue_broadcast",
		},
	}
	if _, err := client.PostMessage(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPostMessage_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer srv.Close()

	client := slack.NewClientWithHTTP(srv.URL, srv.Client())
	_, err := client.PostMessage(context.Background(), slack.Message{Text: "hello"})
	if err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}

func TestPostMessage_NetworkError(t *testing.T) {
	client := slack.NewClient("http://127.0.0.1:0") // nothing listening
	_, err := client.PostMessage(context.Background(), slack.Message{Text: "hello"})
	if err == nil {
		t.Fatal("expected error for network failure, got nil")
	}
}
