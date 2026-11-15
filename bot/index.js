// Бот-обёртка для запуска Mini App «Альфа Курс».
// Работает на long polling (getUpdates), без внешних зависимостей и без webhook.
// Регистрирует кнопку меню и inline-кнопку с web_app URL через Bot API —
// поэтому URL в BotFather вводить не нужно.
//
// Запуск:  npm start
//
// Облачный запуск (Render/Railway): сервис делает сразу три вещи:
//  1) long polling — в облаке api.telegram.org доступен, а у части российских
//     провайдеров он заблокирован, поэтому бот живёт в облаке;
//  2) отдаёт собранный фронтенд (bot/web) — это даёт Telegram постоянный
//     HTTPS-адрес Mini App, который не зависит от туннелей;
//  3) проксирует /api/* на локальный Go-бэкенд через туннель (API_TARGET),
//     поэтому CORS не нужен, а адрес туннеля меняется без передеплоя.

import fs from 'node:fs';

import { createAppServer } from './server.js';

// Загружаем .env вручную, если он есть (локальный запуск).
// В облаке (Render/Railway) файла нет — переменные приходят из окружения.
function loadDotEnv(path = '.env') {
  if (!fs.existsSync(path)) return;
  const text = fs.readFileSync(path, 'utf8');
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

loadDotEnv();

const BOT_TOKEN = process.env.BOT_TOKEN;
const PORT = Number(process.env.PORT ?? process.env.WEBAPP_PORT ?? 0) || null;
const MENU_TEXT = process.env.BOT_MENU_TEXT ?? 'Открыть Альфа Курс';

// Render сам сообщает публичный адрес сервиса — используем его как адрес
// Mini App, если WEBAPP_URL не задан вручную.
const RENDER_URL = (process.env.RENDER_EXTERNAL_URL ?? '').replace(/\/+$/, '');
const WEBAPP_URL = process.env.WEBAPP_URL ?? (RENDER_URL ? `${RENDER_URL}/` : '');

// Адрес туннеля к локальному API и токен для смены этого адреса на лету.
const API_TARGET = process.env.API_TARGET ?? '';
const ADMIN_TOKEN = process.env.ADMIN_TOKEN ?? '';
const KEEPALIVE = ['1', 'true', 'yes'].includes((process.env.KEEPALIVE ?? '').toLowerCase());

const API_BASE = `https://api.telegram.org/bot${BOT_TOKEN}`;
const POLL_TIMEOUT_SECONDS = 30;
const RETRY_DELAY_MS = 3000;

if (!BOT_TOKEN) {
  console.error(
    [
      'BOT_TOKEN не задан.',
      '',
      '1. Создай .env:  copy .env.example .env',
      '2. Вставь токен из BotFather: /mybots -> выбери бота -> API Token',
      '3. Запусти:  npm start',
    ].join('\n'),
  );
  process.exit(1);
}

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

function isPublicHttps(url) {
  try {
    const parsed = new URL(url);
    if (parsed.protocol !== 'https:') return false;
    return !['localhost', '127.0.0.1', '0.0.0.0'].includes(parsed.hostname);
  } catch {
    return false;
  }
}

async function checkTelegramReachable() {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), 10_000);
  try {
    const response = await fetch(`${API_BASE}/getMe`, { signal: controller.signal });
    const data = await response.json();
    if (!data.ok) {
      console.error(`[bot] Telegram отклонил токен: ${data.description ?? response.status}`);
      return false;
    }
    console.log(`[bot] подключён как @${data.result.username} (${data.result.first_name})`);
    return true;
  } catch {
    console.error(
      [
        '[bot] нет доступа к api.telegram.org — бот не сможет получать сообщения.',
        '',
        'Причины и решения:',
        '  1) Сеть блокирует Telegram (частая ситуация у российских провайдеров):',
        '     включи VPN и запусти снова, либо',
        '  2) Запусти бота в облаке (там блокировки нет): Render/Railway,',
        '     команда запуска: npm start, переменные: BOT_TOKEN, WEBAPP_URL,',
        '     переменную PORT платформа задаст сама.',
      ].join('\n'),
    );
    return false;
  } finally {
    clearTimeout(timer);
  }
}

async function call(method, payload = {}) {
  const response = await fetch(`${API_BASE}/${method}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });

  let data;
  try {
    data = await response.json();
  } catch {
    throw new Error(`Telegram API ${method}: некорректный ответ (${response.status})`);
  }

  if (!response.ok || !data.ok) {
    throw new Error(`Telegram API ${method}: ${data.description ?? response.status}`);
  }
  return data.result;
}

function launchKeyboard() {
  return {
    inline_keyboard: [[{ text: MENU_TEXT, web_app: { url: WEBAPP_URL } }]],
  };
}

function noUrlHint() {
  return [
    'Mini App пока не подключён: в WEBAPP_URL нужен публичный HTTPS-адрес.',
    `Сейчас указано: ${WEBAPP_URL || '(пусто)'}`,
    '',
    'Как получить адрес:',
    '1) задеплой сервис на Render (Root Directory: bot, Start Command: npm start)',
    '   — адрес https://<service>.onrender.com/ подставится автоматически;',
    '2) либо собери фронт (npm run build в frontend) и выложи frontend/dist',
    '   на https://app.netlify.com/drop, затем впиши адрес в WEBAPP_URL;',
    '3) перезапусти бота и снова отправь /start.',
  ].join('\n');
}

function startWebServer() {
  const { server } = createAppServer({
    staticDir: process.env.STATIC_DIR ?? undefined,
    apiTarget: API_TARGET,
    adminToken: ADMIN_TOKEN,
  });

  server.listen(PORT, () => {
    console.log(`[web] Mini App и прокси /api на порту ${PORT}`);
    console.log(`[web] адрес приложения: ${WEBAPP_URL || '(WEBAPP_URL не задан)'}`);
    console.log(`[web] API_TARGET: ${API_TARGET || '(не задан — фронт покажет демо-данные)'}`);
    console.log(`[web] админка: ${ADMIN_TOKEN ? 'включена (/__admin/target)' : 'выключена'}`);
  });

  // Бесплатный инстанс Render засыпает после 15 минут без входящих запросов.
  // Самопинг держит сервис (а значит и long polling) активным.
  if (KEEPALIVE && RENDER_URL) {
    const interval = 5 * 60 * 1000;
    setInterval(() => {
      fetch(`${RENDER_URL}/health`).catch(() => {});
    }, interval).unref();
    console.log('[web] KEEPALIVE включён: самопинг каждые 5 минут');
  }
}

async function sendMessage(chatId, text, replyMarkup) {
  const payload = { chat_id: chatId, text, disable_web_page_preview: true };
  if (replyMarkup) payload.reply_markup = replyMarkup;
  await call('sendMessage', payload);
}

async function handleUpdate(update) {
  const message = update.message;
  const text = message?.text?.trim();
  if (!message || !text) return;

  const chatId = message.chat.id;
  const firstName = message.from?.first_name;

  if (text.startsWith('/start')) {
    const greeting = firstName ? `Привет, ${firstName}!` : 'Привет!';
    if (!isPublicHttps(WEBAPP_URL)) {
      await sendMessage(chatId, [greeting, '', noUrlHint()].join('\n'));
      return;
    }
    await sendMessage(
      chatId,
      [
        greeting,
        '',
        'Это «Альфа Курс» — подберём образовательный материал под вашу задачу',
        'и подскажем следующий шаг.',
        '',
        'Нажмите кнопку ниже, чтобы открыть приложение.',
      ].join('\n'),
      launchKeyboard(),
    );
    return;
  }

  if (text.startsWith('/app')) {
    if (!isPublicHttps(WEBAPP_URL)) {
      await sendMessage(chatId, noUrlHint());
      return;
    }
    await sendMessage(chatId, 'Нажмите кнопку ниже, чтобы открыть приложение.', launchKeyboard());
    return;
  }

  if (text.startsWith('/help')) {
    await sendMessage(
      chatId,
      ['/start — приветствие и кнопка запуска', '/app — открыть приложение'].join('\n'),
    );
  }
}

async function setupBot() {
  await call('setMyCommands', {
    commands: [
      { command: 'start', description: 'Запуск и открытие приложения' },
      { command: 'app', description: 'Открыть Альфа Курс' },
      { command: 'help', description: 'Помощь' },
    ],
  });

  if (!isPublicHttps(WEBAPP_URL)) {
    console.warn(
      `[bot] WEBAPP_URL="${WEBAPP_URL}" не является публичным HTTPS — кнопка меню не зарегистрирована.`,
    );
    return;
  }

  // Кнопка слева от поля ввода — открывает Mini App без команд.
  await call('setChatMenuButton', {
    menu_button: {
      type: 'web_app',
      text: MENU_TEXT,
      web_app: { url: WEBAPP_URL },
    },
  });

  console.log(`[bot] кнопка меню зарегистрирована: ${WEBAPP_URL}`);
}

async function poll() {
  let offset = 0;
  let running = true;

  const stop = () => {
    running = false;
    console.log('[bot] остановка…');
  };
  process.on('SIGINT', stop);
  process.on('SIGTERM', stop);

  while (running) {
    try {
      const updates = await call('getUpdates', {
        offset,
        timeout: POLL_TIMEOUT_SECONDS,
        allowed_updates: ['message'],
      });

      for (const update of updates) {
        offset = update.update_id + 1;
        try {
          await handleUpdate(update);
        } catch (error) {
          console.error('[bot] ошибка обработки обновления:', error.message);
        }
      }
    } catch (error) {
      console.error('[bot] ошибка polling:', error.message);
      await sleep(RETRY_DELAY_MS);
    }
  }
}

async function main() {
  console.log('[bot] запуск…');

  // В облаке платформе нужен открытый порт: он же отдаёт Mini App.
  if (PORT) {
    startWebServer();
  }

  const reachable = await checkTelegramReachable();
  if (!reachable && !PORT) {
    console.error('[bot] остановлен: сначала реши проблему доступа к Telegram API.');
    process.exit(1);
  }

  // На случай, если у бота был настроен webhook — снимаем, чтобы работал polling.
  try {
    await call('deleteWebhook', { drop_pending_updates: false });
  } catch (error) {
    console.warn('[bot] deleteWebhook:', error.message);
  }

  try {
    await setupBot();
  } catch (error) {
    console.error('[bot] настройка не удалась:', error.message);
  }

  console.log('[bot] слушаю сообщения. Отправьте /start в чат с ботом.');
  await poll();
}

void main();
