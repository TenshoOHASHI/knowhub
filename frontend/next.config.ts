import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  // /api/* を Gateway にプロキシ（リバースプロキシ）
  // ブラウザは同一オリジン (localhost:3000) にリクエスト → Cookie が自動送信される
  // Next.js がサーバー側で Gateway (localhost:8080) に転送
  //
  // Route Handlers (/api/auth/*) は rewrites より優先されるため、
  // ログイン/登録の Cookie セット処理は Route Handlers が担当
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://localhost:8080/api/:path*',
      },
      {
        source: '/uploads/:path*',
        destination: 'http://localhost:8080/uploads/:path*',
      },
    ];
  },

  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: '**',
      },
    ],
  },
};

export default nextConfig;
