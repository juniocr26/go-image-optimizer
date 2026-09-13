package httpserver

import (
	"testing"
	"time"

	"github.com/juniorosa/go-image-optimizer/backend/internal/config"
)

func TestNewServerAllowsSynchronousCodecWork(t *testing.T) {
	server := NewServer(config.Config{Port: "8080"}, testLogger())

	if server.ReadTimeout != 2*time.Minute {
		t.Fatalf("expected read timeout %s, got %s", 2*time.Minute, server.ReadTimeout)
	}

	if server.WriteTimeout != 2*time.Minute {
		t.Fatalf("expected write timeout %s, got %s", 2*time.Minute, server.WriteTimeout)
	}
}
