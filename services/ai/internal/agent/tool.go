package agent

import "context"

// Tool はエージェントが使用するツールのインターフェース
type Tool interface {
	Name() string
	Description() string
	Run(ctx context.Context, input string) (string, error)
}
