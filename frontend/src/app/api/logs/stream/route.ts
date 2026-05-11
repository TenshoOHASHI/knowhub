import { NextRequest } from 'next/server';

/**
 * SSE ストリーミングプロキシ Route Handler (ログ監視)
 *
 * Next.js の rewrites() は SSE レスポンスをバッファリングするため、
 * Route Handler で Gateway の SSE を 1 イベントずつブラウザに転送する。
 *
 * データフロー:
 *   Browser → GET /api/logs/stream?services=ai,gateway&level=error
 *     → この Route Handler → fetch(Gateway /api/logs/stream)
 *     → ReadableStream → ブラウザにストリーミング転送
 */
export async function GET(req: NextRequest) {
  const gatewayURL =
    process.env.GATEWAY_INTERNAL_URL || 'http://localhost:8080';

  // クエリパラメータをそのまま転送
  const searchParams = req.nextUrl.searchParams.toString();
  const url = `${gatewayURL}/api/logs/stream${searchParams ? `?${searchParams}` : ''}`;

  // Cookie / Authorization を Gateway に転送（認証必須エンドポイント）
  const headers: Record<string, string> = {};
  const cookie = req.headers.get('Cookie');
  const authorization = req.headers.get('Authorization');
  if (cookie) headers.Cookie = cookie;
  if (authorization) headers.Authorization = authorization;

  const gatewayRes = await fetch(url, {
    headers,
    signal: req.signal,
  });

  if (!gatewayRes.ok) {
    const errorText = await gatewayRes.text();
    return new Response(errorText || `Gateway error: ${gatewayRes.status}`, {
      status: gatewayRes.status,
      headers: { 'Content-Type': 'text/plain; charset=utf-8' },
    });
  }

  const stream = gatewayRes.body!;
  return new Response(stream, {
    headers: {
      'Content-Type': 'text/event-stream',
      'Cache-Control': 'no-cache',
      Connection: 'keep-alive',
    },
  });
}
