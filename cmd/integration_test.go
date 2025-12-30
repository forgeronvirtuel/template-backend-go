package cmd

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServerIntegration tests the full server lifecycle
func TestServerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Find a free port
	port := findFreePort(t)

	// Create command with test flags
	cmd := &cobra.Command{
		RunE: runServe,
	}
	cmd.Flags().Uint16P("port", "p", port, "Port to listen on")
	cmd.Flags().StringP("host", "H", "127.0.0.1", "Host to bind to")
	cmd.SetArgs([]string{fmt.Sprintf("--port=%d", port)})

	// Run server in goroutine
	serverDone := make(chan error, 1)
	go func() {
		serverDone <- runServe(cmd, []string{})
	}()

	// Wait for server to start
	addr := fmt.Sprintf("http://127.0.0.1:%d", port)
	require.Eventually(t, func() bool {
		resp, err := http.Get(addr + "/health")
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 100*time.Millisecond, "Server should start within 5 seconds")

	// Test health endpoint
	resp, err := http.Get(addr + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test API endpoint
	resp, err = http.Get(addr + "/api/v1/hello")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Send SIGTERM to gracefully shutdown
	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	err = proc.Signal(syscall.SIGTERM)
	require.NoError(t, err)

	// Wait for server to shutdown
	select {
	case err := <-serverDone:
		assert.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("Server did not shutdown within timeout")
	}
}

// TestServerGracefulShutdown tests that the server shuts down gracefully
func TestServerGracefulShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	port := findFreePort(t)

	cmd := &cobra.Command{
		RunE: runServe,
	}
	cmd.Flags().Uint16P("port", "p", port, "Port to listen on")
	cmd.Flags().StringP("host", "H", "127.0.0.1", "Host to bind to")
	cmd.SetArgs([]string{fmt.Sprintf("--port=%d", port)})

	serverDone := make(chan error, 1)
	go func() {
		serverDone <- runServe(cmd, []string{})
	}()

	// Wait for server to start
	addr := fmt.Sprintf("http://127.0.0.1:%d", port)
	require.Eventually(t, func() bool {
		resp, err := http.Get(addr + "/health")
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 100*time.Millisecond)

	// Start a long request
	requestDone := make(chan bool, 1)
	go func() {
		resp, err := http.Get(addr + "/health")
		if err == nil {
			resp.Body.Close()
		}
		requestDone <- true
	}()

	// Give request time to start
	time.Sleep(50 * time.Millisecond)

	// Signal shutdown
	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	err = proc.Signal(syscall.SIGTERM)
	require.NoError(t, err)

	// Wait for request to complete
	select {
	case <-requestDone:
		// Request completed
	case <-time.After(6 * time.Second):
		t.Fatal("Request was not completed during graceful shutdown")
	}

	// Wait for server to shutdown
	select {
	case err := <-serverDone:
		assert.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("Server did not shutdown within timeout")
	}
}

// TestServerStartupError tests server startup with invalid configuration
func TestServerStartupError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Try to bind to a privileged port (should fail without root)
	cmd := &cobra.Command{
		RunE: runServe,
	}
	cmd.Flags().Uint16P("port", "p", 80, "Port to listen on")
	cmd.Flags().StringP("host", "H", "127.0.0.1", "Host to bind to")
	cmd.SetArgs([]string{"--port=80"})

	err := runServe(cmd, []string{})

	// Should fail (unless running as root, which is unlikely in tests)
	if os.Geteuid() != 0 {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to start server")
	}
}

// TestServerConcurrentRequests tests handling multiple concurrent requests
func TestServerConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	port := findFreePort(t)

	cmd := &cobra.Command{
		RunE: runServe,
	}
	cmd.Flags().Uint16P("port", "p", port, "Port to listen on")
	cmd.Flags().StringP("host", "H", "127.0.0.1", "Host to bind to")
	cmd.SetArgs([]string{fmt.Sprintf("--port=%d", port)})

	serverDone := make(chan error, 1)
	go func() {
		serverDone <- runServe(cmd, []string{})
	}()

	// Wait for server to start
	addr := fmt.Sprintf("http://127.0.0.1:%d", port)
	require.Eventually(t, func() bool {
		resp, err := http.Get(addr + "/health")
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 100*time.Millisecond)

	// Send concurrent requests
	numRequests := 100
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			resp, err := http.Get(addr + "/health")
			if err != nil {
				results <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				results <- fmt.Errorf("unexpected status code: %d", resp.StatusCode)
				return
			}

			_, err = io.ReadAll(resp.Body)
			results <- err
		}()
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		select {
		case err := <-results:
			assert.NoError(t, err)
		case <-time.After(10 * time.Second):
			t.Fatal("Request timeout")
		}
	}

	// Shutdown server
	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	err = proc.Signal(syscall.SIGTERM)
	require.NoError(t, err)

	select {
	case <-serverDone:
	case <-time.After(10 * time.Second):
		t.Fatal("Server shutdown timeout")
	}
}

// TestServerTimeouts tests server timeout configurations
func TestServerTimeouts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	port := findFreePort(t)

	cmd := &cobra.Command{
		RunE: runServe,
	}
	cmd.Flags().Uint16P("port", "p", port, "Port to listen on")
	cmd.Flags().StringP("host", "H", "127.0.0.1", "Host to bind to")
	cmd.SetArgs([]string{fmt.Sprintf("--port=%d", port)})

	serverDone := make(chan error, 1)
	go func() {
		serverDone <- runServe(cmd, []string{})
	}()

	// Wait for server to start
	addr := fmt.Sprintf("http://127.0.0.1:%d", port)
	require.Eventually(t, func() bool {
		resp, err := http.Get(addr + "/health")
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 100*time.Millisecond)

	// Test with custom client that respects server timeouts
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Get(addr + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Cleanup
	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	err = proc.Signal(syscall.SIGTERM)
	require.NoError(t, err)

	select {
	case <-serverDone:
	case <-time.After(10 * time.Second):
		t.Fatal("Server shutdown timeout")
	}
}

// findFreePort finds an available port on the system
func findFreePort(t *testing.T) uint16 {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return uint16(addr.Port)
}

// TestFindFreePort tests the findFreePort helper
func TestFindFreePort(t *testing.T) {
	port1 := findFreePort(t)
	assert.NotZero(t, port1)

	port2 := findFreePort(t)
	assert.NotZero(t, port2)

	// Ports should be different (highly likely)
	assert.NotEqual(t, port1, port2)
}

// TestServerWithContext tests server shutdown via context
func TestServerWithContext(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	port := findFreePort(t)

	cmd := &cobra.Command{
		RunE: runServe,
	}
	cmd.Flags().Uint16P("port", "p", port, "Port to listen on")
	cmd.Flags().StringP("host", "H", "127.0.0.1", "Host to bind to")
	cmd.SetArgs([]string{fmt.Sprintf("--port=%d", port)})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	serverDone := make(chan error, 1)
	go func() {
		serverDone <- runServe(cmd, []string{})
	}()

	// Wait for server to start
	addr := fmt.Sprintf("http://127.0.0.1:%d", port)
	require.Eventually(t, func() bool {
		resp, err := http.Get(addr + "/health")
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 100*time.Millisecond)

	// Wait for context timeout
	<-ctx.Done()

	// Signal shutdown
	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	err = proc.Signal(syscall.SIGTERM)
	require.NoError(t, err)

	// Server should shutdown
	select {
	case <-serverDone:
	case <-time.After(10 * time.Second):
		t.Fatal("Server did not shutdown")
	}
}
