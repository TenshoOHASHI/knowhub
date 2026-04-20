package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/TenshoOHASHI/knowhub/services/wiki/internal/model"
	"github.com/redis/go-redis/v9"
)

const REDIS_CACHE_TIME = time.Minute * 10

// CommandRepository はWrite操作のみ
type ArticleQueryRepository interface {
	FindAll(ctx context.Context) ([]*model.Article, error)
	FindById(ctx context.Context, id string) (*model.Article, error)
}

type mysqlQueryRepository struct {
	rdb *redis.Client
	db  *sql.DB // フォールバック用（キャッシュミス時にMysqlから読む）
}

func NewMysqlQueryRepository(rdb *redis.Client, db *sql.DB) ArticleQueryRepository {
	return &mysqlQueryRepository{
		rdb: rdb,
		db:  db}
}

func (r *mysqlQueryRepository) FindById(ctx context.Context, id string) (*model.Article, error) {
	// redisから読み込む
	key := fmt.Sprintf("article:%s", id)
	data, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		// cache -> json => go struct
		var article model.Article
		json.Unmarshal([]byte(data), &article)
		return &article, nil
	}

	// キャッシュミス
	query := `SELECT id, title, content, created_at, updated_at From articles WHERE id=?`

	// 1件取得
	row := r.db.QueryRowContext(ctx, query, id)

	// 型を定義
	var article model.Article
	// DBデータを構造体にマッピング
	err = row.Scan(&article.ID, &article.Title, &article.Content, &article.CreatedAt, &article.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// 記事が見つかりません。
			return nil, sql.ErrNoRows
		}
		// 他のエラー
		return nil, err
	}

	// Redisにキャッシュ
	jsonData, _ := json.Marshal(article)
	r.rdb.Set(ctx, key, jsonData, REDIS_CACHE_TIME)

	return &article, nil
}

func (r *mysqlQueryRepository) FindAll(ctx context.Context) ([]*model.Article, error) {
	// redisから読み込む
	data, err := r.rdb.Get(ctx, "articles:list").Result()
	if err == nil {
		var articles []*model.Article
		json.Unmarshal([]byte(data), &articles)
		return articles, nil
	}

	// キャッシュミス
	query := `SELECT id, title, content, created_at, updated_at FROM articles`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close() // 接続が終わったら。プールに戻す

	var articles []*model.Article
	for rows.Next() {
		var article model.Article
		err := rows.Scan(&article.ID, &article.Title, &article.Content, &article.CreatedAt, &article.UpdatedAt)
		if err != nil {
			return nil, err
		}
		articles = append(articles, &article)
	}

	//　redisに保存
	jsonData, _ := json.Marshal(articles)
	r.rdb.Set(ctx, "articles:list", jsonData, REDIS_CACHE_TIME)
	return articles, nil
}
