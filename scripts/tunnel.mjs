// Туннель к локальному API с автоматической публикацией адреса на Render.
//
// Зачем: бесплатные туннели (localhost.run и аналоги) выдают новый HTTPS-адрес
// при каждом обрыве. Раньше этот адрес приходилось руками менять в WEBAPP_URL,
// из-за чего Mini App «отваливался». Теперь адрес Telegram постоянный
// (https://<service>.onrender.com), а этот скрипт сам сообщает Render новый
// адрес туннеля через POST /__admin/target — передеплой не нужен.
//
// Запуск (из корня репозитория, где лежат bot/ и scripts/):
//   node scripts/tunnel.mjs              # порт 8080 (Go API)
//   node scripts/tunnel.mjs 4174         # другой порт
//   node scripts/tunnel.mjs --probe      # проверить, что туннель вообще встаёт
//
// Что вставлять:
//   ADMIN_URL   — адрес сервиса на Render, как в адресной строке браузера:
//                 https://alfa-course-bot-xxxx.onrender.com
//                 (без /health, без /api, без слэша на конце)
//   ADMIN_TOKEN — строка, которую ты задал в Render → Environment → ADMIN_TOKEN
//
// Оба значения можно положить в bot/.env. Если ADMIN_URL не задан, скрипт
// спросит его и сохранит в bot/.env, чтобы больше не спрашивать.

import fs from 'node:fs';
import path from 'node:path';
import { spawn } from 'node:child_process';
import { createInterface } from 'node:readline';
import { fileURLToPath } from 'node:url';

const HERE = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(HERE, '..');
const BOT_ENV = path.join(ROOT, 'bot', '.env');

const args = process.argv.slice(2);
const PROBE = args.includes('--probe');
const PORT = args.find((arg) => !arg.startsWith('-')) ?? '8080';

// localhost.run в приветствии печатает ссылки на админку и документацию —
// их нельзя принимать за адрес туннеля. Настоящий туннель выглядит так:
//   https://0b20cae8d31b68.app.lhr.life
//   https://abc123.lhr.life
const TUNNEL_URL_PATTERN = /https:\/\/([a-z0-9][a-z0-9.-]*\.lhr\.life)/gi;
const RESERVED_HOSTS = new Set(['admin.lhr.life', 'www.lhr.life', 'docs.lhr.life']);

function findTunnelUrl(line) {
  for (const match of line.matchAll(TUNNEL_URL_PATTERN)) {
    const host = match[1].toLowerCase().replace(/\.+$/, '');
    if (RESERVED_HOSTS.has(host)) continue;
    return `https://${host}`;
  }
  return '';
}
const RECONNECT_DELAY_MS = 3000;
const PROBE_TIMEOUT_MS = 35_000;

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

function saveEnvValue(key, value) {
  const line = `${key}=${value}`;
  let text = fs.existsSync(BOT_ENV) ? fs.readFileSync(BOT_ENV, 'utf8') : '';
  const pattern = new RegExp(`^${key}=.*$`, 'm');
  if (pattern.test(text)) {
    text = text.replace(pattern, line);
  } else {
    if (text && !text.endsWith('\n')) text += '\n';
    text += `${line}\n`;
  }
  fs.mkdirSync(path.dirname(BOT_ENV), { recursive: true });
  fs.writeFileSync(BOT_ENV, text, 'utf8');
}

function stamp() {
  return new Date().toLocaleTimeString('ru-RU');
}

function info(message) {
  console.log(`${stamp()} [tunnel] ${message}`);
}

function hint(lines) {
  for (const line of lines) console.log(`            ${line}`);
}

async function ask(question) {
  const rl = createInterface({ input: process.stdin, output: process.stdout });
  try {
    return await new Promise((resolve) => rl.question(question, resolve));
  } finally {
    rl.close();
  }
}

async function resolveAdminConfig() {
  let adminUrl = (process.env.ADMIN_URL ?? '').replace(/\/+$/, '');
  let adminToken = process.env.ADMIN_TOKEN ?? '';

  if (!adminUrl && process.stdin.isTTY && !PROBE) {
    console.log('');
    info('адрес сервиса Render не задан (ADMIN_URL).');
    hint([
      'Открой «Альфа Курс» в браузере и скопируй адрес из строки браузера.',
      'Пример: https://alfa-course-bot-abcd.onrender.com',
      'Нужен только домен, без /health, /api и без слэша на конце.',
      '',
    ]);
    const answer = (await ask('Вставь адрес сервиса Render (Enter — пропустить): ')).trim();
    if (answer) {
      adminUrl = answer.replace(/\/+$/, '');
      if (!/^https?:\/\//i.test(adminUrl)) adminUrl = `https://${adminUrl}`;
      saveEnvValue('ADMIN_URL', adminUrl);
      info(`сохранил в bot/.env: ADMIN_URL=${adminUrl}`);
    }
  }

  if (adminUrl && !adminToken && process.stdin.isTTY && !PROBE) {
    console.log('');
    info('ADMIN_TOKEN не задан — без него Render не примет новый адрес.');
    hint([
      '1) Render → твой сервис → Environment → Add Environment Variable',
      '2) Key: ADMIN_TOKEN, Value: любая длинная строка (например, ниже)',
      '3) Save (сервис передеплоится), затем запусти скрипт снова',
      '',
    ]);
    const suggestion = Buffer.from(cryptoRandomHex(16)).toString('hex');
    info(`можно взять такой токен: ${suggestion}`);
    const answer = (await ask('Вставь ADMIN_TOKEN из Render (Enter — пропустить): ')).trim();
    if (answer) {
      adminToken = answer;
      saveEnvValue('ADMIN_TOKEN', adminToken);
      info('сохранил в bot/.env: ADMIN_TOKEN=***');
    }
  }

  return { adminUrl, adminToken };
}

function cryptoRandomHex(bytes) {
  // Без node:crypto ради читаемости: значения всё равно только для демо.
  const alphabet = 'abcdef0123456789';
  let out = '';
  for (let i = 0; i < bytes * 2; i += 1) {
    out += alphabet[Math.floor(Math.random() * alphabet.length)];
  }
  return out;
}

async function publish(adminUrl, adminToken, url) {
  if (!adminUrl) {
    info('ADMIN_URL не задан — впиши адрес вручную в Render → API_TARGET:');
    hint([url]);
    return false;
  }

  try {
    const response = await fetch(`${adminUrl}/__admin/target`, {
      method: 'POST',
      headers: {
        'content-type': 'application/json',
        ...(adminToken ? { 'x-admin-token': adminToken } : {}),
      },
      body: JSON.stringify({ target: url }),
      signal: AbortSignal.timeout(20_000),
    });
    const data = await response.json().catch(() => ({}));

    if (response.ok && data.ok) {
      info(`Render переключён на ${data.api_target ?? url}`);
      return true;
    }

    info(`Render не принял адрес: ${response.status} ${JSON.stringify(data)}`);
    if (response.status === 401 || response.status === 403) {
      hint([
        'ADMIN_TOKEN не совпадает с тем, что в Render, либо не задан на Render.',
        `Проверь: ${adminUrl}/health → admin_enabled должно быть true`,
        `Либо впиши вручную в Render → API_TARGET: ${url}`,
      ]);
    } else {
      hint([`Впиши вручную в Render → API_TARGET: ${url}`]);
    }
    return false;
  } catch (error) {
    info(`не удалось связаться с ${adminUrl}: ${error?.message ?? error}`);
    hint([
      'Проверь, что сервис на Render проснулся и открывается в браузере.',
      `Либо впиши вручную в Render → API_TARGET: ${url}`,
    ]);
    return false;
  }
}

async function checkRender(adminUrl, adminToken) {
  if (!adminUrl) return;
  info(`Render: ${adminUrl}`);
  try {
    const response = await fetch(`${adminUrl}/health`, { signal: AbortSignal.timeout(20_000) });
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      info(`Render ответил ${response.status} — сервис ещё собирается или упал`);
      return;
    }
    info(
      `Render на связи: статика ${data.static_ready ? 'ok' : 'НЕТ'}, ` +
        `API_TARGET ${data.api_target ?? '(пусто)'}, админка ${data.admin_enabled ? 'вкл' : 'выкл'}`,
    );
    if (!data.admin_enabled && adminToken) {
      hint(['ADMIN_TOKEN на Render не задан — адрес придётся вписывать вручную.']);
    }
  } catch (error) {
    info(`Render недоступен: ${error?.message ?? error}`);
  }
}

function connect({ adminUrl, adminToken }) {
  info(`подключаю 127.0.0.1:${PORT} -> localhost.run`);

  const child = spawn(
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
      '-o',
      'ConnectTimeout=15',
      '-R',
      `80:127.0.0.1:${PORT}`,
      'nokey@localhost.run',
    ],
    { stdio: ['ignore', 'pipe', 'pipe'], windowsHide: true },
  );

  child.stdout.setEncoding('utf8');
  child.stderr.setEncoding('utf8');

  let buffer = '';
  let lastUrl = '';
  let spawnFailed = false;

  const onData = (chunk) => {
    if (PROBE) process.stdout.write(chunk);
    buffer += chunk;
    const lines = buffer.split(/\r?\n/);
    buffer = lines.pop() ?? '';
    for (const line of lines) {
      const url = findTunnelUrl(line);
      if (!url) continue;
      if (url === lastUrl) continue;
      lastUrl = url;
      info(`публичный адрес туннеля: ${url}`);
      if (PROBE) {
        void (async () => {
          // Первый запрос может попасть на край туннеля, который ещё не
          // зарегистрировал соединение, поэтому пробуем несколько раз.
          for (let attempt = 1; attempt <= 3; attempt += 1) {
            try {
              const response = await fetch(`${url}/api/health`, {
                signal: AbortSignal.timeout(20_000),
              });
              const body = (await response.text()).slice(0, 120);
              info(`проверка ${url}/api/health -> ${response.status} ${body}`);
              info(
                response.ok
                  ? 'туннель до API работает: Render сможет отдавать реальные рекомендации'
                  : 'туннель поднят, но API ответил ошибкой — проверь docker compose ps',
              );
              child.kill();
              process.exit(response.ok ? 0 : 1);
            } catch (error) {
              info(
                `попытка ${attempt}/3: ${error?.message ?? error}` +
                  (error?.cause?.code ? ` (${error.cause.code})` : ''),
              );
              await new Promise((resolve) => setTimeout(resolve, 2000));
            }
          }
          info('туннель поднят, но публичный адрес не отвечает');
          child.kill();
          process.exit(1);
        })();
        return;
      }
      void publish(adminUrl, adminToken, url);
    }
    if (buffer.length > 2000) buffer = '';
  };

  child.stdout.on('data', onData);
  child.stderr.on('data', onData);

  child.on('error', (error) => {
    spawnFailed = true;
    info(`не удалось запустить ssh: ${error?.message ?? error}`);
    hint([
      'OpenSSH-клиент не найден. Проверь в новом терминале: ssh -V',
      'Windows: Параметры → Приложения → Дополнительные компоненты → OpenSSH-клиент.',
      'Либо запусти scripts\\tunnel.bat — он использует тот же ssh.',
    ]);
    process.exit(1);
  });

  child.on('exit', (code) => {
    if (spawnFailed || PROBE) return;
    info(`соединение закрыто (код ${code ?? '?'}), переподключаюсь через 3 с`);
    setTimeout(() => connect({ adminUrl, adminToken }), RECONNECT_DELAY_MS);
  });

  if (PROBE) {
    setTimeout(() => {
      info('за 35 секунд адрес туннеля не получен — localhost.run недоступен');
      child.kill();
      process.exit(1);
    }, PROBE_TIMEOUT_MS);
  }
}

async function main() {
  if (!fs.existsSync(path.join(ROOT, 'bot'))) {
    console.error(
      '[tunnel] запусти скрипт из корня репозитория: node scripts\\tunnel.mjs',
    );
    process.exit(1);
  }

  const config = await resolveAdminConfig();
  info(`порт API: ${PORT}${PROBE ? ' (режим проверки)' : ''}`);
  await checkRender(config.adminUrl, config.adminToken);

  const shutdown = () => {
    info('остановка…');
    process.exit(0);
  };
  process.on('SIGINT', shutdown);
  process.on('SIGTERM', shutdown);

  connect(config);
}

void main();
