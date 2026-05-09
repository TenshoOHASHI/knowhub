import type { Metadata, Viewport } from 'next';
import './globals.css';
import Navbar from '@/components/Navbar';
import KeyboardShortcuts from '@/components/KeyBoardShortcuts';
import AnalyticsTracker from '@/components/AnalyticsTracker';
import { SidebarProvider } from '@/context/SidebarContext';
import Footer from '@/components/Footer';
import { ToastProvider } from '@/context/ToastContext';
import { AuthProvider } from '@/context/AuthContext';
import { ThemeProvider } from '@/context/ThemeContext';

export const metadata: Metadata = {
  title: 'TenHub - 学びの断片を、知識の資産へ変えるナレッジベース',
  description:
    '技術を記し、考え、発信する。小さなアウトプットが、やがてプロダクトになる。Go マイクロサービス + Next.js で構築された、モダンな技術スタックの実践場。',
  keywords: [
    'ナレッジベース',
    'Wiki',
    'AI',
    'RAG',
    '技術ブログ',
    'Go',
    'Next.js',
  ],
  authors: [{ name: 'Tensho' }],
  openGraph: {
    title: '学びの断片を、知識の資産へ変える',
    description:
      'Go/TypeScript/Next.jsで開発するAIナレッジベース\n技術を記し、考え、発信する。',
    type: 'website',
    url: 'https://www.tenhub.tech',
    siteName: 'TenHub',
    images: ['https://www.tenhub.tech/api/og'],
    locale: 'ja_JP',
  },
  twitter: {
    card: 'summary_large_image',
    title: '学びの断片を、知識の資産へ変える',
    description:
      'Go/TypeScript/Next.jsで開発するAIナレッジベース\n技術を記し、考え、発信する。',
    images: ['https://www.tenhub.tech/api/og'],
  },
};

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  viewportFit: 'cover',
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang='ja' className='h-full antialiased' suppressHydrationWarning>
      <body className='app-shell flex flex-col overflow-hidden'>
        <ThemeProvider>
          <div className='flex flex-col h-full bg-white dark:bg-stone-900 text-black dark:text-stone-100'>
            <ToastProvider>
              <AuthProvider>
                <SidebarProvider>
                  <Navbar />
                  <AnalyticsTracker />
                  <main className='min-h-0 flex-1 overflow-y-auto overflow-x-hidden'>
                    {children}
                  </main>
                </SidebarProvider>
                <Footer />
                <KeyboardShortcuts />
              </AuthProvider>
            </ToastProvider>
          </div>
        </ThemeProvider>
      </body>
    </html>
  );
}
