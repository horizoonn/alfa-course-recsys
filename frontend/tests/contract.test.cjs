const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const ts = require('typescript');
const React = require('react');
const { renderToStaticMarkup } = require('react-dom/server');

globalThis.crypto ??= require('node:crypto').webcrypto;

const loadTypeScript = (module, filename) => {
  const source = fs.readFileSync(filename, 'utf8').replaceAll(
    'import.meta.env',
    "({ VITE_USE_MOCK: 'false', VITE_API_BASE: '/api/' })",
  );
  module._compile(ts.transpileModule(source, {
    compilerOptions: {
      module: ts.ModuleKind.CommonJS,
      target: ts.ScriptTarget.ES2020,
      jsx: ts.JsxEmit.ReactJSX,
    },
  }).outputText, filename);
};
require.extensions['.ts'] = loadTypeScript;
require.extensions['.tsx'] = loadTypeScript;

const Card = require('../src/components/MaterialCard.tsx').default;
const Results = require('../src/screens/Results.tsx').default;
const api = require('../src/api/client.ts');
const { mockDemoProfiles, mockRecommendations } = require('../src/api/mock.ts');

const card = {
  material_key: 'term:00000000-0000-4000-8000-000000000001',
  title: 'Синтетический термин',
  type: 'term',
  safe_summary: null,
  access: 'subscription',
  source_url: 'https://example.com/material',
  reasons: [{ code: 'task_match', text: 'Соответствует задаче' }],
  next_action: { type: 'open_material', label: 'Открыть материал' },
};
const response = {
  request_id: '00000000-0000-4000-8000-000000000002',
  mode: 'topic',
  items: [card],
  limitations: [],
};
const renderCard = (item) => renderToStaticMarkup(React.createElement(Card, {
  item,
  onRefineRequest() {},
}));
const renderResults = (result) => renderToStaticMarkup(React.createElement(Results, {
  result,
  loading: false,
  error: null,
  onRetry() {},
  onChangeRequest() {},
}));

function stubFetch(t, fn) {
  const previous = globalThis.fetch;
  globalThis.fetch = fn;
  t.after(() => { globalThis.fetch = previous; });
}

function checkResponse(result) {
  assert.match(result.request_id, /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/);
  assert.ok(Array.isArray(result.limitations));
  assert.ok(result.items.length <= 3);
  assert.equal(new Set(result.items.map((item) => item.material_key)).size, result.items.length);
  for (const item of result.items) {
    assert.ok(item.material_key.startsWith(item.type + ':'));
    assert.ok(item.reasons.length > 0);
    assert.ok(item.reasons.every((reason) => typeof reason.code === 'string' && reason.text));
    if (item.next_action.type !== 'refine_request') assert.ok(item.source_url);
    assert.doesNotThrow(() => renderCard(item));
  }
  assert.doesNotThrow(() => renderResults(result));
}

test('cards render all access models and structured reasons and actions', () => {
  for (const [access, label] of Object.entries({
    free: 'Бесплатно', subscription: 'По подписке', paid: 'Платно',
    unknown: 'Условия на странице материала',
  })) {
    const html = renderCard({ ...card, access });
    assert.ok(html.includes(label));
    assert.ok(html.includes('Термин'));
    assert.ok(html.includes('Соответствует задаче'));
    assert.ok(html.includes('Открыть материал'));
    assert.ok(!html.includes('card-summary'));
    assert.ok(!html.includes('[object Object]'));
  }
  assert.ok(renderCard({ ...card, access: 'future-model' }).includes('Условия на странице материала'));
});

test('refine action is a button even when a URL is present', () => {
  const html = renderCard({ ...card, next_action: { type: 'refine_request', label: 'Уточнить задачу' } });
  assert.match(html, /<button/);
  assert.ok(html.includes('Уточнить задачу'));
  assert.ok(!html.includes('href='));
  assert.match(renderCard({ ...card, source_url: null }), /Изменить подборку/);
});

test('result limitations and empty state render messages rather than objects', () => {
  const limitations = [{ code: 'NO_MATCHES', message: 'Совпадений нет' }];
  assert.ok(renderResults({ ...response, items: [], limitations }).includes('Совпадений нет'));
  assert.ok(renderResults({ ...response, limitations }).includes('Совпадений нет'));
});

test('client sends profile or demo ID and normalizes the API base', async (t) => {
  const requests = [];
  stubFetch(t, async (url, init) => {
    requests.push({ url, body: JSON.parse(init.body) });
    return new Response(JSON.stringify(response), { status: 200 });
  });
  await api.fetchRecommendations({ task_id: 'finance' });
  await api.fetchRecommendations({ task_id: 'finance' }, 'P02');
  assert.deepEqual(requests, [
    { url: '/api/recommendations', body: { profile: { task_id: 'finance' } } },
    { url: '/api/recommendations', body: { demo_profile_id: 'P02' } },
  ]);
});

test('client preserves nested API error code and message', async (t) => {
  stubFetch(t, async () => new Response(JSON.stringify({
    request_id: response.request_id,
    error: { code: 'INVALID_INPUT', message: 'Неизвестная задача', fields: [] },
  }), { status: 400 }));
  await assert.rejects(api.fetchRecommendations({}), (error) =>
    error.code === 'INVALID_INPUT' && error.message === 'Неизвестная задача' && error.status === 400);
});

test('non-JSON 503 and network failure remain errors', async (t) => {
  stubFetch(t, async () => new Response('Unavailable', { status: 503 }));
  await assert.rejects(api.fetchRecommendations({}), (error) => error.code === 'DATA_UNAVAILABLE');
  globalThis.fetch = async () => { throw new TypeError('Network failed'); };
  await assert.rejects(api.fetchRecommendations({}), (error) => error.code === 'NETWORK_ERROR');
});

test('all demo responses satisfy the frontend contract', async () => {
  for (const demo of mockDemoProfiles) checkResponse(await mockRecommendations({}, demo.id));
});

test('help-mode-only profile is recognized and produces matching mock reasons', async () => {
  const result = await mockRecommendations({ help_mode: 'tool' });
  assert.equal(result.mode, 'fallback');
  assert.ok(result.items.length > 0);
  assert.ok(result.items.every((item) => item.reasons.some((reason) => reason.code === 'help_mode_match')));
  assert.ok(!result.limitations.some((limitation) => limitation.message.includes('Профиль пустой')));
});

test('P10 explicitly reports that history is unavailable', async () => {
  const result = await mockRecommendations({}, 'P10');
  assert.ok(result.limitations.some((limitation) => limitation.code === 'HISTORY_UNAVAILABLE'));
  assert.ok(result.items.every((item) => item.reasons.every((reason) => reason.code !== 'history_match')));
});

test('missing task matches are based on the mock catalog rather than one task name', async () => {
  const result = await mockRecommendations({ task_id: 'team' });
  assert.deepEqual(result.items, []);
  assert.ok(result.limitations.some((limitation) => limitation.code === 'NO_MATCHES'));
});


test('mock request IDs work without crypto.randomUUID on HTTP origins', async (t) => {
  const previous = Object.getOwnPropertyDescriptor(globalThis, 'crypto');
  const webcrypto = require('node:crypto').webcrypto;
  Object.defineProperty(globalThis, 'crypto', {
    configurable: true,
    value: { getRandomValues: webcrypto.getRandomValues.bind(webcrypto) },
  });
  t.after(() => { Object.defineProperty(globalThis, 'crypto', previous); });
  const first = await mockRecommendations({});
  const second = await mockRecommendations({});
  assert.match(first.request_id, /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/);
  assert.notEqual(first.request_id, second.request_id);
});
