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
	now           func() time.Time

	mu     sync.Mutex
	daily  map[string]dailyCounter
	logger *slog.Logger
}

type dailyCounter struct {
	day   string
	count int
}

func NewAIRateLimiter(anonMaxConcurrent, anonDailyLimit int) *AIRateLimiter {
	var sem chan struct{}
	if anonMaxConcurrent > 0 {
		sem = make(chan struct{}, anonMaxConcurrent)
	}

	return &AIRateLimiter{
		anonSemaphore: sem,
		dailyLimit:    anonDailyLimit,
		now:           time.Now,
		daily:         make(map[string]dailyCounter),
		logger:        slog.Default(),
	}
}

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

		// 同時実行制限
		if !l.acquireAnonymousSlot(w, r) {
			return

		}
		// 開いたら解放
		defer l.releaseAnonymousSlot()
		// クライアントのIDを取得
		clientID := anonymousClientID(r)
		if !l.allowDaily(w, r, clientID) {
			return
		}

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

func (l *AIRateLimiter) allowDaily(w http.ResponseWriter, r *http.Request, clientID string) bool {
	// 制限がなければ、そおまま通す
	if l.dailyLimit <= 0 {
		return true
	}

	// 今日の日付を用意
	today := l.now().Format("2006-01-02")

	// ロックを取得
	l.mu.Lock()
	// マップにクライアントidを読み込む、ぞんざいしない場合は、ゼロ値（dailyCounter{day: "", count: 0}）
	counter := l.daily[clientID] // {"id": {day: 2026-01-02, count:0}}
	// dayにアクセス、初期値はからのデータなので、日付に現在の日付をいれて、初期化
	if counter.day != today {
		counter = dailyCounter{day: today}
	}

	// マップからcountプロパティにアクセスし、上限を確認、もし上限を越した場合、ロックを解放し、エラーを返して、終了
	if counter.count >= l.dailyLimit {
		l.mu.Unlock()
		l.logger.Warn("anonymous AI request rejected by daily limit", "path", r.URL.Path)
		// このヘッダーは標準？
		w.Header().Set("Retry-After", secondsUntilTomorrow(l.now()))
		http.Error(w, "anonymous AI daily limit exceeded", http.StatusTooManyRequests)
		return false
	}
	// もし上限を越していなかったら、カウンタープロパティを１増やす
	counter.count++
	// 現在の残り回数を計算
	remaining := l.dailyLimit - counter.count
	// キーにマップを保存
	l.daily[clientID] = counter
	l.mu.Unlock()

	w.Header().Set("X-RateLimit-Limit", intToString(l.dailyLimit))
	w.Header().Set("X-RateLimit-Remaining", intToString(remaining))
	return true
}

func anonymousClientID(r *http.Request) string {
	ip := clientIP(r)
	// 戻ってきたipアドレスを２進数のバイトに変換し、256bit、３２バイト、1文字４バイト分の長さ、１６進数の文字列６４文字に変換？
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
