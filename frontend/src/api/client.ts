import type {
  DeepeningContext,
  DeepeningResponse,
  HistoryProfile,
  DemoProfile,
  MetaResponse,
  Profile,
  RecommendationsRequest,
  RecommendationsResponse,
} from '../types';
import { mockDemoProfiles, mockMeta, mockRecommendations } from './mock';
import { validDemoProfiles, validMeta, validRecommendations, validHistoryProfiles, validDeepening } from './validation';

export const USE_MOCK = import.meta.env.VITE_USE_MOCK === 'true';
const BASE = (import.meta.env.VITE_API_BASE ?? '/api').replace(/\/+$/, '');
export const REQUEST_TIMEOUT_MS = 25_000;

export class ApiError extends Error {
  constructor(
    public code: string,
    message: string,
    public status?: number,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

export function isUnavailable(error: unknown): error is ApiError {
  return error instanceof ApiError && (
    ['NETWORK_ERROR', 'REQUEST_TIMEOUT', 'DATA_UNAVAILABLE'].includes(error.code) ||
    [502, 503, 504].includes(error.status ?? 0)
  );
}

function delay(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function requestJson<T>(url: string, validate: (value: unknown) => value is T, init?: RequestInit): Promise<T> {
  const controller = new AbortController();
  const caller = init?.signal;
  const cancel = () => controller.abort();
  if (caller?.aborted) cancel();
  caller?.addEventListener('abort', cancel, { once: true });
  let timedOut = false;
  const timeout = setTimeout(() => { timedOut = true; controller.abort(); }, REQUEST_TIMEOUT_MS);
  try {
    let res: Response;
    try {
      res = await fetch(url, { ...init, signal: controller.signal });
    } catch {
      if (caller?.aborted) throw new DOMException('Request cancelled', 'AbortError');
      if (timedOut) {
        throw new ApiError('REQUEST_TIMEOUT', 'Сервис долго не отвечает. Попробуйте ещё раз.');
      }
      throw new ApiError('NETWORK_ERROR', 'Сервис недоступен. Проверьте соединение и повторите запрос.');
    }
    if (!res.ok) {
      const unavailable = [502, 503, 504].includes(res.status);
      let code = unavailable ? 'DATA_UNAVAILABLE' : 'REQUEST_FAILED';
      let message = unavailable
        ? 'Сервис временно недоступен. Повторите запрос позже.'
        : 'Ошибка запроса. Повторите попытку позже.';
      try {
        const body = (await res.json()) as { error?: { code?: unknown; message?: unknown } | string };
        if (body.error && typeof body.error === 'object') {
          if (typeof body.error.code === 'string') code = body.error.code;
          if (typeof body.error.message === 'string') message = body.error.message;
        }
      } catch { /* Proxies may return HTML instead of the API error envelope. */ }
      throw new ApiError(code, message, res.status);
    }
    try {
      const value: unknown = await res.json();
      if (!validate(value)) throw new Error('Invalid response shape');
      return value;
    } catch {
      if (caller?.aborted) throw new DOMException('Request cancelled', 'AbortError');
      if (timedOut) {
        throw new ApiError('REQUEST_TIMEOUT', 'Сервис долго не отвечает. Попробуйте ещё раз.');
      }
      throw new ApiError('INVALID_RESPONSE', 'Сервис вернул непонятный ответ. Попробуйте ещё раз.');
    }
  } finally {
    clearTimeout(timeout);
    caller?.removeEventListener('abort', cancel);
  }
}

export async function fetchMeta(): Promise<MetaResponse> {
  if (USE_MOCK) {
    await delay(200);
    return mockMeta;
  }
  return requestJson<MetaResponse>(`${BASE}/meta`, validMeta);
}

export async function fetchDemoProfiles(): Promise<DemoProfile[]> {
  if (USE_MOCK) {
    await delay(200);
    return mockDemoProfiles;
  }
  return requestJson<DemoProfile[]>(`${BASE}/demo-profiles`, validDemoProfiles);
}

export async function fetchRecommendations(
  profile: Profile,
  demoProfileId?: string,
  query?: string,
  signal?: AbortSignal,
  historyProfileId?: string,
): Promise<RecommendationsResponse> {
  const text = query?.trim();
  if (demoProfileId && (text || historyProfileId)) throw new ApiError('INVALID_INPUT', 'Выберите демо-профиль или свой запрос.');
  if (USE_MOCK) {
    if (text) throw new ApiError('MOCK_QUERY_UNAVAILABLE', 'Для поиска по тексту запустите backend.');
    return mockRecommendations(profile, demoProfileId);
  }

  const body: RecommendationsRequest = demoProfileId ? { demo_profile_id: demoProfileId } : { profile, ...(text ? { query: text } : {}), ...(historyProfileId ? { history_profile_id: historyProfileId } : {}) };
  return requestJson<RecommendationsResponse>(`${BASE}/recommendations`, validRecommendations, {
    method: 'POST',
    signal,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
}

export async function fetchHistoryProfiles(): Promise<HistoryProfile[]> {
  if (USE_MOCK) return [];
  return requestJson<HistoryProfile[]>(`${BASE}/history-profiles`, validHistoryProfiles);
}

export async function fetchDeepening(materialKey: string, context: DeepeningContext, signal?: AbortSignal): Promise<DeepeningResponse> {
  if (USE_MOCK) throw new ApiError('MOCK_DEEPENING_UNAVAILABLE', 'Для подбора углубления запустите backend.');
  return requestJson<DeepeningResponse>(`${BASE}/deepening`, validDeepening, {
    method: 'POST',
    signal,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ material_key: materialKey, ...context }),
  });
}
