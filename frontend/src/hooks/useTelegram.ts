import { useCallback, useEffect, useState } from 'react';
import type { TelegramWebApp } from '../telegram';

const SDK_URL = 'https://telegram.org/js/telegram-web-app.js';
const POLL_INTERVAL_MS = 150;
const POLL_TIMEOUT_MS = 5000;
let sdkPromise: Promise<TelegramWebApp | null> | null = null;
const initialized = new WeakSet<TelegramWebApp>();

function currentWebApp(): TelegramWebApp | null {
  if (typeof window === 'undefined') return null;
  const app = window.Telegram?.WebApp;
  // The SDK also creates WebApp in an ordinary browser, with empty initData.
  return app?.initData ? app : null;
}

/**
 * SDK загружается асинхронно: в обычном браузере (вне Telegram) страница
 * не должна зависеть от доступности telegram.org. Внутри Telegram объект
 * window.Telegram обычно уже внедрён клиентом или подхватится за долю секунды.
 */
function loadSdk(): Promise<TelegramWebApp | null> {
  if (typeof window === 'undefined') return Promise.resolve(null);
  const available = currentWebApp();
  if (available) return Promise.resolve(available);
  if (sdkPromise) return sdkPromise;
  sdkPromise = new Promise((resolve) => {
    const finish = (app: TelegramWebApp | null) => {
      clearInterval(poll);
      clearTimeout(timeout);
      script.removeEventListener('error', failed);
      resolve(app);
    };
    const poll = setInterval(() => {
      const app = currentWebApp();
      if (app) finish(app);
    }, POLL_INTERVAL_MS);
    const timeout = setTimeout(() => finish(null), POLL_TIMEOUT_MS);
    const existing = document.querySelector<HTMLScriptElement>('script[data-tg-sdk]');
    const script = existing ?? document.createElement('script');
    const failed = () => finish(null);
    script.addEventListener('error', failed, { once: true });
    if (!existing) {
      script.src = SDK_URL;
      script.async = true;
      script.dataset.tgSdk = 'true';
      document.head.appendChild(script);
    }
  });
  return sdkPromise;
}

function initWebApp(webApp: TelegramWebApp) {
  if (initialized.has(webApp)) return;
  initialized.add(webApp);
  try {
    webApp.ready();
    webApp.expand();
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
    currentWebApp,
  );

  useEffect(() => {
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
  }, []);

  const haptic = useCallback((style: 'light' | 'medium' | 'heavy') => {
    try { webApp?.HapticFeedback?.impactOccurred(style); } catch { /* Optional SDK feature. */ }
  }, [webApp]);

  const notify = useCallback((type: 'error' | 'success' | 'warning') => {
    try { webApp?.HapticFeedback?.notificationOccurred(type); } catch { /* Optional SDK feature. */ }
  }, [webApp]);

  return {
    webApp,
    isTelegram: webApp !== null,
    userName: webApp?.initDataUnsafe?.user?.first_name ?? null,
    haptic,
    notify,
  };
}
