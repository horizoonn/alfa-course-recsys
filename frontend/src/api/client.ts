import type {
  DemoProfile,
  MetaResponse,
  Profile,
  RecommendationsRequest,
  RecommendationsResponse,
} from '../types';
import { mockDemoProfiles, mockMeta, mockRecommendations } from './mock';
import { validDemoProfiles, validMeta, validRecommendations } from './validation';

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
  const timeout = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);
  try {
    let res: Response;
    try {
      res = await fetch(url, { ...init, signal: controller.signal });
    } catch {
      if (controller.signal.aborted) {
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
      if (controller.signal.aborted) {
        throw new ApiError('REQUEST_TIMEOUT', 'Сервис долго не отвечает. Попробуйте ещё раз.');
      }
      throw new ApiError('INVALID_RESPONSE', 'Сервис вернул непонятный ответ. Попробуйте ещё раз.');
    }
  } finally {
    clearTimeout(timeout);
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
): Promise<RecommendationsResponse> {
  if (USE_MOCK) return mockRecommendations(profile, demoProfileId);

  const body: RecommendationsRequest = demoProfileId ? { demo_profile_id: demoProfileId } : { profile };
  return requestJson<RecommendationsResponse>(`${BASE}/recommendations`, validRecommendations, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
}
