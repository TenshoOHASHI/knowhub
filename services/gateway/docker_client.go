package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"regexp"
	"strings"

	"github.com/TenshoOHASHI/knowhub/services/gateway/handler"
)

// DockerClient はDockerコマンドを実行するクライアント
type DockerClient struct {
	composeFile string // docker-compose.yml のパス（例: docker-compose.prod.yml）
	envFile     string // --env-file で渡すファイル（例: .env.production）
}

// NOTE: ContainerJSON は handler.ContainerJSON と同じ構造。
// DockerClient は handler.DockerOperator インターフェースを満たすため、
// ListContainers の戻り値は handler.ContainerJSON を使用する。

// NewDockerClient はDockerClientを作成する
//
// 本番環境では以下のように設定する:
//
//	composeFile = "docker-compose.prod.yml"
//	envFile     = ".env.production"
//
// これにより実行されるコマンドは:
//
//	docker compose -f docker-compose.prod.yml --env-file .env.production logs ...
//
// 【--env-file が必要な理由】
//
//	docker-compose.prod.yml 内で ${MYSQL_ROOT_PASSWORD} や ${DOCKER_REGISTRY} など
//	環境変数の変数展開（interpolation）を使用している。
//	docker compose はデフォルトでカレントディレクトリの .env を読むが、
//	本番では .env.production という別名ファイルを使うため、明示的に指定する必要がある。
//	--env-file を指定しないと、変数が空になりイメージ名やDB接続情報が解決できない。
func NewDockerClient(composeFile, envFile string) *DockerClient {
	return &DockerClient{
		composeFile: composeFile,
		envFile:     envFile,
	}
}

// composeBaseArgs は docker compose コマンドの共通プレフィックス引数を構築する。
//
// 戻り値の例:
//
//	開発環境: ["compose"]
//	本番環境: ["compose", "-f", "docker-compose.prod.yml", "--env-file", ".env.production"]
func (d *DockerClient) composeBaseArgs() []string {
	args := []string{"compose"}
	if d.composeFile != "" {
		args = append(args, "-f", d.composeFile)
	}
	if d.envFile != "" {
		args = append(args, "--env-file", d.envFile)
	}
	return args
}

// 許可されたサービス名のパターン（英数字・ハイフン・アンダースコアのみ）
var validServiceName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// validateServices はサービス名をバリデーションする（コマンドインジェクション防止）
func validateServices(services []string) error {
	for _, s := range services {
		if !validServiceName.MatchString(s) {
			return fmt.Errorf("invalid service name: %q", s)
		}
	}
	return nil
}

// StreamLogs は指定サービスのログをストリーミングする。
// 返された io.ReadCloser を行単位で読み取ること。完了後は Close() すること。
//
// 実行されるコマンド例:
//
//	開発: docker compose logs -f --tail 100 --no-color ai gateway
//	本番: docker compose -f docker-compose.prod.yml --env-file .env.production logs -f --tail 100 --no-color ai gateway
//
// オプション解説:
//   - -f (--follow): ログ出力を継続的に追跡する（tail -f と同じ）
//   - --tail 100:    直近100行から表示を開始する
//   - --no-color:    ANSI カラーコードを除去（SSE で送信するため不要）
//
// 【exec.CommandContext vs exec.Command】
//   - exec.Command("docker", args...):        単純にコマンドを実行する
//   - exec.CommandContext(ctx, "docker", args...): context にキャンセルシグナルを紐付ける
//     → ctx がキャンセルされると、Go ランタイムがプロセスに SIGKILL を送る
//     → HTTPリクエストの context を渡すことで、クライアント切断時に自動停止する
//
// 【cmd.StdoutPipe() の仕組み】
//
//	StdoutPipe() は OS レベルのパイプ（pipe(2)）を作成する:
//	  docker プロセス → [パイプ書き込み端] ──── [パイプ読み取り端] → Go プログラム
//	戻り値の io.ReadCloser はパイプの読み取り端。
//	docker がログを stdout に書くたびに、Go 側で Read() できる。
//	これにより、全出力をメモリに溜めず、1行ずつストリーミング処理が可能になる。
//
//	比較: cmd.Output() は「プロセス終了まで全出力をメモリに蓄積」する。
//	ログストリーミングのように終了しないプロセスには使えない。
//
// 【cmd.Stderr = cmd.Stdout】
//
//	docker compose logs は一部のメッセージを stderr に出力することがある。
//	cmd.Stdout（StdoutPipe で作成したパイプの書き込み端）を cmd.Stderr にも設定すると、
//	stdout と stderr の両方が同じパイプに流れ、Go 側で一括して読み取れる。
func (d *DockerClient) StreamLogs(ctx context.Context, services []string) (io.ReadCloser, error) {
	if err := validateServices(services); err != nil {
		return nil, err
	}

	args := d.composeBaseArgs()
	args = append(args, "logs", "-f", "--tail", "100", "--no-color")
	args = append(args, services...)

	cmd := exec.CommandContext(ctx, "docker", args...)
	slog.Info("docker: streaming logs", "services", services, "cmd", cmd.String())

	// StdoutPipe: docker プロセスの stdout をパイプ経由で読み取るための io.ReadCloser を返す。
	// cmd.Start() の前に呼ぶ必要がある（Start 後はパイプを作成できない）。
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	// stderr も stdout と同じパイプにマージする。
	// docker compose logs はエラーメッセージを stderr に出力する場合があるため、
	// 両方を1つのストリームとして読み取る。
	cmd.Stderr = cmd.Stdout

	// Start: プロセスを非同期で起動する（完了を待たない）。
	// 比較: cmd.Run() = Start() + Wait()（完了まで待つ）。
	// ログストリーミングは終了しないプロセスなので Start() を使う。
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start docker logs: %w", err)
	}

	// goroutine で context のキャンセルを監視し、プロセスを kill する。
	// HTTPリクエストの context を使っているため、クライアントがSSE接続を切断すると
	// ctx.Done() が閉じ、docker logs プロセスが自動的に終了する。
	go func() {
		<-ctx.Done()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}()

	return stdout, nil
}

// ListContainers はDocker Composeサービスの一覧を取得する。
//
// 実行されるコマンド例:
//
//	開発: docker compose ps --format json -a
//	本番: docker compose -f docker-compose.prod.yml --env-file .env.production ps --format json -a
//
// オプション解説:
//   - --format json: 各コンテナの情報を1行1JSON で出力する
//   - -a (--all):    停止中のコンテナも含めて表示する
//
// 【cmd.Output() vs cmd.StdoutPipe()】
//
//	cmd.Output() = cmd.Start() + 全stdout読み取り + cmd.Wait()。
//	プロセスが終了するまで待ち、全出力を []byte で返す。
//	`docker compose ps` はすぐに終了するコマンドなので Output() が適切。
//	（StreamLogs のように終了しないコマンドには使えない）
func (d *DockerClient) ListContainers(ctx context.Context) ([]handler.ContainerJSON, error) {
	args := d.composeBaseArgs()
	args = append(args, "ps", "--format", "json", "-a")

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("docker compose ps: %w", err)
	}

	var containers []handler.ContainerJSON

	// docker compose ps --format json は1行ごとにJSONオブジェクトを出力する
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var raw struct {
			Name    string `json:"Name"`
			Service string `json:"Service"`
			State   string `json:"State"`
			Status  string `json:"Status"`
		}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			slog.Warn("docker: failed to parse container info", "line", line, "error", err)
			continue
		}
		containers = append(containers, handler.ContainerJSON{
			Name:    raw.Name,
			Service: raw.Service,
			State:   raw.State,
			Status:  raw.Status,
		})
	}

	return containers, nil
}

// 許可されたアクション（docker compose のサブコマンドとして実行されるもの）
// セキュリティ: ホワイトリスト方式。ここに登録されていないアクションは拒否される。
// 例: "restart" → `docker compose restart <service>`
var allowedActions = map[string]bool{
	"restart": true,
}

// 許可された exec コマンド（サービス名 → コンテナ内で実行するコマンド引数）
// セキュリティ: サービスごとに実行可能なコマンドを固定。
// ユーザー入力がコマンド引数に入ることはない（引数は全てハードコード）。
//
// 例: "nginx" → {"nginx", "-s", "reload"}
//
//	→ `docker compose exec nginx nginx -s reload`
//	→ nginx コンテナ内で `nginx -s reload` を実行
//
// 【nginx -s reload の意味】
//
//	nginx のマスタープロセスに SIGHUP シグナルを送り、設定ファイルを再読み込みする。
//	プロセスの再起動ではないので、ダウンタイムなしで設定変更を反映できる。
//	-s は signal の略。他に -s stop, -s quit, -s reopen がある。
var allowedExecCommands = map[string][]string{
	"nginx": {"nginx", "-s", "reload"},
}

// ExecCommand はDockerコンテナでコマンドを実行する（許可リスト方式）。
//
// 2種類のアクションをサポートする:
//
// 1. reload: コンテナ内でコマンドを実行する（docker compose exec）
//   - 実行例: docker compose exec nginx nginx -s reload
//   - 「docker compose exec」= 稼働中のコンテナ内でコマンドを実行する
//   - コンテナ自体は停止・再起動しない
//   - allowedExecCommands に登録されたサービスのみ実行可能
//
// 2. restart: コンテナを再起動する（docker compose restart）
//   - 実行例: docker compose restart ai
//   - コンテナを stop → start する
//   - allowedActions に登録されたアクションのみ実行可能
//
// 【cmd.CombinedOutput() の仕組み】
//
//	cmd.Output()         → stdout のみ取得、stderr は破棄
//	cmd.CombinedOutput() → stdout + stderr を結合して取得
//	docker コマンドはエラー情報を stderr に出力するため、CombinedOutput を使って
//	成功/失敗どちらの場合も全出力をキャプチャする。
//
// 【セキュリティモデル】
//
//	Web API からの任意コマンド実行はコマンドインジェクションのリスクがある。
//	このコードでは以下の多層防御を行っている:
//	  1. validateServices: サービス名を正規表現 ^[a-zA-Z0-9_-]+$ でバリデーション
//	  2. allowedActions / allowedExecCommands: 実行可能なコマンドをホワイトリストで制限
//	  3. exec.Command の引数分離: シェル経由ではなく直接実行するため、
//	     "nginx; rm -rf /" のようなシェルインジェクションは不可能
//	     （exec.Command は各引数を個別に execve(2) に渡す）
func (d *DockerClient) ExecCommand(ctx context.Context, service, action string) (string, error) {
	if err := validateServices([]string{service}); err != nil {
		return "", err
	}

	// --- reload: コンテナ内でコマンドを実行 ---
	// docker compose exec <service> <command...>
	if action == "reload" {
		execCmd, ok := allowedExecCommands[service]
		if !ok {
			return "", fmt.Errorf("reload not supported for service: %s", service)
		}

		// 構築されるコマンド例:
		//   docker compose -f docker-compose.prod.yml --env-file .env.production exec nginx nginx -s reload
		//                                                                             ^^^^^ ^^^^^^^^^^^^^^^^
		//                                                                         サービス名  コンテナ内コマンド
		args := d.composeBaseArgs()
		args = append(args, "exec", service)
		args = append(args, execCmd...)

		cmd := exec.CommandContext(ctx, "docker", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return string(output), fmt.Errorf("docker exec %s: %w", service, err)
		}
		return string(output), nil
	}

	// --- restart: コンテナを再起動 ---
	// docker compose restart <service>
	if allowedActions[action] {
		// 構築されるコマンド例: docker compose -f docker-compose.prod.yml --env-file .env.production restart ai
		args := d.composeBaseArgs()
		args = append(args, action, service)

		cmd := exec.CommandContext(ctx, "docker", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return string(output), fmt.Errorf("docker %s %s: %w", action, service, err)
		}
		return string(output), nil
	}

	return "", fmt.Errorf("action not allowed: %s", action)
}
