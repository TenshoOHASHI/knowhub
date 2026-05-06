package repository

import (
	"context"
	"time"

	"github.com/TenshoOHASHI/knowhub/pkg/dbutil"
	"github.com/google/uuid"
)

type AnalyticsRepository interface {
	RecordPageView(ctx context.Context, path, ipHash, userAgent, referrer string) error
	GetSummary(ctx context.Context, days int) (totalViews int32, todayViews int32, dailyViews []DailyCount, pageRanking []PageRank, err error)
}

type DailyCount struct {
	Date  string
	Count int32
}

type PageRank struct {
	Path  string
	Count int32
}

type mysqlAnalyticsRepository struct {
	db dbutil.DB
}

func NewMysqlAnalyticsRepository(db dbutil.DB) AnalyticsRepository {
	return &mysqlAnalyticsRepository{db: db}
}

func (r *mysqlAnalyticsRepository) RecordPageView(ctx context.Context, path, ipHash, userAgent, referrer string) error {
	id := uuid.New().String()
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO page_views (id, path, ip_hash, user_agent, referrer, visited_at) VALUES (?, ?, ?, ?, ?, ?)",
		id, path, ipHash, userAgent, referrer, time.Now(),
	)
	return err
}

func (r *mysqlAnalyticsRepository) GetSummary(ctx context.Context, days int) (int32, int32, []DailyCount, []PageRank, error) {
	// Total views in the period
	var totalViews int32
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM page_views WHERE visited_at >= ?",
		time.Now().AddDate(0, 0, -days),
	).Scan(&totalViews)
	if err != nil {
		return 0, 0, nil, nil, err
	}

	// Today's views
	var todayViews int32
	today := time.Now().Truncate(24 * time.Hour)
	err = r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM page_views WHERE visited_at >= ?",
		today,
	).Scan(&todayViews)
	if err != nil {
		return 0, 0, nil, nil, err
	}

	// Daily breakdown
	rows, err := r.db.QueryContext(ctx,
		"SELECT DATE(visited_at) as d, COUNT(*) as cnt FROM page_views WHERE visited_at >= ? GROUP BY DATE(visited_at) ORDER BY d",
		time.Now().AddDate(0, 0, -days),
	)
	if err != nil {
		return 0, 0, nil, nil, err
	}
	defer rows.Close()

	var dailyViews []DailyCount
	for rows.Next() {
		var dc DailyCount
		if err := rows.Scan(&dc.Date, &dc.Count); err != nil {
			return 0, 0, nil, nil, err
		}
		dailyViews = append(dailyViews, dc)
	}

	// Page ranking (top 10)
	rankRows, err := r.db.QueryContext(ctx,
		"SELECT path, COUNT(*) as cnt FROM page_views WHERE visited_at >= ? GROUP BY path ORDER BY cnt DESC LIMIT 10",
		time.Now().AddDate(0, 0, -days),
	)
	if err != nil {
		return 0, 0, nil, nil, err
	}
	defer rankRows.Close()

	var pageRanking []PageRank
	for rankRows.Next() {
		var pr PageRank
		if err := rankRows.Scan(&pr.Path, &pr.Count); err != nil {
			return 0, 0, nil, nil, err
		}
		pageRanking = append(pageRanking, pr)
	}

	return totalViews, todayViews, dailyViews, pageRanking, nil
}
