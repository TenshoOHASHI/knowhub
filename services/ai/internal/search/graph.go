package search

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

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
	mu        sync.RWMutex
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
	kg.mu.Lock()
	defer kg.mu.Unlock()

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
	kg.mu.Lock()
	defer kg.mu.Unlock()

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
	kg.mu.RLock()
	defer kg.mu.RUnlock()

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
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	entities := make([]*Entity, 0, len(kg.entities))
	for _, e := range kg.entities {
		entities = append(entities, e)
	}
	return entities
}

// GetRelations は全リレーションを返す
func (kg *KnowledgeGraph) GetRelations() []Relation {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

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

// ============================================================
// GraphEngine: SearchEngine インターフェース実装
// ============================================================

// GraphEngine はナレッジグラフベースの検索エンジン
type GraphEngine struct {
	mu             sync.RWMutex
	graph          *KnowledgeGraph
	provider       llm.LLMProvider // エンティティ抽出用 LLM
	docs           map[string]Document // ID -> Document
	articleUpdated map[string]time.Time // 記事ID -> 最終更新時刻
}

// NewGraphEngine は GraphEngine のコンストラクタ
func NewGraphEngine(provider llm.LLMProvider) *GraphEngine {
	return &GraphEngine{
		graph:          NewKnowledgeGraph(),
		provider:       provider,
		docs:           make(map[string]Document),
		articleUpdated: make(map[string]time.Time),
	}
}

// GetGraph は GraphEngine の内部グラフを返す
func (e *GraphEngine) GetGraph() *KnowledgeGraph {
	return e.graph
}

// GetDocs は GraphEngine の保持するドキュメントを返す
func (e *GraphEngine) GetDocs() []Document {
	e.mu.RLock()
	defer e.mu.RUnlock()

	docs := make([]Document, 0, len(e.docs))
	for _, doc := range e.docs {
		docs = append(docs, doc)
	}
	return docs
}

// ============================================================
// 永続化: グラフをファイルに保存/読み込み
// ============================================================

// graphPersist は永続化用のデータ構造
type graphPersist struct {
	Entities       []persistEntity   `json:"entities"`
	Relations      []Relation        `json:"relations"`
	Docs           []Document        `json:"docs"`
	ArticleUpdated map[string]string `json:"article_updated"` // ISO8601形式
}

type persistEntity struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	ArticleIDs []string `json:"article_ids"`
}

// SaveGraph はグラフをJSONファイルに保存する
func (e *GraphEngine) SaveGraph(path string) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Entitiesをシリアライズ用に変換
	entities := make([]persistEntity, 0, len(e.graph.entities))
	for _, ent := range e.graph.entities {
		entities = append(entities, persistEntity{
			ID:         ent.ID,
			Name:       ent.Name,
			Type:       ent.Type,
			ArticleIDs: ent.ArticleIDs,
		})
	}

	// 記事をスライスに変換
	docs := make([]Document, 0, len(e.docs))
	for _, doc := range e.docs {
		docs = append(docs, doc)
	}

	// articleUpdatedをISO8601形式に変換
	articleUpdated := make(map[string]string)
	for id, t := range e.articleUpdated {
		articleUpdated[id] = t.Format(time.RFC3339)
	}

	data := graphPersist{
		Entities:       entities,
		Relations:      e.graph.relations,
		Docs:           docs,
		ArticleUpdated: articleUpdated,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal graph: %w", err)
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("write graph file: %w", err)
	}

	slog.Info("graph: saved to file", "path", path,
		"entities", len(entities),
		"relations", len(e.graph.relations),
		"docs", len(docs))

	return nil
}

// LoadGraph はJSONファイルからグラフを読み込む
func (e *GraphEngine) LoadGraph(path string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	jsonData, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("graph: no saved graph found", "path", path)
			return nil // ファイルがないのはエラーではない
		}
		return fmt.Errorf("read graph file: %w", err)
	}

	var data graphPersist
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("unmarshal graph: %w", err)
	}

	// グラフを復元
	e.graph = NewKnowledgeGraph()

	for _, ent := range data.Entities {
		e.graph.entities[ent.ID] = &Entity{
			ID:         ent.ID,
			Name:       ent.Name,
			Type:       ent.Type,
			ArticleIDs: ent.ArticleIDs,
		}
	}

	e.graph.relations = data.Relations

	// 隣接リストを再構築
	e.graph.adjacency = make(map[string][]string)
	for _, rel := range e.graph.relations {
		e.graph.adjacency[rel.Source] = append(e.graph.adjacency[rel.Source], rel.Target)
		e.graph.adjacency[rel.Target] = append(e.graph.adjacency[rel.Target], rel.Source)
	}

	// ドキュメントを復元
	e.docs = make(map[string]Document)
	for _, doc := range data.Docs {
		e.docs[doc.ID] = doc
	}

	// articleUpdatedを復元
	e.articleUpdated = make(map[string]time.Time)
	for id, ts := range data.ArticleUpdated {
		t, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			slog.Warn("failed to parse article timestamp", "id", id, "timestamp", ts)
			continue
		}
		e.articleUpdated[id] = t
	}

	slog.Info("graph: loaded from file", "path", path,
		"entities", len(e.graph.entities),
		"relations", len(e.graph.relations),
		"docs", len(e.docs))

	return nil
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

【重要】出力は日本語または英語のみを使用してください。中国語・韓国語などの他言語は使用禁止です。

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

	response, err := provider.GenerateWithOptions(ctx, prompt, llm.GenerateOptions{Format: "json"})

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
// 差分更新メソッド
// ============================================================

// UpdateArticle は単一記事のエンティティを抽出してグラフを更新する
func (e *GraphEngine) UpdateArticle(ctx context.Context, doc Document) error {
	if doc.Visibility != "public" {
		return nil // 非公開記事はスキップ
	}

	e.mu.Lock()

	// 古いエンティティを削除（この記事に関連するもの）
	e.removeArticleFromGraph(doc.ID)

	// 新しいエンティティを抽出
	result, err := extractEntities(ctx, e.provider, doc.Title, doc.Content)
	if err != nil {
		e.mu.Unlock()
		return fmt.Errorf("extract entities: %w", err)
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

	// ドキュメントと更新時刻を更新
	e.docs[doc.ID] = doc
	e.articleUpdated[doc.ID] = time.Now()

	e.mu.Unlock()

	slog.Info("graph: updated article",
		"id", doc.ID,
		"title", doc.Title,
		"entities", len(result.Entities),
		"relations", len(result.Relations))

	return nil
}

// RemoveArticle は記事をグラフから削除する
func (e *GraphEngine) RemoveArticle(articleID string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.removeArticleFromGraph(articleID)
	delete(e.docs, articleID)
	delete(e.articleUpdated, articleID)

	slog.Info("graph: removed article", "id", articleID)
}

// removeArticleFromGraph は記事に関連するエンティティ・リレーションを削除する（内部用、muは取得済みであること）
func (e *GraphEngine) removeArticleFromGraph(articleID string) {
	// 削除対象のエンティティIDを収集
	var entitiesToRemove []string
	for id, ent := range e.graph.entities {
		for _, aid := range ent.ArticleIDs {
			if aid == articleID {
				entitiesToRemove = append(entitiesToRemove, id)
				break
			}
		}
	}

	// エンティティから記事IDを削除（記事を持たないエンティティは削除）
	for _, id := range entitiesToRemove {
		if ent, ok := e.graph.entities[id]; ok {
			// 記事IDを削除
			var newArticleIDs []string
			for _, aid := range ent.ArticleIDs {
				if aid != articleID {
					newArticleIDs = append(newArticleIDs, aid)
				}
			}

			if len(newArticleIDs) == 0 {
				// 記事を持たないエンティティは削除
				delete(e.graph.entities, id)
				// 関連するリレーションも削除
				e.removeRelationsWithEntity(id)
			} else {
				ent.ArticleIDs = newArticleIDs
			}
		}
	}
}

// removeRelationsWithEntity は指定エンティティに関連するリレーションを削除する（内部用）
func (e *GraphEngine) removeRelationsWithEntity(entityID string) {
	var newRelations []Relation
	for _, rel := range e.graph.relations {
		if rel.Source != entityID && rel.Target != entityID {
			newRelations = append(newRelations, rel)
		}
	}
	e.graph.relations = newRelations

	// 隣接リストを再構築
	e.graph.adjacency = make(map[string][]string)
	for _, rel := range e.graph.relations {
		e.graph.adjacency[rel.Source] = append(e.graph.adjacency[rel.Source], rel.Target)
		e.graph.adjacency[rel.Target] = append(e.graph.adjacency[rel.Target], rel.Source)
	}
}

// ============================================================
// 差分更新対応の Index メソッド
// ============================================================

// Index はドキュメントを解析し、ナレッジグラフを差分更新する
// 新規・更新された記事のみ処理し、削除された記事はグラフから削除する
func (e *GraphEngine) Index(ctx context.Context, docs []Document) error {
	e.mu.Lock()

	// 現在の記事IDセットを構築
	currentDocIDs := make(map[string]bool)
	for _, doc := range docs {
		if doc.Visibility == "public" {
			currentDocIDs[doc.ID] = true
		}
	}

	// 削除された記事を特定
	var toRemove []string
	for id := range e.docs {
		if !currentDocIDs[id] {
			toRemove = append(toRemove, id)
		}
	}
	e.mu.Unlock()

	// 削除された記事をグラフから削除
	for _, id := range toRemove {
		e.RemoveArticle(id)
	}

	// 新規・更新された記事を処理
	var toProcess []Document

	e.mu.RLock()
	for _, doc := range docs {
		if doc.Visibility != "public" {
			continue
		}

		lastUpdated, exists := e.articleUpdated[doc.ID]
		if !exists {
			// 新規記事
			toProcess = append(toProcess, doc)
		} else {
			// 更新日時を比較（DocumentにUpdatedフィールドがある場合）
			// フィールドがない場合は常に処理
			if doc.UpdatedAt.IsZero() || doc.UpdatedAt.After(lastUpdated) {
				toProcess = append(toProcess, doc)
			}
		}
	}
	e.mu.RUnlock()

	slog.Info("graph: incremental update",
		"total_articles", len(docs),
		"new_or_updated", len(toProcess),
		"removed", len(toRemove))

	// 並列で処理
	semaphore := make(chan struct{}, 5)
	var wg sync.WaitGroup
	var processMu sync.Mutex
	processed := 0

	for _, doc := range toProcess {
		wg.Add(1)
		go func(d Document) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := e.UpdateArticle(ctx, d); err != nil {
				slog.Warn("failed to update article in graph",
					"id", d.ID, "title", d.Title, "error", err)
			}

			processMu.Lock()
			processed++
			processMu.Unlock()
		}(doc)
	}

	wg.Wait()

	slog.Info("graph: incremental update completed",
		"processed", processed,
		"total_entities", len(e.graph.entities),
		"total_relations", len(e.graph.relations))

	return nil
}

// ============================================================
// Search メソッド
// ============================================================

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
	titleMatchBonus := 10  // タイトル一致ボーナス
	partialMatchBonus := 5 // 部分一致ボーナス

	queryLower := normalizeEntityID(query)

	for _, eid := range queryEntityIDs {
		// クエリエンティティを含む記事にはボーナスを与える
		if entity, ok := e.graph.entities[eid]; ok {
			for _, aid := range entity.ArticleIDs {
				var title string
				if doc, ok := e.docs[aid]; ok {
					title = doc.Title
				}
				titleLower := normalizeEntityID(title)

				// 完全一致ボーナス
				if queryLower == titleLower {
					articleScores[aid] += titleMatchBonus
				} else if strings.Contains(titleLower, queryLower) || strings.Contains(queryLower, titleLower) {
					// 部分一致ボーナス（クエリがタイトルに含まれる、またはその逆）
					articleScores[aid] += partialMatchBonus
				}

				// エンティティ名との一致もチェック
				if strings.Contains(titleLower, eid) {
					articleScores[aid] += partialMatchBonus
				}
			}
		}

		related := e.graph.getRelatedArticles(eid, 2)
		for aid, score := range related {
			articleScores[aid] += score
		}
	}

	// ④ スコア降順で SearchResult を生成
	// docID → Document の map を作成
	results := make([]SearchResult, 0, len(articleScores))
	for aid, score := range articleScores {
		doc, ok := e.docs[aid]
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

	results := make([]SearchResult, 0, len(articleScores))
	for aid, score := range articleScores {
		doc, ok := e.docs[aid]
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
