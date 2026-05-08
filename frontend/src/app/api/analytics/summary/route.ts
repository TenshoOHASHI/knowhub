import { NextRequest, NextResponse } from 'next/server';

const API_BASE = process.env.SERVER_API_URL || 'http://localhost:8080/api';

export async function GET(req: NextRequest) {
  try {
    const url = new URL(req.url);
    const days = url.searchParams.get('days') || '30';

    const gatewayRes = await fetch(
      `${API_BASE}/analytics/summary?days=${days}`,
      {
        method: 'GET',
      },
    );

    if (!gatewayRes.ok) {
      return new NextResponse('internal server error', {
        status: gatewayRes.status,
      });
    }

    const data = await gatewayRes.json();

    // Convert snake_case to camelCase for frontend
    const transformed = {
      totalViews: (data.total_views as number) || 0,
      uniqueVisitors: (data.unique_visitors as number) || 0,
      todayViews: (data.today_views as number) || 0,
      dailyViews: (
        (data.daily_views as Array<{
          date: string;
          count: number;
          unique_visitors: number;
        }>) || []
      ).map((d) => ({
        date: d.date,
        count: d.count,
        uniqueVisitors: d.unique_visitors,
      })),
      pageRanking: (
        (data.page_ranking as Array<{ path: string; count: number }>) || []
      ).map((p) => ({
        path: p.path,
        count: p.count,
      })),
      articleRanking: (
        (data.article_ranking as Array<{
          id: string;
          title: string;
          count: number;
        }>) || []
      ).map((a) => ({
        id: a.id,
        title: a.title,
        count: a.count,
      })),
      likeRanking: (
        (data.like_ranking as Array<{
          id: string;
          title: string;
          count: number;
        }>) || []
      ).map((l) => ({
        id: l.id,
        title: l.title,
        count: l.count,
      })),
    };

    return NextResponse.json(transformed);
  } catch {
    return new NextResponse('internal server error', { status: 500 });
  }
}
