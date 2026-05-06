package agent

import (
	"context"
	"log/slog"
	"sync/atomic"

	pb "github.com/TenshoOHASHI/knowhub/proto/ai"
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
			slog.Info("agent: LLM end", "step", step, "response_len", len(response), "response", response)
		},
	}
}

// NewStreamingCallbacks はログ出力 + gRPC stream.Send() を行うコールバックを返す
// 各コールバックが対応するイベントを sendEvent 経由で送信する
func NewStreamingCallbacks(sendEvent func(*pb.AgentStreamEvent) error) *Callbacks {
	var stepIndex int32

	return &Callbacks{
		OnLLMStart: func(ctx context.Context, step int) {
			slog.Info("agent: LLM start", "step", step)
			idx := atomic.AddInt32(&stepIndex, 1) - 1
			_ = sendEvent(&pb.AgentStreamEvent{
				EventType: "step",
				Step: &pb.AgentStepEvent{
					StepIndex: idx,
					Phase:     "llm_thinking",
				},
			})
		},
		OnLLMEnd: func(ctx context.Context, step int, response string) {
			slog.Info("agent: LLM end", "step", step, "response_len", len(response), "response", response)
		},
		OnToolStart: func(ctx context.Context, toolName, input string) {
			slog.Info("agent: tool start", "tool", toolName, "input", input)
			idx := atomic.AddInt32(&stepIndex, 1) - 1
			_ = sendEvent(&pb.AgentStreamEvent{
				EventType: "step",
				Step: &pb.AgentStepEvent{
					StepIndex:   idx,
					Action:      toolName,
					ActionInput: input,
					Phase:       "tool_executing",
				},
			})
		},
		OnToolEnd: func(ctx context.Context, toolName, output string) {
			slog.Info("agent: tool end", "tool", toolName, "output_len", len(output))
			idx := atomic.AddInt32(&stepIndex, 1) - 1
			_ = sendEvent(&pb.AgentStreamEvent{
				EventType: "step",
				Step: &pb.AgentStepEvent{
					StepIndex:   idx,
					Action:      toolName,
					Observation: output,
					Phase:       "tool_complete",
				},
			})
		},
	}
}
