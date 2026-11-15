// Копирует собранный фронтенд (frontend/dist) в bot/web, чтобы Render-сервис
// бота мог отдавать Mini App с постоянного адреса.
//
// Запуск (из каталога bot):
//   npm run sync-web
//
// Render не собирает фронтенд сам (Root Directory = bot), поэтому сборка
// коммитится в репозиторий. Это осознанный компромисс ради стабильного URL:
// бандл маленький и меняется только при изменениях интерфейса.

import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const HERE = path.dirname(fileURLToPath(import.meta.url));
const SOURCE = path.resolve(HERE, '..', 'frontend', 'dist');
const TARGET = path.resolve(HERE, 'web');

function fail(message) {
  console.error(`[sync-web] ${message}`);
  process.exit(1);
}

if (!fs.existsSync(path.join(SOURCE, 'index.html'))) {
  fail(
    [
      `сборка не найдена: ${SOURCE}`,
      'Сначала собери фронтенд:',
      '  cd frontend',
      '  npm run build',
    ].join('\n'),
  );
}

fs.rmSync(TARGET, { recursive: true, force: true });
fs.cpSync(SOURCE, TARGET, { recursive: true });

function walk(dir) {
  const entries = fs.readdirSync(dir, { withFileTypes: true });
  let files = 0;
  let bytes = 0;
  for (const entry of entries) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      const nested = walk(full);
      files += nested.files;
      bytes += nested.bytes;
    } else {
      files += 1;
      bytes += fs.statSync(full).size;
    }
  }
  return { files, bytes };
}

const { files, bytes } = walk(TARGET);
console.log(
  `[sync-web] ${SOURCE} -> ${TARGET}: файлов ${files}, ${(bytes / 1024).toFixed(1)} КБ`,
);
console.log('[sync-web] не забудь закоммитить bot/web — Render отдаёт Mini App из репозитория.');
