package search

import (
	"context"
	"math"
	"sort"
	"strings"
	"unicode"
)

func charType(r rune) int {
	switch {
	case unicode.Is(unicode.Han, r): // 漢字
		return 0
	case unicode.Is(unicode.Hiragana, r): // ひらがな
		return 1
	case unicode.Is(unicode.Katakana, r): // カタカナ
		return 2
	case unicode.Is(unicode.Latin, r): // アルファベット
		return 3
	case unicode.IsDigit(r): // 数字
		return 4
	default:
		return -1 // 区切り文字（スペース、句読点等）
	}
}

// 入力: "Go言語でgRPCを実装する"
// 出力: ["go", "言語", "で", "grpc", "を", "実装", "する"]
func tokenizeOld(text string) []string {
	// 小文字に分割
	text = strings.ToLower(text)
	// 区切り文字で分割(指定した関数がtrueを返す文字で分割)
	// 文字と数値以外のもの、スペース、句読点で分割
	tokens := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	return tokens
}

func tokenize(text string) []string {
	text = strings.ToLower(text)
	var tokens []string
	var current []rune
	var prevType int = -2 // 初期値（未設定）

	for _, r := range text {
		ct := charType(r)

		// 区切り文字 → ここでトークンを確定
		if ct == -1 {
			if len(current) > 0 {
				tokens = append(tokens, string(current))
				current = nil
			}
			prevType = -2
			continue
		}

		// 文字種が変わった → トークンを確定して新しいトークン開始
		if ct != prevType && len(current) > 0 {
			// 一時保存しているテキストを保存
			tokens = append(tokens, string(current))
			// 初期化
			current = nil
		}

		// 現在のテキストを一時保存
		current = append(current, r)
		// 現在のトークンのタイプを更新
		prevType = ct
	}

	// 最後のトークン
	if len(current) > 0 {
		tokens = append(tokens, string(current))
	}

	return tokens
}

// 単語の出現頻度
// 　各単語の出現頻度 / 全体の総単語数　の単語が占める全体の割合
func computeTF(tokens []string) map[string]float64 {
	tf := make(map[string]float64)

	// count
	for _, token := range tokens {
		tf[token]++
	}

	total := float64(len(tokens))
	// 単語とカウントを取り出す
	for word, count := range tf {
		tf[word] = count / total
	}
	return tf
}

// 逆文章頻度
// たくさんの文章に出現する単語は、IDFが低い（価値が低い）、逆に少数の文章にしか出現しない単語は　IDFが高い（その文章の特徴づける単語）
func computeIDF(tokenizedDoc [][]string) map[string]float64 {
	N := float64(len(tokenizedDoc))
	df := make(map[string]float64) // 各単語が出現する文章数をカウント

	// 各文章を走査して、単語が出現する文章数をカウント
	for _, tokens := range tokenizedDoc {
		// 文章毎に初期化
		seen := make(map[string]bool) // 同じ文章内で重複カウントしない
		// 各文章毎に、出現した回数を記録
		for _, token := range tokens {

			// まだ存在しない場合は、保存する
			if !seen[token] {
				df[token]++
				seen[token] = true
			}
		}
	}

	idf := make(map[string]float64)
	for word, docCount := range df {
		// どれだけレアなのか、逆数で計算。
		// その単語が文章に出現する回数は、どのくらいなのか、レア度
		idf[word] = math.Log((N + 1) / (docCount + 1)) // log(N) - log(docCount)

	}

	return idf
}

// 一意のインデックスを付与（数値ベクトルに変換するため）
// データ構造：string文字列 -> [] １つの文章　-> [] 複数の文章 ->> [["Go", "grpc"], ["python", "AI"]
func buildVocabulary(tokenizedDoc [][]string) map[string]int {

	vocab := make(map[string]int)
	idx := 0
	// 複数の文章から、１つの文章を取り出す
	for _, doc := range tokenizedDoc {
		// １つの文章から、各単語を取り出す
		for _, token := range doc {
			// 取り出した単語を、マップのキーに保存、もしマップに存在しなければ、キーと値を保存
			if _, exits := vocab[token]; !exits {
				vocab[token] = idx
				idx++
			}
		}
	}

	return vocab
}

// TF = {"go": 0.5, "grpc": 0.25}
// IDF = {"go": 0.0, "grpc": 0.693}
func buildTFIDFVector(tf map[string]float64, idf map[string]float64, vocab map[string]int) []float64 {
	vec := make([]float64, len(vocab))

	for word, tfVal := range tf {
		// 単語のインデックス番号を取り出す
		if idx, ok := vocab[word]; ok {
			// 1つの文章にベクトルを構築
			vec[idx] = tfVal * idf[word]
		}
	}
	return vec
}

// ２つのベクトルの類似度を計算
// cos(A, B) = (A*B) (||A|| * ||B||)
// 内積とL2ノルム（ベクトルの長さ）

func cosineSimilarity(a, b []float64) float64 {
	var dotProduct float64
	for i := range a {
		dotProduct += a[i] * b[i]
	}
	// ||A||
	var normA float64
	// 累乗和
	for _, v := range a {
		normA += v * v
	}
	normA = math.Sqrt(normA)

	// ||B||
	var normB float64
	for _, v := range b {
		normB += v * v
	}
	normB = math.Sqrt(normB)

	if normA == 0 || normB == 0 {
		return 0.0
	}
	return dotProduct / (normA * normB)
}

type TFIDFEngine struct {
	documents  []Document         // 元の文書データ
	vocabulary map[string]int     // 単語 → インデックス
	idf        map[string]float64 // IDF値（全文書で共通）
	tfidfVecs  [][]float64        // 文書ごとのTF-IDFベクトル
	tokenized  [][]string         // 各文書のトークン（BM25で再利用）
}

func NewTFIDFEngine() *TFIDFEngine {
	return &TFIDFEngine{}
}

func (e *TFIDFEngine) Index(ctx context.Context, docs []Document) error {
	// 1. 元データを保存
	e.documents = docs

	// 2. 各文書をトークン化([["ai", "python", "go"], ["python", "react", "go"],])
	e.tokenized = make([][]string, len(docs))
	for i, doc := range docs {
		// タイトル + 内容 を両方検索対象にする
		text := doc.Title + " " + doc.Content
		e.tokenized[i] = tokenize(text)
	}

	// 3. 語彙を構築（["ai":0, "python": 1, "go": 2]）
	e.vocabulary = buildVocabulary(e.tokenized)

	// 4. IDF を計算
	e.idf = computeIDF(e.tokenized)

	// 5. 各文書の TF-IDF ベクトルを構築
	e.tfidfVecs = make([][]float64, len(docs))
	for i, tokens := range e.tokenized {
		// 単語の頻度
		tf := computeTF(tokens)
		// i番目の文章のトークンをi番目の場所に入れる
		e.tfidfVecs[i] = buildTFIDFVector(tf, e.idf, e.vocabulary)
	}

	return nil
}

func (e *TFIDFEngine) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// 1. クエリをトークン化
	queryTokens := tokenize(query)

	// クエリが空の場合は空結果を返す
	if len(queryTokens) == 0 {
		return []SearchResult{}, nil
	}

	// 2. クエリの TF を計算
	queryTF := computeTF(queryTokens)

	// 3. クエリの TF-IDF ベクトルを構築
	//    クエリにしか出現しない単語の IDF が未定義 → 0 になるが問題ない
	queryVec := buildTFIDFVector(queryTF, e.idf, e.vocabulary)

	// 4. 各文書との cosineSimilarity を計算
	results := make([]SearchResult, 0, len(e.documents))
	for i, doc := range e.documents {
		score := cosineSimilarity(queryVec, e.tfidfVecs[i])

		// スコアが 0 より大きいものだけ結果に含める
		if score > 0 {
			// コンテンツのスニペット（最初の200文字）
			snippet := doc.Content
			if len(snippet) > 200 {
				snippet = snippet[:200] + "..."
			}

			results = append(results, SearchResult{
				ArticleID:      doc.ID,
				Title:          doc.Title,
				Context:        snippet,
				RelevanceScore: score,
			})
		}
	}

	// 5. スコア降順でソート
	sort.Slice(results, func(i, j int) bool {
		return results[i].RelevanceScore > results[j].RelevanceScore
	})

	// 6. limit 件に切り詰め
	if limit < len(results) {
		results = results[:limit]
	}

	return results, nil
}
