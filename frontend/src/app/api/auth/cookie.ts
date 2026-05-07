export function isSecureCookieEnabled() {
  // Cookie の Secure 属性を付けるかどうかを環境変数で制御します。
  //
  // Secure=true:
  //   HTTPS の時だけ Cookie を保存/送信する。
  //   本番HTTPSではこれが安全。
  //
  // Secure=false:
  //   HTTP でも Cookie を保存/送信できる。
  //   ドメイン/SSL設定前に、VPSのIPアドレスで動作確認する時に使う。
  //
  // 注意:
  //   NODE_ENV=production だから常に secure=true にすると、
  //   http://43.xxx.xxx.xxx の初期検証でブラウザが Cookie を保存しない。
  const value = process.env.COOKIE_SECURE;
  if (value != null && value !== '') {
    return value.toLowerCase() === 'true';
  }

  return process.env.NODE_ENV === 'production';
}
