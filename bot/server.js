// HTTP-сервер Render-сервиса: отдаёт собранный фронтенд (bot/web) и проксирует
// /api/* на локальный Go-бэкенд, поднятый наружу через SSH-туннель.
//
// Зачем так:
//  - Telegram Mini App требует постоянный публичный HTTPS-адрес. Туннели
//    localhost.run/serveo/cloudflared выдают новый адрес при каждом обрыве,
//    поэтому в Telegram регистрируется адрес Render (он не меняется), а
//    нестабильным остаётся только внутренний адрес API.
//  - Адрес API хранится в переменной API_TARGET и меняется на лету через
//    POST /__admin/target — передеплой не нужен.
//  - Запросы идут same-origin, поэтому CORS в Go-бэкенде не требуется.
//
// Переменные окружения:
//   API_TARGET    адрес туннеля к локальному API, например https://abc.lhr.life
//   ADMIN_TOKEN   токен для /__admin/* (если не задан — админка выключена)
//   STATIC_DIR    каталог со сборкой фронтенда (по умолчанию ./web)

import fs from 'node:fs';
import http from 'node:http';
import path from 'node:path';
import { Readable } from 'node:stream';
import { fileURLToPath } from 'node:url';

const HERE = path.dirname(fileURLToPath(import.meta.url));

const PROXY_TIMEOUT_MS = 20_000;
const MAX_BODY_BYTES = 1024 * 1024;

const MIME_TYPES = new Map([
  ['.html', 'text/html; charset=utf-8'],
  ['.js', 'text/javascript; charset=utf-8'],
  ['.mjs', 'text/javascript; charset=utf-8'],
  ['.css', 'text/css; charset=utf-8'],
  ['.json', 'application/json; charset=utf-8'],
  ['.svg', 'image/svg+xml'],
  ['.png', 'image/png'],
  ['.jpg', 'image/jpeg'],
  ['.jpeg', 'image/jpeg'],
  ['.webp', 'image/webp'],
  ['.ico', 'image/x-icon'],
  ['.woff', 'font/woff'],
  ['.woff2', 'font/woff2'],
  ['.txt', 'text/plain; charset=utf-8'],
  ['.map', 'application/json; charset=utf-8'],
]);

const HOP_BY_HOP = new Set([
  'connection',
  'keep-alive',
  'transfer-encoding',
  'upgrade',
  'proxy-authenticate',
  'proxy-authorization',
  'te',
  'trailer',
  'host',
  'content-length',
  'accept-encoding',
]);

function normalizeTarget(value) {
  if (!value) return '';
  const trimmed = String(value).trim();
  if (!trimmed) return '';
  if (!/^https?:\/\//i.test(trimmed)) return `https://${trimmed}`;
  return trimmed.replace(/\/+$/, '');
}

function sendJson(res, status, payload) {
  const body = JSON.stringify(payload, null, 2);
  res.writeHead(status, {
    'Content-Type': 'application/json; charset=utf-8',
    'Cache-Control': 'no-store',
    'Content-Length': Buffer.byteLength(body),
  });
  res.end(body);
}

function readBody(req) {
  return new Promise((resolve, reject) => {
    const chunks = [];
    let size = 0;
    req.on('data', (chunk) => {
      size += chunk.length;
      if (size > MAX_BODY_BYTES) {
        reject(new Error('request body too large'));
        req.destroy();
        return;
      }
      chunks.push(chunk);
    });
    req.on('end', () => resolve(Buffer.concat(chunks)));
    req.on('error', reject);
  });
}

export function createAppServer(options = {}) {
  const staticDir = path.resolve(options.staticDir ?? path.join(HERE, 'web'));
  const adminToken = options.adminToken ?? '';
  const log = options.log ?? console;

  const state = {
    apiTarget: normalizeTarget(options.apiTarget),
    updatedAt: null,
    lastError: null,
    lastOkAt: null,
    proxied: 0,
  };

  const hasStatic = fs.existsSync(path.join(staticDir, 'index.html'));
  if (!hasStatic) {
    log.warn?.(`[web] сборка фронтенда не найдена: ${staticDir} (запусти npm run sync-web)`);
  }

  function setTarget(next) {
    const normalized = normalizeTarget(next);
    state.apiTarget = normalized;
    state.updatedAt = new Date().toISOString();
    state.lastError = null;
    return normalized;
  }

  function status() {
    return {
      ok: true,
      service: 'alfa-course',
      static_dir: staticDir,
      static_ready: hasStatic,
      api_target: state.apiTarget || null,
      api_target_updated_at: state.updatedAt,
      proxied_requests: state.proxied,
      last_proxy_ok_at: state.lastOkAt,
      last_proxy_error: state.lastError,
      admin_enabled: Boolean(adminToken),
      uptime_seconds: Math.round(process.uptime()),
    };
  }

  async function handleAdmin(req, res, url) {
    if (!adminToken) {
      sendJson(res, 403, {
        error: 'admin_disabled',
        hint: 'Задай ADMIN_TOKEN в переменных окружения сервиса.',
      });
      return;
    }

    const token = req.headers['x-admin-token'] ?? url.searchParams.get('token');
    if (token !== adminToken) {
      sendJson(res, 401, { error: 'unauthorized' });
      return;
    }

    if (url.pathname === '/__admin/status') {
      sendJson(res, 200, status());
      return;
    }

    if (url.pathname === '/__admin/check') {
      if (!state.apiTarget) {
        sendJson(res, 400, { error: 'api_target_not_configured' });
        return;
      }
      const probe = `${state.apiTarget}/api/health`;
      try {
        const response = await fetch(probe, { signal: AbortSignal.timeout(PROXY_TIMEOUT_MS) });
        const text = await response.text();
        sendJson(res, response.ok ? 200 : 502, {
          ok: response.ok,
          probe,
          status: response.status,
          body: text.slice(0, 500),
        });
      } catch (error) {
        sendJson(res, 502, { ok: false, probe, error: String(error?.message ?? error) });
      }
      return;
    }

    if (url.pathname === '/__admin/target') {
      if (req.method === 'GET') {
        const fromQuery = url.searchParams.get('target');
        if (!fromQuery) {
          sendJson(res, 200, status());
          return;
        }
        const applied = setTarget(fromQuery);
        log.log(`[web] API_TARGET обновлён: ${applied}`);
        sendJson(res, 200, { ok: true, api_target: applied });
        return;
      }

      if (req.method === 'POST' || req.method === 'PUT') {
        let target = url.searchParams.get('target');
        if (!target) {
          const raw = await readBody(req);
          if (raw.length) {
            try {
              const parsed = JSON.parse(raw.toString('utf8'));
              target = parsed?.target ?? parsed?.api_target ?? null;
            } catch {
              sendJson(res, 400, { error: 'invalid_json' });
              return;
            }
          }
        }
        if (!target) {
          sendJson(res, 400, { error: 'target_required' });
          return;
        }
        const applied = setTarget(target);
        log.log(`[web] API_TARGET обновлён: ${applied}`);
        sendJson(res, 200, { ok: true, api_target: applied });
        return;
      }

      sendJson(res, 405, { error: 'method_not_allowed' });
      return;
    }

    sendJson(res, 404, { error: 'not_found' });
  }

  async function handleProxy(req, res, url) {
    if (!state.apiTarget) {
      sendJson(res, 502, {
        error: 'api_target_not_configured',
        hint: 'Укажи API_TARGET (адрес туннеля к локальному API) или вызови POST /__admin/target.',
        // Фронтенд на такой ответ показывает демо-данные и плашку про офлайн-режим.
      });
      return;
    }

    const targetUrl = `${state.apiTarget}${url.pathname}${url.search}`;
    const headers = {};
    for (const [key, value] of Object.entries(req.headers)) {
      if (HOP_BY_HOP.has(key)) continue;
      if (typeof value === 'string') headers[key] = value;
      else if (Array.isArray(value)) headers[key] = value.join(', ');
    }
    headers['accept-encoding'] = 'identity';

    let body;
    if (req.method !== 'GET' && req.method !== 'HEAD') {
      try {
        const raw = await readBody(req);
        if (raw.length) body = raw;
      } catch (error) {
        sendJson(res, 413, { error: 'payload_too_large', detail: String(error?.message ?? error) });
        return;
      }
    }

    try {
      const response = await fetch(targetUrl, {
        method: req.method,
        headers,
        body,
        redirect: 'follow',
        signal: AbortSignal.timeout(PROXY_TIMEOUT_MS),
      });

      const outHeaders = {};
      response.headers.forEach((value, key) => {
        if (HOP_BY_HOP.has(key)) return;
        if (key === 'content-encoding' || key === 'content-length') return;
        outHeaders[key] = value;
      });

      state.proxied += 1;
      state.lastOkAt = new Date().toISOString();
      state.lastError = null;

      res.writeHead(response.status, outHeaders);
      if (!response.body || req.method === 'HEAD') {
        res.end();
        return;
      }
      Readable.fromWeb(response.body).pipe(res);
    } catch (error) {
      const isTimeout = error?.name === 'TimeoutError' || error?.name === 'AbortError';
      state.lastError = `${isTimeout ? 'timeout' : 'network'}: ${String(error?.message ?? error)}`;
      log.warn?.(`[web] прокси ${url.pathname}: ${state.lastError}`);
      sendJson(res, isTimeout ? 504 : 502, {
        error: isTimeout ? 'api_timeout' : 'api_unreachable',
        detail: state.lastError,
        api_target: state.apiTarget,
        hint: 'Туннель до локального API оборвался — обнови API_TARGET через POST /__admin/target.',
      });
    }
  }

  function serveStatic(req, res, url) {
    if (req.method !== 'GET' && req.method !== 'HEAD') {
      sendJson(res, 405, { error: 'method_not_allowed' });
      return;
    }

    let pathname;
    try {
      pathname = decodeURIComponent(url.pathname);
    } catch {
      pathname = url.pathname;
    }
    if (pathname === '/' || pathname === '') pathname = '/index.html';

    const candidates = [path.join(staticDir, pathname), path.join(staticDir, `${pathname}.html`)];
    let filePath = '';
    for (const candidate of candidates) {
      const relative = path.relative(staticDir, candidate);
      if (relative.startsWith('..') || path.isAbsolute(relative)) continue;
      if (fs.existsSync(candidate) && fs.statSync(candidate).isFile()) {
        filePath = candidate;
        break;
      }
    }

    // SPA-фолбэк: любой неизвестный путь отдаёт index.html.
    if (!filePath) filePath = path.join(staticDir, 'index.html');

    if (!fs.existsSync(filePath)) {
      sendJson(res, 404, {
        error: 'frontend_not_built',
        hint: 'Выполни в bot/: npm run sync-web (копирует frontend/dist в bot/web).',
      });
      return;
    }

    const ext = path.extname(filePath).toLowerCase();
    const isHashedAsset = filePath.includes(`${path.sep}assets${path.sep}`);
    const body = fs.readFileSync(filePath);

    res.writeHead(200, {
      'Content-Type': MIME_TYPES.get(ext) ?? 'application/octet-stream',
      'Content-Length': body.length,
      'Cache-Control': isHashedAsset ? 'public, max-age=31536000, immutable' : 'no-cache',
    });
    if (req.method === 'HEAD') res.end();
    else res.end(body);
  }

  const server = http.createServer((req, res) => {
    const url = new URL(req.url ?? '/', `http://${req.headers.host ?? 'localhost'}`);
    const pathname = url.pathname;

    if (pathname === '/health' || pathname === '/healthz') {
      sendJson(res, 200, status());
      return;
    }

    if (pathname.startsWith('/__admin/')) {
      void handleAdmin(req, res, url).catch((error) => {
        log.error?.('[web] админка:', error);
        if (!res.headersSent) sendJson(res, 500, { error: 'internal_error' });
      });
      return;
    }

    if (pathname === '/api' || pathname.startsWith('/api/')) {
      void handleProxy(req, res, url);
      return;
    }

    serveStatic(req, res, url);
  });

  server.keepAliveTimeout = 65_000;
  server.headersTimeout = 70_000;

  return { server, status, setTarget, staticDir };
}
