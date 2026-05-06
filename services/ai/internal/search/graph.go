package search

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/TenshoOHASHI/knowhub/services/ai/internal/llm"
)

// ============================================================
// データ構造: ナレッジグラフのノード・エッジ
// ============================================================

// Entity はグラフの「ノード」: 記事から抽出した固有名詞や概念
// 例: "gRPC", "Protocol Buffers", "Google", "RPC"
type Entity struct {
	ID         string   // 正規化ID（小文字化など）
	Name       string   // 表示名
	Type       string   // 種別: "Technology", "Protocol", "Company" 等
	ArticleIDs []string // このエンティティを含む記事のID一覧
}

// Relation はグラフの「エッジ」: 2つのエンティティ間の関係
// 例: Source="gRPC", Target="Protocol Buffers", Label="uses"
type Relation struct {
	Source string // 始点エンティティID
	Target string // 終点エンティティID
	Label  string // 関係性: "uses", "developed_by", "implements" 等
}

// KnowledgeGraph はエンティティとリレーションを保持するグラフ構造
type KnowledgeGraph struct {
	entities  map[string]*Entity  // entityID → Entity
	relations []Relation          // 全リレーション
	adjacency map[string][]string // entityID → 隣接エンティティID一覧（BFS用）
}

func NewKnowledgeGraph() *KnowledgeGraph {
	return &KnowledgeGraph{
		entities:  make(map[string]*Entity),
		adjacency: make(map[string][]string),
	}
}

// addEntity はエンティティをグラフに追加（既存ならArticleIDを追記）
func (kg *KnowledgeGraph) addEntity(name, entityType, articleID string) {
	id := normalizeEntityID(name)
	if existing, ok := kg.entities[id]; ok {
		// 既存エンティティに記事IDを追加（重複防止）
		for _, aid := range existing.ArticleIDs {
			if aid == articleID {
				return
			}
		}
		existing.ArticleIDs = append(existing.ArticleIDs, articleID)
		return
	}
	kg.entities[id] = &Entity{
		ID:         id,
		Name:       name,
		Type:       entityType,
		ArticleIDs: []string{articleID},
	}
}

// addRelation はリレーションをグラフに追加（隣接リストも更新）
func (kg *KnowledgeGraph) addRelation(source, target, label string) {
	// 重複チェック
	for _, r := range kg.relations {
		if r.Source == source && r.Target == target && r.Label == label {
			return
		}
	}
	kg.relations = append(kg.relations, Relation{
		Source: source,
		Target: target,
		Label:  label,
	})

	// 隣接リスト（無向グラフとして両方向に追加）
	kg.adjacency[source] = append(kg.adjacency[source], target)
	kg.adjacency[target] = append(kg.adjacency[target], source)
}

// getRelatedArticles は指定エンティティから maxHops ホップ以内の
// 全関連エンティティが含まれる記事IDを収集する（BFS）
func (kg *KnowledgeGraph) getRelatedArticles(entityID string, maxHops int) map[string]int {
	visited := make(map[string]bool)
	articleScores := make(map[string]int) // articleID → その記事に含まれる関連エンティティ数

	queue := []struct {
		id   string
		hops int
	}{{id: entityID, hops: 0}}

	visited[entityID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// このエンティティを含む記事を収集
		if entity, ok := kg.entities[current.id]; ok {
			for _, aid := range entity.ArticleIDs {
				// ホップ数が小さいほど高スコア
				articleScores[aid] += maxHops - current.hops + 1
			}
		}

		// 最大ホップ数に達していれば隣接探索を終了
		if current.hops >= maxHops {
			continue
		}

		// 隣接ノードをキューに追加
		for _, neighbor := range kg.adjacency[current.id] {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, struct {
					id   string
					hops int
				}{id: neighbor, hops: current.hops + 1})
			}
		}
	}

	return articleScores
}

// GetEntities は全エンティティをスライスで返す
func (kg *KnowledgeGraph) GetEntities() []*Entity {
	entities := make([]*Entity, 0, len(kg.entities))
	for _, e := range kg.entities {
		entities = append(entities, e)
	}
	return entities
}

// GetRelations は全リレーションを返す
func (kg *KnowledgeGraph) GetRelations() []Relation {
	return kg.relations
}

// GetRelatedArticleIDs は指定記事に含まれるエンティティから BFS で関連記事IDを収集する
func (kg *KnowledgeGraph) GetRelatedArticleIDs(articleID string, maxHops int) map[string]int {
	articleScores := make(map[string]int)

	// 記事を含む全エンティティを起点に BFS
	for _, entity := range kg.entities {
		for _, aid := range entity.ArticleIDs {
			if aid == articleID {
				related := kg.getRelatedArticles(entity.ID, maxHops)
				for rid, score := range related {
					if rid != articleID { // 自分自身は除外
						articleScores[rid] += score
					}
				}
			}
		}
	}

	return articleScores
}

// GetGraph は GraphEngine の内部グラフを返す
func (e *GraphEngine) GetGraph() *KnowledgeGraph {
	return e.graph
}

// GetDocs は GraphEngine の保持するドキュメントを返す
func (e *GraphEngine) GetDocs() []Document {
	return e.docs
}

// normalizeEntityID はエンティティ名を正規化してIDにする
func normalizeEntityID(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// ============================================================
// LLM によるエンティティ・リレーション抽出
// ============================================================

// extractionResult は LLM からの JSON 抽出結果
type extractionResult struct {
	Entities []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"entities"`
	Relations []struct {
		Source string `json:"source"`
		Target string `json:"target"`
		Label  string `json:"label"`
	} `json:"relations"`
}

// extractEntities は LLM で記事からエンティティとリレーションを抽出する
func extractEntities(ctx context.Context, provider llm.LLMProvider, title, content string) (*extractionResult, error) {
	// コンテンツが長すぎる場合は切り詰める（トークン節約）
	text := title + "\n" + content
	if len(text) > 2000 {
		runes := []rune(text)
		if len(runes) > 2000 {
			text = string(runes[:2000])
		}
	}

	prompt := fmt.Sprintf(`以下の技術記事から、エンティティ（固有名詞・概念）とエンティティ間の関係を抽出してください。

記事:
%s

以下のJSON形式で出力してください。他の説明文は不要です。
{
  "entities": [
    {"name": "エンティティ名", "type": "種別(Technology/Protocol/Company/Concept/Language/Tool等)"}
  ],
  "relations": [
    {"source": "エンティティ名", "target": "エンティティ名", "label": "関係(uses/developed_by/implements/is_a/part_of等)"}
  ]
}`, text)

	response, err := provider.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM entity extraction failed: %w", err)
	}

	// JSON 部分（```json ... ```）を抽出
	jsonStr := extractJSON(response)

	var result extractionResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		slog.Warn("failed to parse entity extraction JSON", "error", err, "response", response)
		return nil, fmt.Errorf("JSON parse error: %w", err)
	}

	return &result, nil
}

// extractJSON は LLM レスポンスから JSON ブロックを取り出す
func extractJSON(response string) string {
	// ```json ... ``` パターン
	if idx := strings.Index(response, "```json"); idx != -1 {
		start := idx + 7
		if end := strings.Index(response[start:], "```"); end != -1 {
			return strings.TrimSpace(response[start : start+end])
		}
	}
	// ``` ... ``` パターン
	if idx := strings.Index(response, "```"); idx != -1 {
		start := idx + 3
		if end := strings.Index(response[start:], "```"); end != -1 {
			return strings.TrimSpace(response[start : start+end])
		}
	}
	// { ... } パターン（JSON ブロックを探す）
	if start := strings.Index(response, "{"); start != -1 {
		if end := strings.LastIndex(response, "}"); end > start {
			return response[start : end+1]
		}
	}
	return response
}

// ============================================================
// GraphEngine: SearchEngine インターフェース実装
// ============================================================

// GraphEngine はナレッジグラフベースの検索エンジン
type GraphEngine struct {
	graph    *KnowledgeGraph
	provider llm.LLMProvider // エンティティ抽出用 LLM
	docs     []Document      // 元ドキュメント（snippet 用）
}

// NewGraphEngine は GraphEngine のコンストラクタ
func NewGraphEngine(provider llm.LLMProvider) *GraphEngine {
	return &GraphEngine{
		graph:    NewKnowledgeGraph(),
		provider: provider,
	}
}

// Index は全ドキュメントを LLM で解析し、ナレッジグラフを構築する
func (e *GraphEngine) Index(ctx context.Context, docs []Document) error {
	e.docs = docs

	for _, doc := range docs {
		result, err := extractEntities(ctx, e.provider, doc.Title, doc.Content)
		if err != nil {
			slog.Warn("skipping entity extraction for article",
				"id", doc.ID, "title", doc.Title, "error", err)
			continue
		}

		// エンティティをグラフに追加
		for _, ent := range result.Entities {
			if ent.Name == "" {
				continue
			}
			e.graph.addEntity(ent.Name, ent.Type, doc.ID)
		}

		// リレーションをグラフに追加
		for _, rel := range result.Relations {
			if rel.Source == "" || rel.Target == "" {
				continue
			}
			sourceID := normalizeEntityID(rel.Source)
			targetID := normalizeEntityID(rel.Target)
			e.graph.addRelation(sourceID, targetID, rel.Label)
		}

		slog.Info("graph: extracted entities from article",
			"id", doc.ID,
			"title", doc.Title,
			"entities", len(result.Entities),
			"relations", len(result.Relations))
	}

	slog.Info("graph: knowledge graph built",
		"total_entities", len(e.graph.entities),
		"total_relations", len(e.graph.relations))

	return nil
}

// Search はクエリからエンティティを特定し、BFS で関連記事を収集する
func (e *GraphEngine) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// ① クエリからエンティティを抽出（LLM を使用）
	queryResult, err := extractEntities(ctx, e.provider, "", query)
	if err != nil {
		slog.Warn("failed to extract entities from query, falling back to keyword match", "error", err)
		// フォールバック: クエリトークンから直接エンティティを探す
		return e.searchByTokens(query, limit), nil
	}

	// ② 抽出したエンティティIDを正規化
	var queryEntityIDs []string
	for _, ent := range queryResult.Entities {
		if ent.Name != "" {
			id := normalizeEntityID(ent.Name)
			// グラフに存在するエンティティのみ対象
			if _, ok := e.graph.entities[id]; ok {
				queryEntityIDs = append(queryEntityIDs, id)
			}
		}
	}

	// エンティティが見つからない場合はトークンベースにフォールバック
	if len(queryEntityIDs) == 0 {
		return e.searchByTokens(query, limit), nil
	}

	// ③ 各エンティティから BFS（2-hop）で関連記事を収集
	articleScores := make(map[string]int)
	for _, eid := range queryEntityIDs {
		related := e.graph.getRelatedArticles(eid, 2)
		for aid, score := range related {
			articleScores[aid] += score
		}
	}

	// ④ スコア降順で SearchResult を生成
	// docID → Document の map を作成
	docMap := make(map[string]Document)
	for _, doc := range e.docs {
		docMap[doc.ID] = doc
	}

	results := make([]SearchResult, 0, len(articleScores))
	for aid, score := range articleScores {
		doc, ok := docMap[aid]
		if !ok {
			continue
		}

		snippet := doc.Content
		if len(snippet) > 200 {
			runes := []rune(snippet)
			if len(runes) > 200 {
				snippet = string(runes[:200]) + "..."
			}
		}

		results = append(results, SearchResult{
			ArticleID:      aid,
			Title:          doc.Title,
			Context:        snippet,
			RelevanceScore: float64(score),
		})
	}

	// ⑤ スコア降順ソート
	sort.Slice(results, func(i, j int) bool {
		return results[i].RelevanceScore > results[j].RelevanceScore
	})

	// ⑥ limit で切り詰め
	if limit < len(results) {
		results = results[:limit]
	}

	return results, nil
}

// searchByTokens は LLM 抽出に失敗した際のフォールバック:
// クエリをトークン化し、エンティティ名と部分一致するものを探す
func (e *GraphEngine) searchByTokens(query string, limit int) []SearchResult {
	tokens := tokenize(query)

	articleScores := make(map[string]int)
	for _, token := range tokens {
		// エンティティ名と部分一致するものを探す
		for id, entity := range e.graph.entities {
			if strings.Contains(entity.Name, token) || strings.Contains(id, token) {
				related := e.graph.getRelatedArticles(id, 2)
				for aid, score := range related {
					articleScores[aid] += score
				}
			}
		}
	}

	docMap := make(map[string]Document)
	for _, doc := range e.docs {
		docMap[doc.ID] = doc
	}

	results := make([]SearchResult, 0, len(articleScores))
	for aid, score := range articleScores {
		doc, ok := docMap[aid]
		if !ok {
			continue
		}

		snippet := doc.Content
		if len(snippet) > 200 {
			runes := []rune(snippet)
			if len(runes) > 200 {
				snippet = string(runes[:200]) + "..."
			}
		}

		results = append(results, SearchResult{
			ArticleID:      aid,
			Title:          doc.Title,
			Context:        snippet,
			RelevanceScore: float64(score),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].RelevanceScore > results[j].RelevanceScore
	})

	if limit < len(results) {
		results = results[:limit]
	}

	return results
}
