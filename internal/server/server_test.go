package server
package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNew(t *testing.T) {
	router := gin.New()
	srv := New("localhost:8080", router)

	assert.NotNil(t, srv)
	assert.NotNil(t, srv.httpServer)
	assert.Equal(t, "localhost:8080", srv.httpServer.Addr)
	assert.Equal(t, router, srv.httpServer.Handler)
}

func TestNewServerTimeouts(t *testing.T) {
	router := gin.New()
	srv := New("localhost:8080", router)

	assert.Equal(t, 10*time.Second, srv.httpServer.ReadTimeout)
	assert.Equal(t, 10*time.Second, srv.httpServer.WriteTimeout)
	assert.Equal(t, 5*time.Second, srv.httpServer.ReadHeaderTimeout)
	assert.Equal(t, 120*time.Second, srv.httpServer.IdleTimeout)
	assert.Equal(t, 1<<20, srv.httpServer.MaxHeaderBytes)
}

func TestNewWithDifferentAddresses(t *testing.T) {
	tests := []struct {
		name string
		addr string
	}{
		{
			name: "localhost with port",
			addr: "localhost:8080",
		},
		{
			name: "0.0.0.0 with port",
			addr: "0.0.0.0:3000",
		},
		{
			name: "127.0.0.1 with port",
			addr: "127.0.0.1:9000",
		},
		{
			name: "high port",
			addr: "localhost:65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			srv := New(tt.addr, router)

			assert.NotNil(t, srv)
			assert.Equal(t, tt.addr, srv.httpServer.Addr)
		})
	}
}

func TestShutdown(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	srv := New("localhost:0", router) // Port 0 for random port

	// Start a test server
	testServer := httptest.NewServer(srv.httpServer.Handler)
	defer testServer.Close()

	// Update server address to test server
	srv.httpServer.Addr = testServer.Listener.Addr().String()

	// Test shutdown
	err := srv.Shutdown()
	assert.NoError(t, err)
}

func TestShutdownWithContext(t *testing.T) {
	router := gin.New()
	srv := New("localhost:0", router)

	// Create a test server
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	// Shutdown should complete quickly
	done := make(chan error, 1)
	go func() {
		done <- srv.Shutdown()
	}()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("Shutdown took too long")
	}
}

func TestServerConfiguration(t *testing.T) {
	t.Run("ReadTimeout is set", func(t *testing.T) {
		router := gin.New()
		srv := New("localhost:8080", router)
		assert.Equal(t, 10*time.Second, srv.httpServer.ReadTimeout)
	})

	t.Run("WriteTimeout is set", func(t *testing.T) {
		router := gin.New()
		srv := New("localhost:8080", router)
		assert.Equal(t, 10*time.Second, srv.httpServer.WriteTimeout)
	})

	t.Run("ReadHeaderTimeout is set", func(t *testing.T) {
		router := gin.New()
		srv := New("localhost:8080", router)
		assert.Equal(t, 5*time.Second, srv.httpServer.ReadHeaderTimeout)
	})

	t.Run("IdleTimeout is set", func(t *testing.T) {
		router := gin.New()
		srv := New("localhost:8080", router)
		assert.Equal(t, 120*time.Second, srv.httpServer.IdleTimeout)
	})

	t.Run("MaxHeaderBytes is set", func(t *testing.T) {
		router := gin.New()
		srv := New("localhost:8080", router)
		assert.Equal(t, 1<<20, srv.httpServer.MaxHeaderBytes)
	})
}

func TestServerWithHandler(t *testing.T) {
	router := gin.New()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	srv := New("localhost:0", router)
	
	// Verify handler is set correctly
	assert.NotNil(t, srv.httpServer.Handler)
	
	// Test handler with httptest
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ping", nil)
	srv.httpServer.Handler.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "pong")
}

func TestServerStructure(t *testing.T) {
	router := gin.New()
	srv := New("localhost:8080", router)

	t.Run("Server has httpServer field", func(t *testing.T) {
		assert.NotNil(t, srv.httpServer)
	})

	t.Run("httpServer is configured", func(t *testing.T) {
		assert.NotEmpty(t, srv.httpServer.Addr)
		assert.NotNil(t, srv.httpServer.Handler)
		assert.NotZero(t, srv.httpServer.ReadTimeout)
		assert.NotZero(t, srv.httpServer.WriteTimeout)
	})
}

func TestServerTimeoutValues(t *testing.T) {
	router := gin.New()
	srv := New("localhost:8080", router)

	// Verify all timeouts are reasonable values
	assert.Greater(t, srv.httpServer.ReadTimeout, time.Duration(0))
	assert.Greater(t, srv.httpServer.WriteTimeout, time.Duration(0))
	assert.Greater(t, srv.httpServer.ReadHeaderTimeout, time.Duration(0))
	assert.Greater(t, srv.httpServer.IdleTimeout, time.Duration(0))

	// Verify ReadHeaderTimeout is less than ReadTimeout (security best practice)
	assert.Less(t, srv.httpServer.ReadHeaderTimeout, srv.httpServer.ReadTimeout)

	// Verify IdleTimeout is greater than WriteTimeout
	assert.Greater(t, srv.httpServer.IdleTimeout, srv.httpServer.WriteTimeout)
}

func BenchmarkNew(b *testing.B) {
	router := gin.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = New("localhost:8080", router)
	}
}

func BenchmarkShutdown(b *testing.B) {
	router := gin.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		srv := New("localhost:0", router)
		b.StartTimer()

		_ = srv.Shutdown()
	}
}

// TestServerIntegrationExample demonstrates how to test the server
// in an integration test scenario (not run by default)
func TestServerIntegrationExample(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "integration test")
	})

	// Create test server on random port
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	// Make request to test server
	resp, err := http.Get(testServer.URL + "/test")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
