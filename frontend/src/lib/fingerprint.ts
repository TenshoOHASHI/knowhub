// Generate a stable anonymous fingerprint for the current browser
// Uses a random salt stored in localStorage + user agent
// No raw IP or PII is stored

const STORAGE_KEY = 'tenhub_fp_salt';

async function sha256(message: string): Promise<string> {
  const encoder = new TextEncoder();
  const data = encoder.encode(message);
  const hashBuffer = await crypto.subtle.digest('SHA-256', data);
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  return hashArray.map((b) => b.toString(16).padStart(2, '0')).join('');
}

function getOrCreateSalt(): string {
  if (typeof window === 'undefined') return '';
  let salt = localStorage.getItem(STORAGE_KEY);
  if (!salt) {
    salt = crypto.randomUUID();
    localStorage.setItem(STORAGE_KEY, salt);
  }
  return salt;
}

export async function getFingerprint(): Promise<string> {
  const salt = getOrCreateSalt();
  const ua = navigator.userAgent;
  return sha256(salt + ua);
}
