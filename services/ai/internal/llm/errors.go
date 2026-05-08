package llm

import (
	"fmt"
	"net/http"
)

// HTTPError は LLM プロバイダーからのHTTPエラーを表す
// ステータスコードとレスポンス本文を保持する
type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("LLM API returned status %d: %s", e.StatusCode, e.Body)
}

// IsRateLimit は429レートリミットエラーかどうかを判定
func (e *HTTPError) IsRateLimit() bool {
	return e.StatusCode == http.StatusTooManyRequests
}

// IsUnauthorized は401認証エラーかどうかを判定
func (e *HTTPError) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// NewHTTPError はHTTPErrorを作成する
func NewHTTPError(statusCode int, body string) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Body:       body,
	}
}
