import { NextRequest, NextResponse } from 'next/server';

const API_BASE = process.env.SERVER_API_URL || 'http://localhost:8080/api';

export async function POST(req: NextRequest) {
  try {
    const body = await req.json();
    const forwardedFor = req.headers.get('x-forwarded-for');
    const realIp = req.headers.get('x-real-ip');
    const userAgent = req.headers.get('user-agent');
    const referrer = req.headers.get('referer');

    const headers = new Headers({ 'Content-Type': 'application/json' });
    if (forwardedFor) headers.set('X-Forwarded-For', forwardedFor);
    if (realIp) headers.set('X-Real-IP', realIp);
    if (userAgent) headers.set('User-Agent', userAgent);
    if (referrer) headers.set('Referer', referrer);

    const gatewayRes = await fetch(`${API_BASE}/analytics/ping`, {
      method: 'POST',
      headers,
      body: JSON.stringify(body),
    });
    return new NextResponse(null, { status: gatewayRes.status });
  } catch {
    return new NextResponse(null, { status: 204 });
  }
}
