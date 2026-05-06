import { NextRequest } from 'next/server';

/**
 * SSE ストリーミングプロキシ Route Handler
 *
 * なぜこれが必要か:
 *   Next.js の rewrites() は SSE レスポンスをバッファリングしてしまう。
 *   そのためブラウザにイベントが一気に届き、リアルタイム表示ができない。
 *
 *   この Route Handler が Gateway から SSE を受信し、
 *   1イベントずつブラウザに転送する (ReadableStream → TransformStream)。
 *
 * データフロー:
 *   Browser → POST /api/ai/agent/stream → このRoute Handler
 *     → fetch(Gateway) → Gateway の SSE レスポンス
 *     → ReadableStream で1チャンクずつ読み取り
 *     → ブラウザにそのまま転送 (Streaming SSR)
 */
export async function POST(req: NextRequest) {
  // この Route Handler は「JSONを処理する場所」ではなく、
  // 「Browser から来た body を Gateway へそのまま運ぶ proxy」。
  //
  // Browser が送る body の実体:
  // {
  //   "question": "...",
  //   "model": "deepseek-chat",
  //   "api_key": "...",
  //   "search_engine": "hybrid",
  //   "enable_web_search": true,
  //   "history": "[{\"role\":\"user\",\"content\":\"...\"}]"
  // }
  //
  // req.json() にすると JSON文字列 -> object -> JSON文字列 に戻すことになる。
  // ここでは中身を加工しないため、text() で読んでそのまま Gateway に渡す。
  const body = await req.text();
  const gatewayURL =
    process.env.GATEWAY_INTERNAL_URL || 'http://localhost:8080';

  // Gateway に渡すヘッダー。
  // Content-Type は JSON body を正しく decode してもらうために必要。
  //
  // Cookie / Authorization はログイン状態を Gateway に伝えるために必要。
  // これがないと Browser -> Next.js まではログイン済みでも、
  // Next.js -> Gateway では未ログイン扱いになり、AI rate limit 対象になる。
  const headers: Record<string, string> = {
    'Content-Type': req.headers.get('Content-Type') || 'application/json',
  };
  const cookie = req.headers.get('Cookie');
  const authorization = req.headers.get('Authorization');
  if (cookie) headers.Cookie = cookie;
  if (authorization) headers.Authorization = authorization;

  // チャット機能は認証なしで利用可能
  // API Key, model, question などの情報は body に含まれている
  const gatewayRes = await fetch(`${gatewayURL}/api/ai/agent/stream`, {
    method: 'POST',
    headers,
    body,
  });

  if (!gatewayRes.ok) {
    // Gateway が 429 などを返した場合、本文と rate limit ヘッダーを捨てずに
    // Browser 側へ伝播する。
    //
    // 例:
    // status: 429
    // body: "anonymous AI daily limit exceeded"
    // Retry-After: "3600"
    //
    // frontend/src/lib/api.ts はこの body/header を読んで、
    // 「混雑中」または「本日の上限到達」の表示文に変換する。
    const errorText = await gatewayRes.text();
    const responseHeaders = new Headers({
      'Content-Type': 'text/plain; charset=utf-8',
    });
    copyHeader(gatewayRes, responseHeaders, 'Retry-After');
    copyHeader(gatewayRes, responseHeaders, 'X-RateLimit-Limit');
    copyHeader(gatewayRes, responseHeaders, 'X-RateLimit-Remaining');

    return new Response(errorText || `Gateway error: ${gatewayRes.status}`, {
      status: gatewayRes.status,
      headers: responseHeaders,
    });
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

function copyHeader(from: Response, to: Headers, name: string) {
  const value = from.headers.get(name);
  if (value) to.set(name, value);
}
