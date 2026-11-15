# Сервис Mini App «Альфа Курс» на Render

Один Node-сервис (без зависимостей, Node 20.6+) делает три вещи:

1. **long polling** Telegram — по `/start` присылает кнопку, которая открывает Mini App,
   и регистрирует кнопку меню слева от поля ввода;
2. **отдаёт собранный фронтенд** из `bot/web` — это даёт Telegram постоянный
   HTTPS-адрес, который не меняется при обрывах туннеля;
3. **проксирует `/api/*`** на локальный Go-бэкенд, поднятый наружу через SSH-туннель.

```
Telegram ──> https://<service>.onrender.com/  (Render, постоянный адрес)
                    │
                    ├── /            -> bot/web (SPA)
                    ├── /api/*       -> туннель -> 127.0.0.1:8080 (Go API + Postgres)
                    └── long polling -> api.telegram.org
```

Почему так: бесплатные туннели (localhost.run, serveo, trycloudflare) выдают новый
адрес при каждом обрыве. Раньше адрес приходилось руками менять в `WEBAPP_URL`,
и Mini App «отваливался». Теперь в Telegram зарегистрирован адрес Render,
а меняется только внутренний `API_TARGET` — и это делается одной командой,
без передеплоя.

**URL в BotFather вводить не нужно:** бот сам вызывает `setChatMenuButton`
и передаёт inline-кнопку `web_app` через Bot API.

## Переменные окружения

| Переменная | Обязательна | Назначение |
|---|---|---|
| `BOT_TOKEN` | да | токен из BotFather |
| `WEBAPP_URL` | нет | адрес Mini App; на Render подставляется из `RENDER_EXTERNAL_URL` автоматически |
| `API_TARGET` | да для реальных данных | адрес туннеля к локальному API, например `https://abc123.lhr.life` |
| `ADMIN_TOKEN` | рекомендуется | токен для `/__admin/*` (смена `API_TARGET` без передеплоя) |
| `KEEPALIVE` | нет | `1` — самопинг `/health` каждые 5 минут, чтобы бесплатный инстанс не засыпал |
| `STATIC_DIR` | нет | каталог со сборкой, по умолчанию `./web` |
| `PORT` | на Render задаёт платформа | порт HTTP-сервера |

## Деплой на Render

1. New → **Web Service** → репозиторий `alfa-course-recsys`, ветка `front`;
2. **Root Directory**: `bot`, **Build Command**: `npm install`, **Start Command**: `npm start`,
   план Free;
3. Environment:
   - `BOT_TOKEN` — из BotFather;
   - `ADMIN_TOKEN` — придумать длинную строку (например, `openssl rand -hex 16`);
   - `KEEPALIVE` = `1`;
   - `API_TARGET` — текущий адрес туннеля (можно не задавать сразу: обновится скриптом);
4. после деплоя в логах должно быть:

```
[bot] подключён как @your_bot (@your_bot)
[web] Mini App и прокси /api на порту 10000
[web] адрес приложения: https://your-service.onrender.com/
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

Render пересоберёт сервис автоматически (1–2 минуты).

## Туннель к локальному API

Go-бэкенд и Postgres остаются на ноутбуке (Docker), наружу — только API:

```bat
:: 1) поднять стек
cd c:\hack_alfa
docker compose up -d
curl http://127.0.0.1:8080/api/health

:: 2) в bot\.env добавить один раз:
::    ADMIN_URL=https://your-service.onrender.com
::    ADMIN_TOKEN=<тот же, что в Render>

:: 3) туннель с автопубликацией адреса
node scripts\tunnel.mjs
```

Скрипт держит SSH-соединение, при обрыве переподключается и после каждого
переподключения сам отправляет новый адрес в Render (`POST /__admin/target`).
Mini App при этом продолжает работать — адрес в Telegram не меняется.

Если скрипт не используется, адрес можно обновить вручную:

```bat
curl -X POST https://your-service.onrender.com/__admin/target ^
  -H "content-type: application/json" ^
  -H "x-admin-token: <ADMIN_TOKEN>" ^
  -d "{\"target\":\"https://new-address.lhr.life\"}"
```

## Диагностика

| Запрос | Что показывает |
|---|---|
| `GET /health` | состояние сервиса: готова ли статика, текущий `API_TARGET`, время последнего успешного прокси и текст последней ошибки |
| `GET /__admin/status` (с `x-admin-token`) | то же + сколько запросов проксировано |
| `GET /__admin/check` | проверка «жив ли туннель»: сервис сам дёргает `<API_TARGET>/api/health` |
| `POST /__admin/target` | смена `API_TARGET` на лету |

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
| `Telegram API ...: Unauthorized` | неверный токен (взят от другого бота или с пробелом) |
| `кнопка меню не зарегистрирована` | `WEBAPP_URL` не публичный HTTPS; на Render проверь, что `RENDER_EXTERNAL_URL` доступен |
| Mini App открывается, но показывает плашку «демо-данные» | туннель до API лёг: перезапусти `node scripts\tunnel.mjs` или обнови `API_TARGET` |
| `502 api_unreachable` в `/health` | то же самое; в `last_proxy_error` будет текст ошибки |
| `Conflict: terminated by other getUpdates request` | запущено два экземпляра бота — останови лишний |
| Сервис на Render «спит», бот не отвечает | бесплатный инстанс засыпает через 15 минут без входящих запросов; включи `KEEPALIVE=1` или пингуй `/health` внешним cron (например, cron-job.org) |

## Ограничения MVP

- long polling, без webhook и без БД — состояние не хранится;
- `initData` от Telegram на бэкенд не отправляется (для трека 1 не требуется);
- `API_TARGET`, изменённый через админку, живёт до перезапуска инстанса
  (после рестарта берётся значение из переменных окружения);
- уведомления и напоминания не реализованы — это функциональность трека 2.
