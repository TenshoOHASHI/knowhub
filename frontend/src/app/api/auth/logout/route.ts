import { NextResponse } from 'next/server';

// ログアウト Route Handler
// HttpOnly Cookie は JavaScript から削除できないため、
// サーバー側で MaxAge=0 にして Cookie を削除する
export async function POST() {
  const response = NextResponse.json({ ok: true });
  response.cookies.set('token', '', {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    maxAge: 0, // Cookie を即座に削除
    path: '/',
  });
  return response;
}
