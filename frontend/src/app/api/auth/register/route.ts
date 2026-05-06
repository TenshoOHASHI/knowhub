import { NextRequest, NextResponse } from 'next/server';

// ユーザー登録 Route Handler
// login/route.ts と同じ理由で Route Handler を経由する
export async function POST(req: NextRequest) {
  const body = await req.json();
  const gatewayURL =
    process.env.GATEWAY_INTERNAL_URL || 'http://localhost:8080';

  // local:
  //   http://localhost:8080
  // production docker:
  //   http://gateway:8080
  const res = await fetch(`${gatewayURL}/api/user/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });

  if (!res.ok) {
    return NextResponse.json({ error: 'register failed' }, { status: 400 });
  }

  const data = await res.json();

  const response = NextResponse.json({ user: data.user });
  response.cookies.set('token', data.token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    maxAge: 86400,
    path: '/',
  });

  return response;
}
