package compass

import (
	"testing"
)

func TestLocalIPv4(t *testing.T) {
	// Test will pass if function does not panic
	_, err := LocalIPv4()
	if err != nil {
		t.Error(err)
	}
}
