# Сервис Mini App «Альфа Курс» на Render

Один Node-сервис (без зависимостей, Node 20.6+) делает три вещи:

1. **long polling** Telegram — по `/start` присылает кнопку, которая открывает Mini App,
   и регистрирует кнопку меню слева от поля ввода;
2. **отдаёт собранный фронтенд** из `bot/web` — это даёт Telegram постоянный
   HTTPS-адрес, который не меняется при обрывах туннеля;
3. **проксирует `/api/*`** на локальный Go-бэкенд, поднятый наружу через SSH-туннель.

Боевой адрес: **https://alfa-course-recsys.onrender.com**

```
Telegram ──> https://alfa-course-recsys.onrender.com/   (Render, постоянный адрес)
                 │
                 ├── /            -> bot/web (SPA)
                 ├── /api/*       -> туннель -> 127.0.0.1:8080 (Go API + Postgres)
                 └── long polling -> api.telegram.org

Адрес туннеля: GitHub, ветка tunnel, файл bot/tunnel-url.txt
Render читает его каждые две минуты -> менять переменные в панели не нужно
```

Почему так: бесплатные туннели (localhost.run, serveo, trycloudflare) выдают новый
адрес при каждом обрыве, а у части российских провайдеров `onrender.com` вообще
недоступен из локальной сети — ни панель, ни сам сервис. Поэтому адрес туннеля
публикуется через `github.com` (он доступен всегда), а сервис сам его подхватывает.

**URL в BotFather вводить не нужно:** бот сам вызывает `setChatMenuButton`
и передаёт inline-кнопку `web_app` через Bot API.

## Переменные окружения

| Переменная | Обязательна | Назначение |
|---|---|---|
| `BOT_TOKEN` | да | токен из BotFather |
| `WEBAPP_URL` | нет | адрес Mini App. На Render **игнорируется**: всегда берётся `RENDER_EXTERNAL_URL`, чтобы старая переменная не увела кнопку в мёртвый туннель |
| `TUNNEL_STATE_URL` | нет | откуда читать адрес туннеля; по умолчанию ветка `tunnel` этого репозитория |
| `API_TARGET` | нет | стартовый адрес туннеля (перекрывается значением из GitHub) |
| `ADMIN_TOKEN` | нет | включает `/__admin/*` — смена `API_TARGET` по HTTP, если до Render есть доступ |
| `KEEPALIVE` | нет | по умолчанию включён на Render: самопинг `/health` каждые 5 минут, чтобы инстанс не засыпал |
| `STATIC_DIR` | нет | каталог со сборкой, по умолчанию `./web` |
| `PORT` | на Render задаёт платформа | порт HTTP-сервера |

## Деплой на Render

1. New → **Web Service** → репозиторий `alfa-course-recsys`, ветка `front`
   (уже создан: сервис `alfa-course-recsys`, Node, Free, Oregon);
2. **Root Directory**: `bot`, **Build Command**: `npm install`, **Start Command**: `npm start`,
   план Free, Auto-Deploy: On;
3. Environment: достаточно `BOT_TOKEN`. Остальное работает по умолчанию;
4. в логах после деплоя:

```
[bot] подключён как @your_bot
[web] Mini App и прокси /api на порту 10000
[web] адрес приложения: https://your-service.onrender.com/
[web] API_TARGET обновлён из GitHub: https://....app.lhr.life
[bot] кнопка меню зарегистрирована: https://your-service.onrender.com/
```

## Обновить фронтенд в сервисе

Render собирает только каталог `bot`, поэтому сборка фронтенда коммитится в `bot/web`:

```bat
cd c:\hack_alfa\frontend
npm run build
cd ..\bot
npm run sync-web
git add web && git commit -m "chore(bot): обновить сборку фронтенда" && git push
```

## Туннель к локальному API

Go-бэкенд и Postgres остаются на ноутбуке (Docker), наружу — только API:

```bat
:: 1) поднять стек
cd c:\hack_alfa
docker compose up -d
curl http://127.0.0.1:8080/api/health

:: 2) туннель: держит соединение, при обрыве переподключается
::    и после каждого переподключения публикует новый адрес в GitHub
node scripts\tunnel.mjs
```

Скрипт делает всё сам: поднимает SSH-туннель, ловит выданный адрес
(`https://<id>.app.lhr.life`), коммитит его в ветку `tunnel` и пушит.
Render подхватывает адрес в течение двух минут — Telegram и переменные окружения
трогать не нужно. Ветка `tunnel` не деплоится, поэтому пуши не вызывают пересборку.

Полезные режимы:

```bat
node scripts\tunnel.mjs --probe          :: проверить туннель и доступность API по публичному адресу
node scripts\tunnel.mjs --seconds=30     :: протестировать публикацию адреса и выйти
node scripts\tunnel.mjs 4174             :: туннелировать другой порт
```

Если GitHub-канал недоступен, адрес копируется в буфер обмена — тогда его нужно
вписать в Render → Environment → `API_TARGET` вручную.

## Диагностика

| Запрос | Что показывает |
|---|---|
| `GET /health` | готова ли статика, текущий `API_TARGET`, время последнего успешного прокси, текст последней ошибки |
| `GET /__admin/status` (с `x-admin-token`) | то же + число проксированных запросов |
| `GET /__admin/check` | сервис сам проверяет `<API_TARGET>/api/health` |
| `POST /__admin/target` | смена `API_TARGET` по HTTP |

Локальный смоук-тест без Telegram и Docker:

```bat
cd c:\hack_alfa\bot
npm run smoke
```

## Локальный запуск (если Telegram доступен)

```bat
cd c:\hack_alfa\bot
copy .env.example .env
:: вписать BOT_TOKEN, при необходимости WEBAPP_URL и WEBAPP_PORT
set WEBAPP_PORT=3000
npm start
```

Без `PORT`/`WEBAPP_PORT` сервис не поднимает HTTP-сервер и работает только как бот.

## Команды бота

| Команда | Действие |
|---|---|
| `/start` | приветствие + кнопка запуска |
| `/app` | ещё раз прислать кнопку |
| `/help` | список команд |

## Если что-то не работает

| Симптом | Причина и решение |
|---|---|
| `BOT_TOKEN не задан` | нет `bot/.env` или пустой токен |
| `Telegram API ...: Unauthorized` | неверный токен |
| `кнопка меню не зарегистрирована` | `WEBAPP_URL` не публичный HTTPS (на Render проверь `RENDER_EXTERNAL_URL`) |
| Mini App сообщает о недоступности каталога | проверить API на порту 8080 и перезапустить `node scripts\tunnel.mjs`; рабочий адрес будет опубликован автоматически |
| `502 api_unreachable` в `/health` | то же; текст ошибки — в `last_proxy_error` |
| `Conflict: terminated by other getUpdates request` | запущено два экземпляра бота |
| Сервис «спит», бот не отвечает | бесплатный инстанс засыпает без входящих запросов; на Render `KEEPALIVE` включён по умолчанию |

## Ограничения MVP

- long polling, без webhook и без БД — состояние не хранится;
- `initData` от Telegram на бэкенд не отправляется (для трека 1 не требуется);
- адрес туннеля проверяется каждые две минуты; перед переключением сервис
  проверяет `/api/health`, чтобы не принять отключённый туннель;
- рабочая сборка не заменяет реальные рекомендации учебными примерами:
  при недоступном API показывает ошибку и повтор запроса;
- уведомления и напоминания не реализованы — это функциональность трека 2.
