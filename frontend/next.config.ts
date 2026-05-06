import type { NextConfig } from 'next';

const gatewayInternalUrl = process.env.GATEWAY_INTERNAL_URL || 'http://localhost:8080';

const nextConfig: NextConfig = {
  output: 'standalone',
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
        destination: `${gatewayInternalUrl}/api/:path*`,
      },
      {
        source: '/uploads/:path*',
        destination: `${gatewayInternalUrl}/uploads/:path*`,
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
