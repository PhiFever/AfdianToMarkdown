package afdian

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 服务器前两次返回 500，第三次成功；验证失败后会自动重试
func TestNewRequestGet_RetriesOnServerError(t *testing.T) {
	// 缩短重试间隔与次数，加速测试
	oldDelay, oldAttempts := RetryDelay, RetryAttempts
	RetryDelay, RetryAttempts = time.Millisecond, 3
	defer func() { RetryDelay, RetryAttempts = oldDelay, oldAttempts }()

	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&hits, 1) < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	body, err := NewRequestGet("example.com", srv.URL, "", srv.URL)
	assert.NoError(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&hits), "应在第 3 次成功前重试 2 次")
	assert.Contains(t, string(body), `"ok":true`)
}

// 服务器返回 404；验证 4xx 客户端错误立即失败、不重试
func TestNewRequestGet_NoRetryOnClientError(t *testing.T) {
	oldDelay, oldAttempts := RetryDelay, RetryAttempts
	RetryDelay, RetryAttempts = time.Millisecond, 3
	defer func() { RetryDelay, RetryAttempts = oldDelay, oldAttempts }()

	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := NewRequestGet("example.com", srv.URL, "", srv.URL)
	assert.Error(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&hits), "4xx 应立即失败，不应重试")
}

// 服务器响应慢于超时阈值；验证超时被触发、尽快返回错误，而非无限挂起
func TestNewRequestGet_TimesOut(t *testing.T) {
	oldTimeout, oldDelay, oldAttempts := RequestTimeout, RetryDelay, RetryAttempts
	RequestTimeout, RetryDelay, RetryAttempts = 50*time.Millisecond, time.Millisecond, 1
	defer func() {
		RequestTimeout, RetryDelay, RetryAttempts = oldTimeout, oldDelay, oldAttempts
	}()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		_, _ = w.Write([]byte("too late"))
	}))
	defer srv.Close()

	start := time.Now()
	_, err := NewRequestGet("example.com", srv.URL, "", srv.URL)
	elapsed := time.Since(start)

	assert.Error(t, err, "慢响应应触发超时错误")
	assert.Less(t, elapsed, 400*time.Millisecond, "应在超时后尽快返回，而非等待完整响应")
}
