import { NextRequest, NextResponse } from 'next/server';

// ログイン Route Handler
// Browser → Next.js Route Handler → Gateway
//
// なぜ直接 Gateway に fetch しないのか:
// Gateway は Set-Cookie で token を返すが、クロスオリジンの場合、
// ブラウザの SameSite=Lax 制限により Cookie がブラウザに保存されない。
// そのため Route Handler が中継し、Next.js 側からブラウザに Cookie をセットする。
export async function POST(req: NextRequest) {
  const body = await req.json();
  const gatewayURL =
    process.env.GATEWAY_INTERNAL_URL || 'http://localhost:8080';

  // サーバー間通信: Gateway にログインリクエスト（SameSite 制限の対象外）
  // local:
  //   http://localhost:8080
  // production docker:
  //   http://gateway:8080
  const res = await fetch(`${gatewayURL}/api/user/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });

  if (!res.ok) {
    return NextResponse.json({ error: 'login failed' }, { status: 401 });
  }

  // Gateway のレスポンスから user と token を取り出す
  const data = await res.json();

  // Next.js サーバーからブラウザに HttpOnly Cookie をセット
  // これでブラウザ側の同一origin Cookieとして保存される
  const response = NextResponse.json({ user: data.user });
  response.cookies.set('token', data.token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    maxAge: 86400, // 24時間
    path: '/',
  });

  return response;
}
