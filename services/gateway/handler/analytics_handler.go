package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"strings"

	pb "github.com/TenshoOHASHI/knowhub/proto/wiki"
)

type AnalyticsHandler struct {
	client pb.WikiServicesClient
}

func NewAnalyticsHandler(client pb.WikiServicesClient) *AnalyticsHandler {
	return &AnalyticsHandler{client: client}
}

// RecordPageView ビーコン受信
func (h *AnalyticsHandler) RecordPageView(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Path == "" {
		http.Error(w, "path is required", http.StatusBadRequest)
		return
	}

	// Hash IP for privacy. Prefer proxy headers added by Nginx/Next.js.
	ipHash := hashIP(clientIP(r))
	userAgent := r.Header.Get("User-Agent")
	if len(userAgent) > 500 {
		userAgent = userAgent[:500]
	}
	referrer := r.Header.Get("Referer")
	if len(referrer) > 500 {
		referrer = referrer[:500]
	}

	_, err := h.client.RecordPageView(r.Context(), &pb.RecordPageViewRequest{
		Path:      req.Path,
		IpHash:    ipHash,
		UserAgent: userAgent,
		Referrer:  referrer,
	})
	if err != nil {
		slog.Error("failed to record page view", "error", err, "path", req.Path)
		// Don't return error to client — analytics is non-critical
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetAnalyticsSummary アナリティクス概要
func (h *AnalyticsHandler) GetAnalyticsSummary(w http.ResponseWriter, r *http.Request) {
	days := int32(30)
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := parseInteger(d); err == nil && parsed > 0 {
			days = int32(parsed)
		}
	}

	resp, err := h.client.GetAnalyticsSummary(r.Context(), &pb.GetAnalyticsSummaryRequest{
		Days: days,
	})
	if err != nil {
		slog.Error("failed to get analytics summary", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func hashIP(ip string) string {
	h := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(h[:])
}

func clientIP(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func parseInteger(s string) (int, error) {
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		result = result*10 + int(c-'0')
	}
	return result, nil
}
