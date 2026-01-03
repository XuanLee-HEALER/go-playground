package e2e

import (
	"net/http"
	"testing"
	"time"
)

// TestHealthEndpoint verifies the container responds to /health.
func TestHealthEndpoint(t *testing.T) {
	client := &http.Client{Timeout: 3 * time.Second}

	resp, err := client.Get("http://localhost:8080/health")
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %v", resp.Status)
	}
}
