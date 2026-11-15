// Туннель к локальному API с автоматической публикацией адреса на Render.
//
// Зачем: бесплатные туннели (localhost.run и аналоги) выдают новый HTTPS-адрес
// при каждом обрыве. Раньше этот адрес приходилось руками менять в WEBAPP_URL,
// из-за чего Mini App «отваливался». Теперь адрес Telegram постоянный
// (https://<service>.onrender.com), а этот скрипт сам сообщает Render новый
// адрес туннеля через POST /__admin/target — передеплой не нужен.
//
// Запуск (из корня репозитория):
//   node scripts/tunnel.mjs            # порт 8080 (Go API)
//   node scripts/tunnel.mjs 4174       # если туннелируешь dev/preview фронта
//
// Переменные (можно положить в bot/.env):
//   ADMIN_URL    https://<service>.onrender.com
//   ADMIN_TOKEN  значение ADMIN_TOKEN из настроек Render
//
// Скрипт держит соединение, при обрыве переподключается и каждый раз
// публикует новый адрес. Ctrl+C — остановка.

import fs from 'node:fs';
import path from 'node:path';
import { spawn } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const HERE = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(HERE, '..');

function loadEnvFile(file) {
  if (!fs.existsSync(file)) return;
  const text = fs.readFileSync(file, 'utf8');
  for (const rawLine of text.split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!line || line.startsWith('#')) continue;
    const eq = line.indexOf('=');
    if (eq === -1) continue;
    const key = line.slice(0, eq).trim();
    let value = line.slice(eq + 1).trim();
    if (
      (value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"))
    ) {
      value = value.slice(1, -1);
    }
    if (!(key in process.env)) process.env[key] = value;
  }
}

loadEnvFile(path.join(ROOT, 'bot', '.env'));
loadEnvFile(path.join(ROOT, '.env'));

const PORT = process.argv[2] ?? '8080';
const ADMIN_URL = (process.env.ADMIN_URL ?? '').replace(/\/+$/, '');
const ADMIN_TOKEN = process.env.ADMIN_TOKEN ?? '';

const TUNNEL_URL_PATTERN = /https:\/\/[a-z0-9][a-z0-9.-]*\.(?:lhr\.life|localhost\.run)/i;
const RECONNECT_DELAY_MS = 3000;

let child = null;
let stopping = false;
let publishedUrl = '';
let pendingUrl = '';

function stamp() {
  return new Date().toLocaleTimeString('ru-RU');
}

function info(message) {
  console.log(`${stamp()} [tunnel] ${message}`);
}

async function publish(url) {
  if (!ADMIN_URL) {
    info('ADMIN_URL не задан — адрес нужно вписать вручную в API_TARGET на Render:');
    console.log(`            API_TARGET=${url}`);
    return;
  }

  if (!ADMIN_TOKEN) {
    info('ADMIN_TOKEN не задан — Render отклонит смену адреса (403/401).');
  }

  try {
    const response = await fetch(`${ADMIN_URL}/__admin/target`, {
      method: 'POST',
      headers: {
        'content-type': 'application/json',
        ...(ADMIN_TOKEN ? { 'x-admin-token': ADMIN_TOKEN } : {}),
      },
      body: JSON.stringify({ target: url }),
      signal: AbortSignal.timeout(20_000),
    });
    const data = await response.json().catch(() => ({}));

    if (response.ok && data.ok) {
      publishedUrl = url;
      info(`Render переключён на ${data.api_target ?? url}`);
    } else {
      info(`Render не принял адрес: ${response.status} ${JSON.stringify(data)}`);
      console.log(`            впиши вручную API_TARGET=${url}`);
    }
  } catch (error) {
    info(`не удалось связаться с ${ADMIN_URL}: ${error?.message ?? error}`);
    console.log(`            впиши вручную API_TARGET=${url}`);
  }
}

function handleLine(line) {
  const match = line.match(TUNNEL_URL_PATTERN);
  if (!match) return;
  const url = match[0].replace(/\/+$/, '');
  if (url === pendingUrl) return;
  pendingUrl = url;
  info(`публичный адрес туннеля: ${url}`);
  void publish(url);
}

function connect() {
  info(`подключаю 127.0.0.1:${PORT} -> localhost.run`);

  child = spawn(
    'ssh',
    [
      '-o',
      'StrictHostKeyChecking=no',
      '-o',
      'ServerAliveInterval=15',
      '-o',
      'ServerAliveCountMax=3',
      '-o',
      'TCPKeepAlive=yes',
      '-o',
      'ExitOnForwardFailure=yes',
      '-R',
      `80:127.0.0.1:${PORT}`,
      'nokey@localhost.run',
    ],
    { stdio: ['ignore', 'pipe', 'pipe'], windowsHide: true },
  );

  child.stdout.setEncoding('utf8');
  child.stderr.setEncoding('utf8');

  let buffer = '';
  const onData = (chunk) => {
    buffer += chunk;
    const lines = buffer.split(/\r?\n/);
    buffer = lines.pop() ?? '';
    for (const line of lines) handleLine(line);
    // Некоторые версии localhost.run печатают адрес без перевода строки.
    if (buffer.length > 400) buffer = '';
  };

  child.stdout.on('data', onData);
  child.stderr.on('data', onData);

  child.on('error', (error) => {
    info(`ssh недоступен: ${error?.message ?? error}`);
    info('проверь, что OpenSSH-клиент установлен: ssh -V');
  });

  child.on('exit', (code) => {
    if (stopping) return;
    info(`соединение закрыто (код ${code ?? '?'}), переподключаюсь через 3 с`);
    setTimeout(connect, RECONNECT_DELAY_MS);
  });
}

function shutdown() {
  if (stopping) return;
  stopping = true;
  info('остановка…');
  if (child) child.kill();
  setTimeout(() => process.exit(0), 300);
}

process.on('SIGINT', shutdown);
process.on('SIGTERM', shutdown);

async function main() {
  info(`порт: ${PORT}`);
  info(`Render: ${ADMIN_URL || '(ADMIN_URL не задан)'}`);
  if (ADMIN_URL && ADMIN_TOKEN) {
    try {
      const response = await fetch(`${ADMIN_URL}/__admin/status`, {
        headers: { 'x-admin-token': ADMIN_TOKEN },
        signal: AbortSignal.timeout(20_000),
      });
      const data = await response.json().catch(() => ({}));
      if (response.ok) {
        info(`Render на связи, текущий API_TARGET: ${data.api_target ?? '(пусто)'}`);
      } else {
        info(`Render ответил ${response.status} — проверь ADMIN_TOKEN`);
      }
    } catch (error) {
      info(`Render недоступен: ${error?.message ?? error}`);
    }
  }
  connect();
}

void main();
