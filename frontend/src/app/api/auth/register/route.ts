import { NextRequest, NextResponse } from 'next/server';

// ユーザー登録 Route Handler
// login/route.ts と同じ理由で Route Handler を経由する
export async function POST(req: NextRequest) {
  const body = await req.json();

  const res = await fetch('http://localhost:8080/api/user/register', {
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
