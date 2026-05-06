package search

import (
	"context"
	"math"
	"sort"
)

// BM25
// 文章の長さを正規化することで、公平にす（長い文章ほど、出現回数が多くなる）
type BM25Engine struct {
	documents  []Document         // 元の文書データ
	vocabulary map[string]int     // 単語 → インデックス
	idf        map[string]float64 // BM25版 IDF値
	tokenized  [][]string         // 各文書のトークン
	avgDl      float64            // 全文書の平均トークン数
	docLens    []int              // 各文書のトークン数
	k1         float64            // TF飽和パラメータ（頻度ｆの効き方を調整する、増えすぎないように抑制する）
	b          float64            // 文書長正規化パラメータ（長さ補正で平均長さに対して、ペナルティを追加）、またbは重み、75%の長さを信じる、25%は無視
}

func NewBM25Engine() *BM25Engine {
	return &BM25Engine{
		k1: 1.5,
		b:  0.75,
	}
}

// 生のカウントを使用します（IDFは全単語数で割る）
func computeTermFreq(tokens []string) map[string]int {
	freq := make(map[string]int)
	for _, token := range tokens {
		freq[token]++
	}
	return freq
}

// 出現した単語が各文章に出現する度合い（情報量の重み付け）
func computeBM25IDF(tokenizedDocs [][]string) map[string]float64 {
	// 文章数
	N := float64(len(tokenizedDocs))
	// 単語毎の出現数
	df := make(map[string]float64)

	for _, tokens := range tokenizedDocs {
		// 文章毎に初期化
		seen := make(map[string]bool)
		for _, token := range tokens {
			// 存在しない場合は、保存し、カウント
			if !seen[token] {
				df[token]++
				seen[token] = true
			}
		}
	}

	idf := make(map[string]float64)
	// n: その単語を含む文章の数
	for word, n := range df {
		idf[word] = math.Log((N-n+0.5)/(n+0.5) + 1)
	}
	return idf
}

func (e *BM25Engine) Index(ctx context.Context, docs []Document) error {
	// 1. 元データを保存
	e.documents = docs

	// 2. 各文書をトークン化 + 文書長を記録
	e.tokenized = make([][]string, len(docs))
	e.docLens = make([]int, len(docs))
	var totalTokens int
	for i, doc := range docs {
		text := doc.Title + " " + doc.Content
		e.tokenized[i] = tokenize(text)    // e.tokened: [][]string
		e.docLens[i] = len(e.tokenized[i]) // e.docLens: []string
		totalTokens += e.docLens[i]        // total = [1, 4, 4] => 9
	}

	// 3. 平均トークン数
	if len(docs) > 0 {
		// 全体の単語を、文書数で均等に割ったら1文書あたり何単語か
		// docsは全文章の単語数の合計
		e.avgDl = float64(totalTokens) / float64(len(docs))
	}

	// 4. 語彙を構築（tfidf.go の関数を再利用）
	e.vocabulary = buildVocabulary(e.tokenized)

	// 5. BM25版 IDF を計算
	e.idf = computeBM25IDF(e.tokenized)

	return nil
}

func (e *BM25Engine) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// ① クエリをトークン化
	queryTokens := tokenize(query)
	if len(queryTokens) == 0 {
		return []SearchResult{}, nil
	}

	// ③ 各文書とスコア計算 ← ここがメイン
	results := make([]SearchResult, 0, len(e.documents))
	for i, doc := range e.documents {
		// ③-a 文書の生出現回数
		docFreq := computeTermFreq(e.tokenized[i]) // i番目の文章を渡す

		// ③-b ここでスコア計算！
		var score float64
		for _, qToken := range queryTokens {
			f := float64(docFreq[qToken]) // その単語の文書内出現回数
			idfVal := e.idf[qToken]       // BM25版 IDF
			dl := float64(e.docLens[i])   // 文書のトークン数

			// 長い文章 => Lが大きい -> 分母が大きい -> スコア下がる
			numerator := f * (e.k1 + 1)                      // k1はどれくらい効かせるか
			denominator := f + e.k1*(1-e.b+e.b*(dl/e.avgDl)) // 「平均より長い？短い？」、長いペナルティ（bが長さを考慮し補正）
			score += idfVal * numerator / denominator
		}

		// ③-c スコアが 0 より大きければ結果に追加
		if score > 0 {
			snippet := doc.Content
			if len(snippet) > 200 {
				runes := []rune(snippet)
				if len(runes) > 200 {
					snippet = string(runes[:200]) + "..."
				}
			}
			results = append(results, SearchResult{
				ArticleID:      doc.ID,
				Title:          doc.Title,
				Context:        snippet,
				RelevanceScore: score,
			})
		}
	}

	// ④ ソート
	sort.Slice(results, func(i, j int) bool {
		return results[i].RelevanceScore > results[j].RelevanceScore
	})

	// ⑤ 切り詰め
	if limit < len(results) {
		results = results[:limit]
	}

	return results, nil
}
