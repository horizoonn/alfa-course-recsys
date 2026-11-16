// Смоук-тест веб-сервера без Telegram и без Docker:
//   npm run smoke
//
// Проверяет, что сервис отдаёт SPA, health-check, админку и корректно
// отвечает 502, когда туннель к API недоступен (фронт тогда уходит в демо-режим).

import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { createAppServer } from './server.js';

const HERE = path.dirname(fileURLToPath(import.meta.url));
const TOKEN = 'smoke-token';
const UNREACHABLE = 'http://127.0.0.1:9';

const silent = { log() {}, warn() {}, error() {} };

let failures = 0;

function check(name, condition, details = '') {
  if (condition) {
    console.log(`  ok   ${name}`);
    return;
  }
  failures += 1;
  console.log(`  FAIL ${name}${details ? ` — ${details}` : ''}`);
}

async function readJson(response) {
  const text = await response.text();
  try {
    return { json: JSON.parse(text), text };
  } catch {
    return { json: null, text };
  }
}

const { server } = createAppServer({
  staticDir: path.join(HERE, 'web'),
  apiTarget: UNREACHABLE,
  adminToken: TOKEN,
  revision: 'smoke-test',
  log: silent,
});

await new Promise((resolve) => server.listen(0, '127.0.0.1', resolve));
const base = `http://127.0.0.1:${server.address().port}`;
console.log(`[smoke] сервер поднят: ${base}`);

try {
  const index = await fetch(`${base}/`);
  const html = await index.text();
  check(
    'GET / отдаёт HTML',
    index.status === 200 && index.headers.get('content-type')?.includes('text/html'),
  );
  check('HTML содержит точку монтирования SPA', html.includes('id="root"'));

  const spa = await fetch(`${base}/results/deep-link`);
  const spaHtml = await spa.text();
  check('SPA-фолбэк отдаёт index.html', spa.status === 200 && spaHtml.includes('id="root"'));

  const health = await fetch(`${base}/health`);
  const healthJson = (await readJson(health)).json;
  check('GET /health = 200', health.status === 200);
  check('health сообщает о готовой статике', healthJson?.static_ready === true);
  check('health сообщает версию размещённой сборки', healthJson?.revision === 'smoke-test');
  check('health отдаёт текущий API_TARGET', healthJson?.api_target === UNREACHABLE);

  // Сначала проверяем недоступный API (пока цель — заведомо закрытый порт),
  // и только потом меняем цель через админку.
  const proxied = await fetch(`${base}/api/meta`);
  const proxiedBody = await readJson(proxied);
  check('недоступный API = 502', proxied.status === 502, `получено ${proxied.status}`);
  check(
    'в ответе понятная ошибка',
    proxiedBody.json?.error === 'api_unreachable',
    proxiedBody.text.slice(0, 120),
  );

  const afterError = (await readJson(await fetch(`${base}/health`))).json;
  check('health фиксирует последнюю ошибку прокси', typeof afterError?.last_proxy_error === 'string');

  const denied = await fetch(`${base}/__admin/status`);
  check('админка без токена = 401', denied.status === 401);

  const wrong = await fetch(`${base}/__admin/target`, {
    method: 'POST',
    headers: { 'content-type': 'application/json', 'x-admin-token': 'wrong' },
    body: JSON.stringify({ target: 'https://evil.example' }),
  });
  check('админка с неверным токеном = 401', wrong.status === 401);

  const updated = await fetch(`${base}/__admin/target`, {
    method: 'POST',
    headers: { 'content-type': 'application/json', 'x-admin-token': TOKEN },
    body: JSON.stringify({ target: 'https://abc123.lhr.life/' }),
  });
  const updatedJson = (await readJson(updated)).json;
  check('смена API_TARGET = 200', updated.status === 200);
  check(
    'адрес нормализован (без хвостового слэша)',
    updatedJson?.api_target === 'https://abc123.lhr.life',
    JSON.stringify(updatedJson),
  );

  const adminStatus = (await readJson(
    await fetch(`${base}/__admin/status`, { headers: { 'x-admin-token': TOKEN } }),
  )).json;
  check('админка видит новый API_TARGET', adminStatus?.api_target === 'https://abc123.lhr.life');

  const notFound = await fetch(`${base}/__admin/unknown`, {
    headers: { 'x-admin-token': TOKEN },
  });
  check('неизвестный путь админки = 404', notFound.status === 404);
} finally {
  server.close();
}

if (failures > 0) {
  console.error(`\n[smoke] провалено проверок: ${failures}`);
  process.exit(1);
}
console.log('\n[smoke] все проверки пройдены');
