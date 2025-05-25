package events

import (
	"testing"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestWrapUpdate(t *testing.T) {
	update := &Update{
		UpdateID: 123,
		Message: &Message{
			MessageID: 1,
			Date:      1609459200,
			Text:      "Hello, World!",
			Chat: &Chat{
				ID:   456,
				Type: "private",
			},
			From: &User{
				ID:        789,
				IsBot:     false,
				FirstName: "John",
				Username:  "john_doe",
			},
		},
	}

	connectionID := sdktypes.NewConnectionID()
	event, err := WrapUpdate(update, connectionID)
	if err != nil {
		t.Fatalf("WrapUpdate failed: %v", err)
	}

	if event.Type() != "telegram_update" {
		t.Errorf("Expected event type 'telegram_update', got %q", event.Type())
	}

	if event.ConnectionID() != connectionID {
		t.Errorf("Expected connection ID %v, got %v", connectionID, event.ConnectionID())
	}

	// Check that the data contains the wrapped update
	data := event.Data()
	if data == nil {
		t.Fatal("Event data is nil")
	}

	// The data should be a map containing the update
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		t.Fatal("Event data is not a map")
	}

	if _, exists := dataMap["update"]; !exists {
		t.Error("Event data missing 'update' field")
	}
}

func TestWrapMessage(t *testing.T) {
	message := &Message{
		MessageID: 1,
		Date:      1609459200,
		Text:      "Hello, World!",
		Chat: &Chat{
			ID:   456,
			Type: "private",
		},
		From: &User{
			ID:        789,
			IsBot:     false,
			FirstName: "John",
			Username:  "john_doe",
		},
	}

	connectionID := sdktypes.NewConnectionID()
	event, err := WrapMessage(message, connectionID)
	if err != nil {
		t.Fatalf("WrapMessage failed: %v", err)
	}

	if event.Type() != "telegram_message" {
		t.Errorf("Expected event type 'telegram_message', got %q", event.Type())
	}

	if event.ConnectionID() != connectionID {
		t.Errorf("Expected connection ID %v, got %v", connectionID, event.ConnectionID())
	}

	// Check that the data contains the wrapped message
	data := event.Data()
	if data == nil {
		t.Fatal("Event data is nil")
	}

	// The data should be a map containing the message
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		t.Fatal("Event data is not a map")
	}

	if _, exists := dataMap["message"]; !exists {
		t.Error("Event data missing 'message' field")
	}
}
