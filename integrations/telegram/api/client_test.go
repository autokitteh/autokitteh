package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"unsafe"
)

func TestNewClient(t *testing.T) {
    client := NewClient("123456:ABC-DEF")
    if client == nil {
        t.Fatal("NewClient returned nil")
    }
}

// Helper function to set baseURL for testing using reflection
func setBaseURLForTesting(client *Client, url string) {
    v := reflect.ValueOf(client).Elem()
    baseURLField := v.FieldByName("baseURL")
    if !baseURLField.IsValid() {
        panic("baseURL field not found")
    }
    
    // Use unsafe to modify unexported field
    baseURLPtr := (*string)(unsafe.Pointer(baseURLField.UnsafeAddr()))
    *baseURLPtr = url
}

func TestGetMe(t *testing.T) {
    // Mock server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/getMe" {
            t.Errorf("Expected path '/getMe', got %q", r.URL.Path)
        }
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{
            "ok": true,
            "result": {
                "id": 123456789,
                "is_bot": true,
                "first_name": "Test Bot",
                "username": "testbot"
            }
        }`))
    }))
    defer server.Close()

    client := NewClient("test-token")
    setBaseURLForTesting(client, server.URL)

    ctx := context.Background()
    user, err := client.GetMe(ctx)
    if err != nil {
        t.Fatalf("GetMe failed: %v", err)
    }

    if user.ID != 123456789 {
        t.Errorf("Expected user ID 123456789, got %d", user.ID)
    }

    if !user.IsBot {
        t.Error("Expected user to be a bot")
    }

    if user.FirstName != "Test Bot" {
        t.Errorf("Expected first name 'Test Bot', got %q", user.FirstName)
    }

    if user.Username != "testbot" {
        t.Errorf("Expected username 'testbot', got %q", user.Username)
    }
}
func TestSendMessage(t *testing.T) {
    // Mock server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/sendMessage" {
            t.Errorf("Expected path '/sendMessage', got %q", r.URL.Path)
        }
        
        if r.Method != http.MethodPost {
            t.Errorf("Expected POST method, got %q", r.Method)
        }

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{
            "ok": true,
            "result": {
                "message_id": 1,
                "date": 1609459200,
                "chat": {
                    "id": 123,
                    "type": "private"
                },
                "text": "Hello, World!"
            }
        }`))
    }))
    defer server.Close()

    client := NewClient("test-token")
    setBaseURLForTesting(client, server.URL)

    ctx := context.Background()
    
    // Create SendMessageParams struct instead of individual parameters
    params := SendMessageParams{
        ChatID: 123,
        Text:   "Hello, World!",
    }
    
    message, err := client.SendMessage(ctx, params)
    if err != nil {
        t.Fatalf("SendMessage failed: %v", err)
    }

    if message.MessageID != 1 {
        t.Errorf("Expected message ID 1, got %d", message.MessageID)
    }

    if message.Text != "Hello, World!" {
        t.Errorf("Expected text 'Hello, World!', got %q", message.Text)
    }

    if message.Chat.ID != 123 {
        t.Errorf("Expected chat ID 123, got %d", message.Chat.ID)
    }
}