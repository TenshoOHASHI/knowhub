// Analytics beacon — sends page views via navigator.sendBeacon

export function sendPageView(path: string) {
  if (typeof window === 'undefined') return;
  if (!navigator.sendBeacon) return;

  const payload = JSON.stringify({ path });
  navigator.sendBeacon('/api/analytics/ping', payload);
}
