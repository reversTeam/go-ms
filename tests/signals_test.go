package tests

import (
	"syscall"
	"testing"
	"time"

	"github.com/reversTeam/go-ms/core"
	"github.com/stretchr/testify/assert"
)

type mockServer struct {
	gracefulStopCalled bool
}

func (m *mockServer) GracefulStop() error {
	m.gracefulStopCalled = true
	return nil
}

func TestCatchStopSignals(t *testing.T) {
	server := &mockServer{}
	core.AddServerGracefulStop(server)

	done := core.CatchStopSignals()

	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)

	select {
	case <-done:
		assert.True(t, server.gracefulStopCalled, "Server's GracefulStop should be called")
	case <-time.After(time.Second * 5):
		t.Fatal("Timeout waiting for signal processing")
	}
}
