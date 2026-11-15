import { useEffect, useState } from 'react';
import type { TelegramWebApp } from '../telegram';

const SDK_URL = 'https://telegram.org/js/telegram-web-app.js';
const POLL_INTERVAL_MS = 150;
const POLL_TIMEOUT_MS = 5000;

/**
 * SDK загружается асинхронно: в обычном браузере (вне Telegram) страница
 * не должна зависеть от доступности telegram.org. Внутри Telegram объект
 * window.Telegram обычно уже внедрён клиентом или подхватится за долю секунды.
 */
function loadSdk(): Promise<TelegramWebApp | null> {
  return new Promise((resolve) => {
    const startedAt = Date.now();
    const poll = setInterval(() => {
      const wa = window.Telegram?.WebApp;
      if (wa) {
        clearInterval(poll);
        resolve(wa);
        return;
      }
      if (!document.querySelector('script[data-tg-sdk]')) {
        const script = document.createElement('script');
        script.src = SDK_URL;
        script.dataset.tgSdk = 'true';
        document.head.appendChild(script);
      }
      if (Date.now() - startedAt > POLL_TIMEOUT_MS) {
        clearInterval(poll);
        resolve(null);
      }
    }, POLL_INTERVAL_MS);
  });
}

function initWebApp(webApp: TelegramWebApp) {
  webApp.ready();
  webApp.expand();
  try {
    webApp.setHeaderColor(webApp.themeParams.bg_color ?? 'bg_color');
    webApp.setBackgroundColor(webApp.themeParams.bg_color ?? 'bg_color');
  } catch {
    // старые клиенты не поддерживают методы цвета — не критично
  }
}

/**
 * В обычном браузере возвращает isTelegram=false —
 * интерфейс остаётся полностью работоспособным.
 */
export function useTelegram() {
  const [webApp, setWebApp] = useState<TelegramWebApp | null>(
    () => window.Telegram?.WebApp ?? null,
  );

  useEffect(() => {
    if (webApp) {
      initWebApp(webApp);
      return;
    }
    let cancelled = false;
    void loadSdk().then((wa) => {
      if (!cancelled && wa) {
        setWebApp(wa);
        initWebApp(wa);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [webApp]);

  const haptic = (style: 'light' | 'medium' | 'heavy') => {
    webApp?.HapticFeedback?.impactOccurred(style);
  };

  const notify = (type: 'error' | 'success' | 'warning') => {
    webApp?.HapticFeedback?.notificationOccurred(type);
  };

  return {
    webApp,
    isTelegram: webApp !== null,
    userName: webApp?.initDataUnsafe?.user?.first_name ?? null,
    haptic,
    notify,
  };
}
