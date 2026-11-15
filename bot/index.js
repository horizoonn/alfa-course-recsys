// Бот-обёртка для запуска Mini App «Альфа Курс».
// Работает на long polling (getUpdates), без внешних зависимостей и без webhook.
// Регистрирует кнопку меню и inline-кнопку с web_app URL через Bot API —
// поэтому URL в BotFather вводить не нужно.
//
// Запуск:  npm start   (читает .env через node --env-file)

const BOT_TOKEN = process.env.BOT_TOKEN;
const WEBAPP_URL = process.env.WEBAPP_URL ?? '';
const MENU_TEXT = process.env.BOT_MENU_TEXT ?? 'Открыть Альфа Курс';

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
    '1) задеплой папку frontend/dist на https://app.netlify.com/drop',
    '2) полученный https://xxxx.netlify.app впиши в bot/.env как WEBAPP_URL',
    '3) перезапусти бота и снова отправь /start',
  ].join('\n');
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
