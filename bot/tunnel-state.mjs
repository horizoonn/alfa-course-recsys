function targetFromText(text) {
  let value = text.trim();
  if (value.startsWith('{')) {
    const document = JSON.parse(value);
    if (document.encoding !== 'base64' || typeof document.content !== 'string') return '';
    value = Buffer.from(document.content, 'base64').toString('utf8').trim();
  }
  const target = new URL(value);
  if (target.protocol !== 'https:' || target.username || target.password || target.search || target.hash) return '';
  return target.toString().replace(/\/+$/, '');
}

export async function readTunnelTarget(stateUrl, { fallbackUrl, currentTarget = '', fetcher = fetch } = {}) {
  for (const source of [stateUrl, fallbackUrl].filter(Boolean)) {
    try {
      const url = new URL(source);
      url.searchParams.set('_', String(Date.now()));
      const response = await fetcher(url.toString(), {
        headers: { accept: 'application/vnd.github.raw', 'user-agent': 'alfa-course-bot' },
        signal: AbortSignal.timeout(10_000), cache: 'no-store',
      });
      if (!response.ok) continue;
      const text = await response.text();
      if (text.length > 4096) continue;
      const target = targetFromText(text);
      if (!target) continue;
      if (target === currentTarget) return target;
      // Never replace a live target with a stale, disconnected tunnel.
      const health = await fetcher(`${target}/api/health`, { signal: AbortSignal.timeout(10_000) });
      if (health.ok && (await health.json()).status === 'ok') return target;
    } catch { /* Try the fallback; keep the last target if neither works. */ }
  }
  return '';
}
