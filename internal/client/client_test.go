package client

import (
	"testing"
)

func TestAuthHeader_KeyFormat(t *testing.T) {
	c := New("https://cloud.appscan.com", "my-key-id", "my-key-secret", 30)
	header := c.AuthHeader()

	if header != "my-key-id:my-key-secret" {
		t.Errorf("expected 'my-key-id:my-key-secret', got '%s'", header)
	}
}
