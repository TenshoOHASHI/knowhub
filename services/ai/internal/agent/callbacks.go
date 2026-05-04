package agent

import (
	"context"
	"log/slog"
)

// Callbacks はエージェント実行中のイベントコールバック
type Callbacks struct {
	OnToolStart func(ctx context.Context, toolName, input string)
	OnToolEnd   func(ctx context.Context, toolName, output string)
	OnLLMStart  func(ctx context.Context, step int)
	OnLLMEnd    func(ctx context.Context, step int, response string)
}

// NewLoggingCallbacks は slog でログ出力するコールバックを返す
func NewLoggingCallbacks() *Callbacks {
	return &Callbacks{
		OnToolStart: func(ctx context.Context, toolName, input string) {
			slog.Info("agent: tool start", "tool", toolName, "input", input)
		},
		OnToolEnd: func(ctx context.Context, toolName, output string) {
			slog.Info("agent: tool end", "tool", toolName, "output_len", len(output))
		},
		OnLLMStart: func(ctx context.Context, step int) {
			slog.Info("agent: LLM start", "step", step)
		},
		OnLLMEnd: func(ctx context.Context, step int, response string) {
			slog.Info("agent: LLM end", "step", step, "response_len", len(response))
		},
	}
}
