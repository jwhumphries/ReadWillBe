package components

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestPaginationButton_Accessibility(t *testing.T) {
	// render the component
	buf := new(bytes.Buffer)
	component := PaginationButton("test-id", "/test", 1, false, "Test Label")
	err := component.Render(context.Background(), buf)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	output := buf.String()

	// Check for aria-label
	if !strings.Contains(output, `aria-label="Test Label"`) {
		t.Errorf("expected aria-label='Test Label', got: %s", output)
	}

	// Check for data-tip
	if !strings.Contains(output, `data-tip="Test Label"`) {
		t.Errorf("expected data-tip='Test Label', got: %s", output)
	}

	// Check for tooltip class
	if !strings.Contains(output, `tooltip`) {
		t.Errorf("expected tooltip class, got: %s", output)
	}
}
