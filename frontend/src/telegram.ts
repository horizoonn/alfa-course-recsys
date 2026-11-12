// Минимальные объявления Telegram WebApp SDK.
// SDK подключается скриптом в index.html (https://telegram.org/js/telegram-web-app.js).
export interface TelegramWebAppUser {
  id?: number;
  first_name?: string;
  last_name?: string;
  username?: string;
  language_code?: string;
}

export interface TelegramHapticFeedback {
  impactOccurred(style: 'light' | 'medium' | 'heavy' | 'rigid' | 'soft'): void;
  notificationOccurred(type: 'error' | 'success' | 'warning'): void;
}

export interface TelegramWebApp {
  ready(): void;
  expand(): void;
  close(): void;
  initData: string;
  initDataUnsafe: { user?: TelegramWebAppUser };
  colorScheme: 'light' | 'dark';
  themeParams: Record<string, string>;
  setHeaderColor(color: string): void;
  setBackgroundColor(color: string): void;
  HapticFeedback?: TelegramHapticFeedback;
}

declare global {
  interface Window {
    Telegram?: { WebApp: TelegramWebApp };
  }
}
