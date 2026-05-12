package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type AIRateLimiter struct {
	anonSemaphore chan struct{}
	dailyLimit    int
	deepseekLimit int
	now           func() time.Time

	mu             sync.Mutex
	daily          map[string]dailyCounter
	deepseekDaily  map[string]dailyCounter
	logger         *slog.Logger
}

type dailyCounter struct {
	day   string
	count int
}

func NewAIRateLimiter(anonMaxConcurrent, anonDailyLimit, deepseekFreeDailyLimit int) *AIRateLimiter {
	var sem chan struct{}
	if anonMaxConcurrent > 0 {
		sem = make(chan struct{}, anonMaxConcurrent)
	}

	return &AIRateLimiter{
		anonSemaphore: sem,
		dailyLimit:    anonDailyLimit,
		deepseekLimit: deepseekFreeDailyLimit,
		now:           time.Now,
		daily:         make(map[string]dailyCounter),
		deepseekDaily: make(map[string]dailyCounter),
		logger:        slog.Default(),
	}
}

// Middleware は同時接続制限と日次上限チェック（カウントなし）を行う。
// 実際のカウントはハンドラーレベルで IncrementDaily / IncrementDeepSeekDaily を呼ぶ。
func (l *AIRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// AIと関係ないなら制限しない、また　CORS preflight なら制限しない（ブラウザーへの確認だけだから）
		if !strings.HasPrefix(r.URL.Path, "/api/ai/") || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// ログイン済みなら制限なしで返す
		if userID, ok := r.Context().Value("userID").(string); ok && userID != "" {
			next.ServeHTTP(w, r)
			return
		}

		// 同時実行制限のみ（日次制限はハンドラーレベルで実施）
		if !l.acquireAnonymousSlot(w, r) {
			return

		}
		// 開いたら解放
		defer l.releaseAnonymousSlot()

		next.ServeHTTP(w, r)
	})
}

func (l *AIRateLimiter) acquireAnonymousSlot(w http.ResponseWriter, r *http.Request) bool {
	// 同時実行制限がない場合、そのまま通す
	if l.anonSemaphore == nil {
		return true
	}

	select {
	// 空きがあれば、空の構造体を入れる
	case l.anonSemaphore <- struct{}{}:
		return true
	default:
		l.logger.Warn("anonymous AI request rejected by concurrency limit", "path", r.URL.Path)
		w.Header().Set("Retry-After", "30")
		http.Error(w, "too many anonymous AI requests; please try again later", http.StatusTooManyRequests)
		return false
	}
}

func (l *AIRateLimiter) releaseAnonymousSlot() {
	// そのものが作られていない場合は終了
	if l.anonSemaphore == nil {
		return
	}

	// データが存在する場合、空のデータを書き出す、データが入ってくるまで待機状態
	<-l.anonSemaphore
}

// CheckDaily は外部モデル用の日次上限チェックのみ行う（カウントしない）。
// 上限超過の場合は 429 を返して false を返す。
func (l *AIRateLimiter) CheckDaily(w http.ResponseWriter, clientID string) bool {
	// 制限がなければ、そのまま通す
	if l.dailyLimit <= 0 {
		return true
	}

	today := l.now().Format("2006-01-02")

	l.mu.Lock()
	counter := l.daily[clientID]
	if counter.day != today {
		counter = dailyCounter{day: today}
	}

	if counter.count >= l.dailyLimit {
		l.mu.Unlock()
		l.logger.Warn("anonymous AI request rejected by daily limit", "client_id", clientID)
		w.Header().Set("Retry-After", secondsUntilTomorrow(l.now()))
		http.Error(w, "anonymous AI daily limit exceeded", http.StatusTooManyRequests)
		return false
	}
	l.mu.Unlock()

	return true
}

// CheckDeepSeekDaily は DeepSeek Free 用の日次上限チェックのみ行う（カウントしない）。
// 上限超過の場合は 429 を返して false を返す。
func (l *AIRateLimiter) CheckDeepSeekDaily(w http.ResponseWriter, clientID string) bool {
	if l.deepseekLimit <= 0 {
		return true
	}

	today := l.now().Format("2006-01-02")

	l.mu.Lock()
	counter := l.deepseekDaily[clientID]
	if counter.day != today {
		counter = dailyCounter{day: today}
	}

	if counter.count >= l.deepseekLimit {
		l.mu.Unlock()
		l.logger.Warn("anonymous AI request rejected by DeepSeek free daily limit", "client_id", clientID)
		w.Header().Set("Retry-After", secondsUntilTomorrow(l.now()))
		http.Error(w, "DeepSeek free daily limit exceeded", http.StatusTooManyRequests)
		return false
	}
	l.mu.Unlock()

	return true
}

// IncrementDaily は外部モデル用の日次カウンターをインクリメントする。
func (l *AIRateLimiter) IncrementDaily(clientID string) {
	if l.dailyLimit <= 0 {
		return
	}

	today := l.now().Format("2006-01-02")

	l.mu.Lock()
	defer l.mu.Unlock()

	counter := l.daily[clientID]
	if counter.day != today {
		counter = dailyCounter{day: today}
	}
	counter.count++
	l.daily[clientID] = counter
}

// IncrementDeepSeekDaily は DeepSeek Free 用の日次カウンターをインクリメントする。
func (l *AIRateLimiter) IncrementDeepSeekDaily(clientID string) {
	if l.deepseekLimit <= 0 {
		return
	}

	today := l.now().Format("2006-01-02")

	l.mu.Lock()
	defer l.mu.Unlock()

	counter := l.deepseekDaily[clientID]
	if counter.day != today {
		counter = dailyCounter{day: today}
	}
	counter.count++
	l.deepseekDaily[clientID] = counter
}

// GetDailyLimit は外部モデル用の日次制限値を返す。
func (l *AIRateLimiter) GetDailyLimit() int {
	return l.dailyLimit
}

// GetDeepSeekDailyLimit は DeepSeek Free 用の日次制限値を返す。
func (l *AIRateLimiter) GetDeepSeekDailyLimit() int {
	return l.deepseekLimit
}

// GetRemainingDaily は外部モデル用の残り回数を返す。
func (l *AIRateLimiter) GetRemainingDaily(clientID string) int {
	if l.dailyLimit <= 0 {
		return -1 // 制限なし
	}
	today := l.now().Format("2006-01-02")
	l.mu.Lock()
	defer l.mu.Unlock()
	counter := l.daily[clientID]
	if counter.day != today {
		return l.dailyLimit
	}
	remaining := l.dailyLimit - counter.count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetRemainingDeepSeekDaily は DeepSeek Free 用の残り回数を返す。
func (l *AIRateLimiter) GetRemainingDeepSeekDaily(clientID string) int {
	if l.deepseekLimit <= 0 {
		return -1 // 制限なし
	}
	today := l.now().Format("2006-01-02")
	l.mu.Lock()
	defer l.mu.Unlock()
	counter := l.deepseekDaily[clientID]
	if counter.day != today {
		return l.deepseekLimit
	}
	remaining := l.deepseekLimit - counter.count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// AnonymousClientID はリクエストから匿名クライアントIDを生成する。
func AnonymousClientID(r *http.Request) string {
	ip := clientIP(r)
	sum := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(sum[:])[:16]
}

// リクエストを受け取る
func clientIP(r *http.Request) string {
	// nginxが送ったヘッダー情報から、クライアントのipアドレスを取得
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		// ヘッダー情報の中身は、マップで値が配列、だから、splitでカンマ区切りで分割して、ipアドレスの中身だけ取り出す
		parts := strings.Split(forwardedFor, ",")
		//　前後空白を削除し、ipアドレスがからでなければ、それを返す
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}

	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}

	// リモート先のアドレスにポートが付いている場合、hostだけ取り出す。
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}

	return r.RemoteAddr
}

func secondsUntilTomorrow(now time.Time) string {
	// 明日０時
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	// 現在の時刻から、明日の日付まで何秒か
	seconds := int(tomorrow.Sub(now).Seconds())
	// 保険、マイナスなど
	if seconds < 1 {
		seconds = 1
	}
	// 数値を文字列に変換
	return intToString(seconds)
}

func intToString(n int) string {
	return strconv.Itoa(n)
}
