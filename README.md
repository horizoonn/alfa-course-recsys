# alfa-course-recsys

При первом запуске создать `.env`:

```bash
cp .env.example .env
```

Проверить настройки в `.env`, при необходимости изменить их. Положить backup кейса по указанному там пути и выполнить:

```bash
docker compose up
```

Приложение: http://localhost:8080/api/health. Восстановление пустой БД, подготовка каталога, TF-IDF и генерация Go-сервера выполняются автоматически. Настройки берутся из `.env`.

Интерфейс после запуска Compose: http://localhost:5173. Mini App также можно
разрабатывать отдельно: `cd frontend`, `npm ci`, `npm run dev`.

- [Интерфейс и Telegram Mini App](frontend/README.md): запуск, сборка, проверки и обновление `bot/web`.
- [Backend API](backend/README.md): каталог и рекомендации.
- [Сервис бота и размещение](bot/README.md): постоянный HTTPS-адрес и прокси к API.
- [Вопросы к бэкенду для финальной проверки](frontend/BACKEND_HANDOFF.md).
