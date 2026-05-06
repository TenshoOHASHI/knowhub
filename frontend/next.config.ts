import type { NextConfig } from 'next';

// Next.js server から Gateway へ接続する内部URLです。
//
// local:
//   GATEWAY_INTERNAL_URL 未設定 -> http://localhost:8080
//
// production docker:
//   docker-compose.prod.yml で GATEWAY_INTERNAL_URL=http://gateway:8080 を渡す
//   gateway は Docker Compose の service名で、Docker内部DNSで解決されます。
const gatewayInternalUrl = process.env.GATEWAY_INTERNAL_URL || 'http://localhost:8080';

const nextConfig: NextConfig = {
  output: 'standalone',
  // /api/* を Gateway にプロキシ（リバースプロキシ）
  // Browser は同一originの /api/* に送るため Cookie が自動送信される
  // Next.js server が Gateway へ server-to-server で転送する
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
