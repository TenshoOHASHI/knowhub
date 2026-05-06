package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAIRateLimiterRejectsConcurrentAnonymousRequest(t *testing.T) {
	limiter := NewAIRateLimiter(1, 0)

	entered := make(chan struct{})
	release := make(chan struct{})
	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-release
		w.WriteHeader(http.StatusNoContent)
	}))

	firstReq := httptest.NewRequest(http.MethodPost, "/api/ai/agent/stream", nil)
	firstReq.RemoteAddr = "203.0.113.10:12345"
	firstResp := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		handler.ServeHTTP(firstResp, firstReq)
		close(done)
	}()

	<-entered

	secondReq := httptest.NewRequest(http.MethodPost, "/api/ai/ask", nil)
	secondReq.RemoteAddr = "203.0.113.11:12345"
	secondResp := httptest.NewRecorder()
	handler.ServeHTTP(secondResp, secondReq)

	if secondResp.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", secondResp.Code)
	}

	close(release)
	<-done
}

func TestAIRateLimiterBypassesLoggedInUser(t *testing.T) {
	limiter := NewAIRateLimiter(1, 0)
	limiter.anonSemaphore <- struct{}{}

	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/ai/ask", nil)
	req = req.WithContext(context.WithValue(req.Context(), "userID", "owner"))
	//　HTTPレスポンスを記録するモック
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("expected logged-in request to bypass limiter, got %d", resp.Code)
	}
}

func TestAIRateLimiterRejectsAnonymousDailyLimit(t *testing.T) {
	limiter := NewAIRateLimiter(0, 2)
	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/ai/ask", nil)
		req.RemoteAddr = "203.0.113.10:12345"
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)
		if resp.Code != http.StatusNoContent {
			t.Fatalf("request %d: expected 204, got %d", i+1, resp.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/api/ai/ask", nil)
	req.RemoteAddr = "203.0.113.10:12345"
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after daily limit, got %d", resp.Code)
	}
}
