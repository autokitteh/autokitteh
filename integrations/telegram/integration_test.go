package telegram

import (
	"testing"

	"go.autokitteh.dev/autokitteh/integrations/common"
)

func TestNew(t *testing.T) {
	integration := New(nil)
	if integration == nil {
		t.Fatal("New() returned nil")
	}

	desc := integration.Get()
	if desc.UniqueName().String() != "telegram" {
		t.Errorf("Expected unique name 'telegram', got %q", desc.UniqueName().String())
	}

	if desc.DisplayName() != "Telegram" {
		t.Errorf("Expected display name 'Telegram', got %q", desc.DisplayName())
	}

	if desc.LogoURL().String() != "/static/images/telegram.svg" {
		t.Errorf("Expected logo URL '/static/images/telegram.svg', got %q", desc.LogoURL().String())
	}
}

func TestDescriptor(t *testing.T) {
	desc := common.Descriptor("telegram", "Telegram", "/static/images/telegram.svg")
	
	if desc.UniqueName().String() != "telegram" {
		t.Errorf("Expected unique name 'telegram', got %q", desc.UniqueName().String())
	}

	if desc.DisplayName() != "Telegram" {
		t.Errorf("Expected display name 'Telegram', got %q", desc.DisplayName())
	}

	if desc.LogoURL().String() != "/static/images/telegram.svg" {
		t.Errorf("Expected logo URL '/static/images/telegram.svg', got %q", desc.LogoURL().String())
	}
}
