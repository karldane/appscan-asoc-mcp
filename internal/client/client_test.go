package client

import (
	"encoding/base64"
	"fmt"
	"testing"
)

func TestAuthHeader(t *testing.T) {
	c := New("https://cloud.appscan.com", "my-key-id", "my-key-secret", 30)
	header := c.AuthHeader()

	expected := base64.StdEncoding.EncodeToString([]byte("my-key-id:my-key-secret"))
	if header != fmt.Sprintf("Basic %s", expected) {
		t.Errorf("expected header 'Basic %s', got '%s'", expected, header)
	}
}

func TestAuthHeader_SpecialChars(t *testing.T) {
	c := New("https://cloud.appscan.com", "key+with/special:chars", "secret:with:special", 30)
	header := c.AuthHeader()

	expected := base64.StdEncoding.EncodeToString([]byte("key+with/special:chars:secret:with:special"))
	if header != fmt.Sprintf("Basic %s", expected) {
		t.Errorf("expected header 'Basic %s', got '%s'", expected, header)
	}
}
