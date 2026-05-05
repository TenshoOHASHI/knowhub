'use client';

import { useEffect } from 'react';
import { usePathname } from 'next/navigation';
import { sendPageView } from '@/lib/analytics';

export default function AnalyticsTracker() {
  const pathname = usePathname();

  useEffect(() => {
    // Send page view beacon on route change
    sendPageView(pathname);
  }, [pathname]);

  return null;
}
