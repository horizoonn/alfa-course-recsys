import test from 'node:test';
import assert from 'node:assert/strict';
import { readTunnelTarget } from './tunnel-state.mjs';

test('accepts a raw state URL only after checking the real API', async () => {
  const calls = [];
  const target = await readTunnelTarget('https://example.com/state?ref=tunnel', {
    fetcher: async (url) => {
      calls.push(url);
      return calls.length === 1 ? new Response('https://working.lhr.life/\n') : Response.json({ status: 'ok' });
    },
  });
  assert.equal(target, 'https://working.lhr.life');
  assert.equal(new URL(calls[0]).searchParams.get('ref'), 'tunnel');
  assert.ok(new URL(calls[0]).searchParams.has('_'));
  assert.equal(calls[1], 'https://working.lhr.life/api/health');
});

test('uses a fresh fallback when the primary state points at a dead tunnel', async () => {
  const target = await readTunnelTarget('https://example.com/primary', {
    fallbackUrl: 'https://example.com/fallback',
    fetcher: async (url) => {
      if (url.includes('/primary')) return new Response('https://dead.lhr.life');
      if (url.includes('/fallback')) return new Response('https://live.lhr.life');
      if (url.includes('dead.lhr.life')) return new Response('Unavailable', { status: 502 });
      return Response.json({ status: 'ok' });
    },
  });
  assert.equal(target, 'https://live.lhr.life');
});

test('supports GitHub JSON and does not assume that HTTP 200 means healthy API', async () => {
  const state = Response.json({ encoding: 'base64', content: Buffer.from('https://live.lhr.life').toString('base64') });
  let count = 0;
  const target = await readTunnelTarget('https://example.com/state', {
    fetcher: async () => ++count === 1 ? state : Response.json({ status: 'ok' }),
  });
  assert.equal(target, 'https://live.lhr.life');
  count = 0;
  assert.equal(await readTunnelTarget('https://example.com/state', {
    fetcher: async () => ++count === 1 ? new Response('https://live.lhr.life') : new Response('<html>Error</html>'),
  }), '');
});

test('invalid or unavailable state keeps the existing target', async () => {
  for (const value of ['http://insecure.example', 'https://user:password@example.com', 'not a URL']) {
    assert.equal(await readTunnelTarget('https://example.com/state', { fetcher: async () => new Response(value) }), '');
  }
});
