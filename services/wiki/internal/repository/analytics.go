package repository

import (
	"context"
	"time"

	"github.com/TenshoOHASHI/knowhub/pkg/dbutil"
	"github.com/google/uuid"
)

type AnalyticsRepository interface {
	RecordPageView(ctx context.Context, path, ipHash, userAgent, referrer string) error
	GetSummary(ctx context.Context, days int) (totalViews int32, uniqueVisitors int32, todayViews int32, dailyViews []DailyCount, pageRanking []PageRank, articleRanking []ArticleRank, likeRanking []LikeRank, err error)
}

type DailyCount struct {
	Date           string
	Count          int32
	UniqueVisitors int32
}

type PageRank struct {
	Path  string
	Count int32
}

type ArticleRank struct {
	ID         string
	Title      string
	Count      int32
	Visibility string
}

type LikeRank struct {
	ID    string
	Title string
	Count int32
}

type mysqlAnalyticsRepository struct {
	db dbutil.DB
}

func NewMysqlAnalyticsRepository(db dbutil.DB) AnalyticsRepository {
	return &mysqlAnalyticsRepository{db}
}

func (r *mysqlAnalyticsRepository) RecordPageView(ctx context.Context, path, ipHash, userAgent, referrer string) error {
	id := uuid.New().String()
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO page_views (id, path, ip_hash, user_agent, referrer, visited_at) VALUES (?, ?, ?, ?, ?, ?)",
		id, path, ipHash, userAgent, referrer, time.Now(),
	)
	return err
}

func (r *mysqlAnalyticsRepository) GetSummary(ctx context.Context, days int) (int32, int32, int32, []DailyCount, []PageRank, []ArticleRank, []LikeRank, error) {
	cutoff := time.Now().AddDate(0, 0, -days)

	// Total views in the period
	var totalViews int32
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM page_views WHERE visited_at >= ? AND path NOT IN ('/admin', '/saved', '/chat')",
		cutoff,
	).Scan(&totalViews)
	if err != nil {
		return 0, 0, 0, nil, nil, nil, nil, err
	}

	// Unique visitors (by IP hash)
	var uniqueVisitors int32
	err = r.db.QueryRowContext(ctx,
		"SELECT COUNT(DISTINCT ip_hash) FROM page_views WHERE visited_at >= ? AND path NOT IN ('/admin', '/saved', '/chat')",
		cutoff,
	).Scan(&uniqueVisitors)
	if err != nil {
		return 0, 0, 0, nil, nil, nil, nil, err
	}

	// Today's views
	var todayViews int32
	today := time.Now().Truncate(24 * time.Hour)
	err = r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM page_views WHERE visited_at >= ? AND path NOT IN ('/admin', '/saved', '/chat')",
		today,
	).Scan(&todayViews)
	if err != nil {
		return 0, 0, 0, nil, nil, nil, nil, err
	}

	// Daily breakdown with unique visitors
	rows, err := r.db.QueryContext(ctx,
		`SELECT DATE(visited_at) as d, COUNT(*) as cnt, COUNT(DISTINCT ip_hash) as unique_cnt
		FROM page_views
		WHERE visited_at >= ? AND path NOT IN ('/admin', '/saved', '/chat')
		GROUP BY DATE(visited_at)
		ORDER BY d`,
		cutoff,
	)
	if err != nil {
		return 0, 0, 0, nil, nil, nil, nil, err
	}
	defer rows.Close()

	var dailyViews []DailyCount
	for rows.Next() {
		var dc DailyCount
		if err := rows.Scan(&dc.Date, &dc.Count, &dc.UniqueVisitors); err != nil {
			return 0, 0, 0, nil, nil, nil, nil, err
		}
		dailyViews = append(dailyViews, dc)
	}

	// Page ranking (exclude admin pages, individual wiki articles)
	rankRows, err := r.db.QueryContext(ctx,
		`SELECT path, COUNT(*) as cnt
		FROM page_views
		WHERE visited_at >= ?
		AND path NOT IN ('/admin', '/', '/login', '/register', '/admin/register', '/admin/login')
		AND path NOT LIKE '/wiki/%'
		AND path NOT LIKE '/admin/%'
		GROUP BY path
		ORDER BY cnt DESC
		LIMIT 10`,
		cutoff,
	)
	if err != nil {
		return 0, 0, 0, nil, nil, nil, nil, err
	}
	defer rankRows.Close()

	var pageRanking []PageRank
	for rankRows.Next() {
		var pr PageRank
		if err := rankRows.Scan(&pr.Path, &pr.Count); err != nil {
			return 0, 0, 0, nil, nil, nil, nil, err
		}
		pageRanking = append(pageRanking, pr)
	}

	// Article ranking (join with articles table to get titles)
	articleRows, err := r.db.QueryContext(ctx,
		`SELECT COALESCE(a.id, ''), COALESCE(a.title, 'Unknown'), COUNT(pv.id) as cnt, COALESCE(a.visibility, 'unknown')
		FROM page_views pv
		LEFT JOIN articles a ON a.id = REPLACE(pv.path, '/wiki/', '')
		WHERE pv.visited_at >= ? AND pv.path LIKE '/wiki/%' AND LENGTH(pv.path) > 6 AND LENGTH(pv.path) < 50
		GROUP BY a.id, a.title, a.visibility
		HAVING a.id IS NOT NULL AND a.visibility = 'public' AND COUNT(pv.id) > 0
		ORDER BY cnt DESC
		LIMIT 10`,
		cutoff,
	)
	if err != nil {
		return 0, 0, 0, nil, nil, nil, nil, err
	}
	defer articleRows.Close()

	var articleRanking []ArticleRank
	for articleRows.Next() {
		var ar ArticleRank
		if err := articleRows.Scan(&ar.ID, &ar.Title, &ar.Count, &ar.Visibility); err != nil {
			return 0, 0, 0, nil, nil, nil, nil, err
		}
		articleRanking = append(articleRanking, ar)
	}

	// Like ranking (join with articles table)
	likeRows, err := r.db.QueryContext(ctx,
		`SELECT a.id, a.title, COUNT(l.id) as cnt
		FROM article_likes l
		INNER JOIN articles a ON a.id = l.article_id
		GROUP BY a.id, a.title
		ORDER BY cnt DESC
		LIMIT 10`,
	)
	if err != nil {
		return 0, 0, 0, nil, nil, nil, nil, err
	}
	defer likeRows.Close()

	var likeRanking []LikeRank
	for likeRows.Next() {
		var lr LikeRank
		if err := likeRows.Scan(&lr.ID, &lr.Title, &lr.Count); err != nil {
			return 0, 0, 0, nil, nil, nil, nil, err
		}
		likeRanking = append(likeRanking, lr)
	}

	return totalViews, uniqueVisitors, todayViews, dailyViews, pageRanking, articleRanking, likeRanking, nil
}
