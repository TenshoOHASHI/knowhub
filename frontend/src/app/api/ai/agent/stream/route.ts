import { NextRequest } from 'next/server';

/**
 * SSE ストリーミングプロキシ Route Handler
 *
 * なぜこれが必要か:
 *   Next.js の rewrites() は SSE レスポンスをバッファリングしてしまう。
 *   そのためブラウザにイベントが一気に届き、リアルタイム表示ができない。
 *
 *   この Route Handler が Gateway から SSE を受信し、
 *   1イベントずつブラウザに転送する (ReadabbleStream → TransformStream)。
 *
 * データフロー:
 *   Browser → POST /api/ai/agent/stream → このRoute Handler
 *     → fetch(Gateway) → Gateway の SSE レスポンス
 *     → ReadableStream で1チャンクずつ読み取り
 *     → ブラウザにそのまま転送 (Streaming SSR)
 */
export async function POST(req: NextRequest) {
  const body = await req.json();

  // チャット機能は認証なしで利用可能
  // API Key, model, question などの情報は body に含まれている
  const gatewayRes = await fetch('http://localhost:8080/api/ai/agent/stream', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });

  if (!gatewayRes.ok) {
    return new Response(
      JSON.stringify({ error: `Gateway error: ${gatewayRes.status}` }),
      {
        status: gatewayRes.status,
        headers: { 'Content-Type': 'application/json' },
      },
    );
  }

  // Gateway のレスポンスボディ (SSE ReadableStream) を
  // そのままブラウザにストリーミング転送
  const stream = gatewayRes.body!;
  return new Response(stream, {
    headers: {
      'Content-Type': 'text/event-stream',
      'Cache-Control': 'no-cache',
      Connection: 'keep-alive',
    },
  });
}
