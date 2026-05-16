/* eslint-disable @next/next/no-img-element */
import { ImageResponse } from 'next/og';

export const runtime = 'edge';

export const size = { width: 1200, height: 630 };

export async function GET() {
  const logoUrl = `${process.env.NEXT_PUBLIC_BASE_URL || 'http://localhost:3000'}/dark.png`;

  const keywords = [
    'Go',
    'gRPC',
    'CQRS',
    'Next.js',
    'TypeScript',
    'Docker',
    'MySQL',
    'Redis',
    'Protocol Buffers',
    'Microservices',
    'React',
    'JavaScript',
    'Tailwind',
    'Git',
  ];

  const keywordPositions = [
    { top: '12%', left: '18%' },
    { top: '20%', left: '78%' },
    { top: '28%', left: '35%' },
    { top: '32%', left: '82%' },
    { top: '38%', left: '22%' },
    { top: '45%', left: '62%' },
    { top: '58%', left: '15%' },
    { top: '64%', left: '71%' },
    { top: '72%', left: '28%' },
    { top: '78%', left: '55%' },
    { top: '85%', left: '38%' },
    { top: '50%', left: '85%' },
    { top: '42%', left: '8%' },
    { top: '68%', left: '88%' },
  ];

  return new ImageResponse(
    <div
      style={{
        background: '#09090b',
        width: '100%',
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        fontFamily:
          '"Noto Sans JP", "Hiragino Kaku Gothic ProN", "Meiryo", sans-serif',
        padding: '80px',
        position: 'relative',
      }}
    >
      {/* Background grid */}
      <div
        style={{
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          backgroundImage:
            'linear-gradient(rgba(39, 39, 42, 0.3) 1px, transparent 1px), linear-gradient(90deg, rgba(39, 39, 42, 0.3) 1px, transparent 1px)',
          backgroundSize: '40px 40px',
        }}
      />

      {/* Floating keywords */}
      {keywords.map((kw, i) => (
        <div
          key={kw}
          style={{
            position: 'absolute',
            top: keywordPositions[i]?.top || '50%',
            left: keywordPositions[i]?.left || '50%',
            fontFamily: 'monospace',
            fontSize: `${0.65 + (i % 3) * 0.1}rem`,
            color: '#292524',
            fontWeight: 600,
          }}
        >
          {kw}
        </div>
      ))}

      {/* Left accent bar */}
      <div
        style={{
          position: 'absolute',
          left: 0,
          top: 0,
          bottom: 0,
          width: '4px',
          background:
            'linear-gradient(180deg, #57534e 0%, #a8a29e 50%, #57534e 100%)',
        }}
      />

      {/* Content wrapper */}
      <div
        style={{
          position: 'relative',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
        }}
      >
        {/* Logo + Title */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '20px',
            marginBottom: '50px',
          }}
        >
          <img
            src={logoUrl}
            alt='Logo'
            width={115}
            height={100}
            style={{
              width: 115,
              height: 100,
              filter: 'brightness(0.7) invert(1)',
            }}
          />
          <div
            style={{
              display: 'flex',
              fontSize: '84px',
              fontWeight: 700,
              letterSpacing: '-3px',
            }}
          >
            <span style={{ color: '#fafaf9' }}>Ten</span>
            <span style={{ color: '#78716c' }}>Hub</span>
          </div>
        </div>

        {/* Divider line */}
        <div
          style={{
            width: '80px',
            height: '2px',
            background:
              'linear-gradient(90deg, transparent, #57534e, transparent)',
            marginBottom: '40px',
          }}
        />

        {/* Tagline line 1 */}
        <div
          style={{
            fontSize: '26px',
            color: '#a8a29e',
            marginBottom: '8px',
          }}
        >
          技術を記し、考え、発信する。
        </div>

        {/* Tagline line 2 */}
        <div
          style={{
            fontSize: '26px',
            color: '#a8a29e',
            marginBottom: '40px',
          }}
        >
          小さなアウトプットが、やがてプロダクトになる。
        </div>

        {/* Build cycle */}
        <div
          style={{
            fontSize: '15px',
            fontFamily: 'monospace',
            color: '#71717a',
            letterSpacing: '1.5px',
          }}
        >
          build → learn → share → repeat_
        </div>
      </div>
    </div>,
    {
      width: 1200,
      height: 630,
      headers: {
        'Cache-Control': 'public, max-age=86400',
      },
    },
  );
}
