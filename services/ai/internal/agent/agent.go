package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/TenshoOHASHI/knowhub/services/ai/internal/llm"
)

// AgentResult はエージェントの実行結果
type AgentResult struct {
	Answer  string
	Steps   []AgentStepResult
	Sources []AgentSourceResult
}

// AgentStepResult は1ステップの実行記録
type AgentStepResult struct {
	Thought     string
	Action      string
	ActionInput string
	Observation string
}

// AgentSourceResult は参照元情報
type AgentSourceResult struct {
	ArticleID string
	Title     string
	URL       string
}

// Agent は ReAct パターンのエージェント
type Agent struct {
	provider  llm.LLMProvider
	tools     map[string]Tool
	maxIter   int
	callbacks *Callbacks
}

func NewAgent(provider llm.LLMProvider, tools []Tool, maxIter int, callbacks *Callbacks) *Agent {
	toolMap := make(map[string]Tool)
	for _, t := range tools {
		toolMap[t.Name()] = t
	}
	if maxIter <= 0 {
		maxIter = 10
	}
	return &Agent{
		provider:  provider,
		tools:     toolMap,
		maxIter:   maxIter,
		callbacks: callbacks,
	}
}

// Run は質問に対して ReAct ループを実行する
func (a *Agent) Run(ctx context.Context, question string) (*AgentResult, error) {
	toolDescriptions := a.buildToolDescriptions()

	toolGuidance := `
ツール選択のガイドライン:
- まず search_wiki でWiki内を検索してください
- search_wiki で見つけた記事は read_article で全文を取得してください。スニペットだけで回答してはいけません
- 同じツールを同じクエリで2回以上呼び出してはいけません
- read_article を使わずに Final Answer を出力してはいけません`

	if _, hasWeb := a.tools["web_search"]; hasWeb {
		toolGuidance += `
- search_wiki の後、必ず web_search でも検索してください。両方の結果を比較して回答すること
- web_search を使った後は、必ず read_url で上位の検索結果URLの本文を取得してください
- read_url を使わずに Final Answer を出力してはいけません`
	}

	jst := time.FixedZone("JST", 9*60*60)
	currentTime := time.Now().In(jst).Format("2006年1月2日 15:04 MST")

	systemPrompt := fmt.Sprintf(`あなたは技術ナレッジベースのアシスタントです。
現在の時刻: %s
質問に答えるために、利用可能なツールを使って情報を収集してください。
必ずツールを使ってから回答してください。ツールを使わずに回答してはいけません。

利用可能なツール:
%s
%s

必ず以下の形式を守ってください。例:

Thought: ユーザーの質問に関連する情報をWikiで検索する必要がある
Action: search_wiki
Action Input: {"query":"検索キーワード"}

この後、システムがツールを実行して "Observation: ..." を返します。
Observation の内容を読んで、次の行動を決めてください。
情報が足りなければ別のツールを呼び出してください。
情報が十分なら以下の形式で回答してください:

Final Answer: ユーザーへの回答文（日本語）

厳守ルール:
- 最初の出力は必ず Thought → Action → Action Input の形式にすること
- 回答を直接書いてはいけない。必ずツールを1回以上呼び出してから Final Answer を出力すること
- Action Input は {"key":"value"} のJSON形式にすること
- ツールは一度に1つだけ呼び出すこと
- Thought, Action, Action Input を1行ずつ出力すること`, currentTime, toolDescriptions, toolGuidance)

	// 会話履歴を構築
	var conversation []llm.Message
	conversation = append(conversation, llm.Message{Role: "system", Content: systemPrompt})
	conversation = append(conversation, llm.Message{Role: "user", Content: question})

	var steps []AgentStepResult
	var sources []AgentSourceResult

	// 同じアクションの繰り返し検出用
	seen := make(map[string]int) // "action:input" → 回数

	// 必須ツールの実行追跡
	usedTools := make(map[string]bool)
	_, hasReadArticle := a.tools["read_article"]
	_, hasWebSearch := a.tools["web_search"]

	for i := 0; i < a.maxIter; i++ {
		a.callbacks.OnLLMStart(ctx, i)

		response, err := a.provider.Chat(ctx, conversation)
		if err != nil {
			return nil, fmt.Errorf("LLM call failed at step %d: %w", i+1, err)
		}

		a.callbacks.OnLLMEnd(ctx, i, response)

		// Final Answer チェック
		if finalAnswer := parseFinalAnswer(response); finalAnswer != "" {
			// 必須ツールの実行チェック
			var missing []string
			if hasReadArticle && !usedTools["read_article"] {
				missing = append(missing, "read_article")
			}
			if hasWebSearch && usedTools["web_search"] && !usedTools["read_url"] {
				missing = append(missing, "read_url")
			}

			if len(missing) > 0 {
				// 必須ツールが未使用 → Final Answer をブロック
				forceMsg := fmt.Sprintf(
					"まだ %s を使っていません。ツールを使ってから Final Answer を出力してください。\nThought: もっと情報が必要\nAction: %s",
					strings.Join(missing, " と "),
					missing[0],
				)
				conversation = append(conversation, llm.Message{Role: "assistant", Content: response})
				conversation = append(conversation, llm.Message{Role: "user", Content: forceMsg})
				continue
			}

			return &AgentResult{
				Answer:  finalAnswer,
				Steps:   steps,
				Sources: sources,
			}, nil
		}

		// Thought / Action / Action Input をパース
		step := parseStep(response)

		// Action Input が空の場合、Thought から推測して JSON を生成
		if step.ActionInput == "" && step.Action != "" {
			step.ActionInput = inferActionInput(step.Action, step.Thought, question)
		}

		if step.Action == "" {
			// ツール呼び出しがない → LLM に再プロンプト
			conversation = append(conversation, llm.Message{Role: "assistant", Content: response})
			conversation = append(conversation, llm.Message{
				Role:    "user",
				Content: "ツールを呼び出してください。Thought → Action → Action Input の形式で出力してください。",
			})
			continue
		}

		tool, ok := a.tools[step.Action]
		if !ok {
			obs := fmt.Sprintf("エラー: ツール '%s' は存在しません。利用可能なツール: %s", step.Action, a.availableTools())
			step.Observation = obs
			steps = append(steps, step)
			conversation = append(conversation, llm.Message{Role: "assistant", Content: response})
			conversation = append(conversation, llm.Message{Role: "user", Content: "Observation: " + obs})
			continue
		}

		// ツール実行
		a.callbacks.OnToolStart(ctx, step.Action, step.ActionInput)
		obs, err := tool.Run(ctx, step.ActionInput)
		if err != nil {
			obs = fmt.Sprintf("エラー: %v", err)
		}
		a.callbacks.OnToolEnd(ctx, step.Action, obs)

		// 使用済みツールを記録
		usedTools[step.Action] = true

		// 同じアクションの繰り返し検出
		actionKey := step.Action + ":" + step.ActionInput
		seen[actionKey]++
		if seen[actionKey] > 1 {
			// 同じアクションを繰り返した → 別のツールに切り替えを促す
			if _, hasWeb := a.tools["web_search"]; hasWeb && step.Action == "search_wiki" {
				obs += "\n\n※ このクエリは既に実行済みです。同じ結果が返ります。web_search に切り替えて外部の情報を検索してください。"
			} else {
				obs += "\n\n※ このクエリは既に実行済みです。別のアクションを試すか、情報が十分なら Final Answer を出力してください。"
			}
		}

		step.Observation = obs
		steps = append(steps, step)

		// ソース収集
		collectSources(&sources, step)

		// 会話に追加
		conversation = append(conversation, llm.Message{Role: "assistant", Content: response})
		conversation = append(conversation, llm.Message{Role: "user", Content: "Observation: " + obs})
	}

	// max iterations に到達 → 最終回答を強制
	forcePrompt := "最大ステップ数に到達しました。これまでの情報を元に最終回答を出力してください。\nFinal Answer: ..."
	conversation = append(conversation, llm.Message{Role: "user", Content: forcePrompt})

	response, err := a.provider.Chat(ctx, conversation)
	if err != nil {
		return nil, fmt.Errorf("final LLM call failed: %w", err)
	}

	finalAnswer := parseFinalAnswer(response)
	if finalAnswer == "" {
		finalAnswer = response
	}

	return &AgentResult{
		Answer:  finalAnswer,
		Steps:   steps,
		Sources: sources,
	}, nil
}

func (a *Agent) buildToolDescriptions() string {
	var descs []string
	for _, t := range a.tools {
		descs = append(descs, fmt.Sprintf("- %s: %s", t.Name(), t.Description()))
	}
	return strings.Join(descs, "\n")
}

func (a *Agent) availableTools() string {
	var names []string
	for name := range a.tools {
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}

// --- パーサー ---

var (
	// (): 保存グループ、マッチしたら保存
	// (?i):命令形大文字小文字を無視
	// \s \S:空白、スペース、Tabキー（見えない文字を探す）、空白以外の文字（見える文字を探す）
	// .*?: なんでもいい1文字、０回以上の繰り返し（なんでもいい文字列、欲張りGreedyな性格で、大食い）、1文字つけるたびに、確認、最短でマッチ（遠慮がちな性格）
	// [\s\S]*: 改行も含めてすべての文字
	// (?:\n|$): キャプチャしないグループで、改行もしくは末尾（行の中身だけ欲しい、改行コードは必要ない）

	// 後半小文字大文字すべてマッチ、finalの後１つ上の空白、改行、タブの後コロンにマッチ、その後は、改行を含む全ての文字にマッチ、
	// Final Answer:という行を見つけて、その後ろにある内容（改行を含む長い文章すべて）抽出
	reFinalAnswer = regexp.MustCompile(`(?i)Final\s+Answer\s*:\s*([\s\S]*)`)
	reThought     = regexp.MustCompile(`(?i)Thought\s*:\s*(.*?)(?:\n|$)`)
	reAction      = regexp.MustCompile(`(?i)Action\s*:\s*(.*?)(?:\n|$)`)
	reActionInput = regexp.MustCompile(`(?is)Action\s+Input\s*:\s*(.*?)(?:\n(?:Thought|Action|Final)|$)`)
)

func parseFinalAnswer(text string) string {
	matches := reFinalAnswer.FindStringSubmatch(text)
	if len(matches) < 2 {
		return ""
	}
	return strings.TrimSpace(matches[1])
}

func parseStep(text string) AgentStepResult {
	step := AgentStepResult{}

	if m := reThought.FindStringSubmatch(text); len(m) > 1 {
		step.Thought = strings.TrimSpace(m[1])
	}
	if m := reAction.FindStringSubmatch(text); len(m) > 1 {
		step.Action = strings.TrimSpace(m[1])
	}
	if m := reActionInput.FindStringSubmatch(text); len(m) > 1 {
		step.ActionInput = strings.TrimSpace(m[1])
	}

	return step
}

func collectSources(sources *[]AgentSourceResult, step AgentStepResult) {
	// read_article の action_input から article_id を収集
	if step.Action == "read_article" {
		var in struct {
			ArticleID string `json:"article_id"`
		}
		if err := json.Unmarshal([]byte(step.ActionInput), &in); err == nil && in.ArticleID != "" {
			*sources = append(*sources, AgentSourceResult{ArticleID: in.ArticleID})
		}
	}

	// 重複チェック用ヘルパー
	exists := func(articleID, url string) bool {
		for _, s := range *sources {
			if articleID != "" && s.ArticleID == articleID {
				return true
			}
			if url != "" && s.URL == url {
				return true
			}
		}
		return false
	}

	// search_wiki の observation から記事IDとタイトルを抽出
	if step.Action == "search_wiki" {
		re := regexp.MustCompile(`\[\s*([^\]]+)\s*\]\s*\(ID:\s*([a-f0-9-]+)`)
		matches := re.FindAllStringSubmatch(step.Observation, -1)
		for _, m := range matches {
			if len(m) > 2 && !exists(m[2], "") {
				*sources = append(*sources, AgentSourceResult{
					ArticleID: m[2],
					Title:     strings.TrimSpace(m[1]),
				})
			}
		}
	}

	// web_search の observation から URL を収集
	if step.Action == "web_search" {
		re := regexp.MustCompile(`URL:\s*(https?://[^\s]+)`)
		matches := re.FindAllStringSubmatch(step.Observation, -1)
		for _, m := range matches {
			if len(m) > 1 && !exists("", m[1]) {
				*sources = append(*sources, AgentSourceResult{URL: m[1]})
			}
		}
	}

	// read_url の action_input から URL を収集
	if step.Action == "read_url" {
		var in struct {
			URL string `json:"url"`
		}
		if err := json.Unmarshal([]byte(step.ActionInput), &in); err == nil && in.URL != "" && !exists("", in.URL) {
			*sources = append(*sources, AgentSourceResult{URL: in.URL})
		}
	}
}

// truncate は文字列を maxRunes まで切り詰める
func truncate(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "\n... (truncated)"
}

// inferActionInput は Action Input が空のとき、元の質問から推測して JSON を生成する
func inferActionInput(action, thought, question string) string {
	// Thought は「検索する必要がある」などのメタ発言なので使わない
	// 元のユーザー質問をクエリとして使う
	query := question

	switch action {
	case "search_wiki", "web_search":
		return fmt.Sprintf(`{"query":"%s"}`, escapeJSON(query))
	case "read_article":
		return `{"article_id":""}`
	case "read_url":
		return `{"url":""}`
	case "list_articles":
		return ""
	default:
		return fmt.Sprintf(`{"query":"%s"}`, escapeJSON(query))
	}
}

// escapeJSON は JSON 文字列値として安全にエスケープする
func escapeJSON(s string) string {
	// 文字にダブル"があると、構造が壊れてしまうため、エスケープを追加
	s = strings.ReplaceAll(s, `"`, `\"`)
	// 改行がある場合は、半角スペースに置換する
	s = strings.ReplaceAll(s, "\n", " ")
	// 前後の空白行を削除
	s = strings.TrimSpace(s)
	// 長すぎる場合は最初の100文字に切り詰める
	runes := []rune(s)
	if len(runes) > 100 {
		s = string(runes[:100])
	}
	return s
}

// --- パイプライン実行 ---

// RunPipeline は固定順序でツールを自動実行し、最後にLLMで回答する
// search_wiki → read_article → (web_search → read_url) → LLM回答
func (a *Agent) RunPipeline(ctx context.Context, question string, history []llm.Message) (*AgentResult, error) {
	var steps []AgentStepResult
	var sources []AgentSourceResult
	var contextBuilder strings.Builder

	searchInput := fmt.Sprintf(`{"query":"%s"}`, escapeJSON(question))

	// Step 1: search_wiki（自動実行）
	if tool, ok := a.tools["search_wiki"]; ok {
		a.callbacks.OnToolStart(ctx, "search_wiki", searchInput)
		obs, err := tool.Run(ctx, searchInput)
		if err != nil {
			obs = fmt.Sprintf("エラー: %v", err)
		}
		a.callbacks.OnToolEnd(ctx, "search_wiki", obs)

		steps = append(steps, AgentStepResult{
			Thought:     "Wiki内を検索",
			Action:      "search_wiki",
			ActionInput: searchInput,
			Observation: obs,
		})
		collectSources(&sources, steps[len(steps)-1])

		// Step 2: read_article（検索結果の上位記事を自動取得）
		if readTool, ok := a.tools["read_article"]; ok {
			articleID := extractFirstArticleID(obs)
			if articleID != "" {
				articleInput := fmt.Sprintf(`{"article_id":"%s"}`, articleID)
				a.callbacks.OnToolStart(ctx, "read_article", articleInput)
				articleObs, err := readTool.Run(ctx, articleInput)
				if err != nil {
					articleObs = fmt.Sprintf("エラー: %v", err)
				}
				a.callbacks.OnToolEnd(ctx, "read_article", articleObs)

				contextBuilder.WriteString(fmt.Sprintf("## Wiki記事\n%s\n\n", articleObs))
				steps = append(steps, AgentStepResult{
					Thought:     "Wiki記事の全文を取得",
					Action:      "read_article",
					ActionInput: articleInput,
					Observation: truncate(articleObs, 500),
				})
				collectSources(&sources, steps[len(steps)-1])
			}
		}
	}

	// Step 3: web_search（有効な場合のみ自動実行）
	if webTool, ok := a.tools["web_search"]; ok {
		a.callbacks.OnToolStart(ctx, "web_search", searchInput)
		webObs, err := webTool.Run(ctx, searchInput)
		if err != nil {
			webObs = fmt.Sprintf("エラー: %v", err)
		}
		a.callbacks.OnToolEnd(ctx, "web_search", webObs)

		steps = append(steps, AgentStepResult{
			Thought:     "Webで検索",
			Action:      "web_search",
			ActionInput: searchInput,
			Observation: webObs,
		})
		collectSources(&sources, steps[len(steps)-1])

		// Step 4: read_url（上位結果のURLを自動取得）
		if urlTool, ok := a.tools["read_url"]; ok {
			topURL := extractFirstURL(webObs)
			if topURL != "" {
				urlInput := fmt.Sprintf(`{"url":"%s"}`, topURL)
				a.callbacks.OnToolStart(ctx, "read_url", urlInput)
				urlObs, err := urlTool.Run(ctx, urlInput)
				if err != nil {
					urlObs = fmt.Sprintf("エラー: %v", err)
				}
				a.callbacks.OnToolEnd(ctx, "read_url", urlObs)

				contextBuilder.WriteString(fmt.Sprintf("## Web情報\n%s\n\n", urlObs))
				steps = append(steps, AgentStepResult{
					Thought:     "Webページの本文を取得",
					Action:      "read_url",
					ActionInput: urlInput,
					Observation: truncate(urlObs, 500),
				})
				collectSources(&sources, steps[len(steps)-1])
			}
		}
	}

	// Step 5: LLMで最終回答を生成
	jst := time.FixedZone("JST", 9*60*60)
	currentTime := time.Now().In(jst).Format("2006年1月2日 15:04 MST")

	messages := []llm.Message{
		{
			Role: "system",
			Content: fmt.Sprintf(
				"あなたは技術ナレッジベースのアシスタントです。\n現在の時刻: %s\n"+
					"以下の情報を参考に、ユーザーの質問に日本語で丁寧に回答してください。"+
					"情報源の文体や一人称を真似してはいけません。独自の丁寧な口調で回答してください。",
				currentTime,
			),
		},
	}
	// 会話履歴を追加
	messages = append(messages, history...)
	// 収集した情報 + 質問
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: fmt.Sprintf("収集した情報:\n%s\n\n質問: %s", contextBuilder.String(), question),
	})

	a.callbacks.OnLLMStart(ctx, -1)
	answer, err := a.provider.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}
	a.callbacks.OnLLMEnd(ctx, -1, answer)

	return &AgentResult{
		Answer:  answer,
		Steps:   steps,
		Sources: sources,
	}, nil
}

// extractFirstArticleID は observation から最初の記事IDを抽出する
func extractFirstArticleID(obs string) string {
	re := regexp.MustCompile(`\(ID:\s*([a-f0-9-]+)`)
	m := re.FindStringSubmatch(obs)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}

// extractFirstURL は observation から最初のURLを抽出する
func extractFirstURL(obs string) string {
	re := regexp.MustCompile(`URL:\s*(https?://[^\s]+)`)
	m := re.FindStringSubmatch(obs)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}
