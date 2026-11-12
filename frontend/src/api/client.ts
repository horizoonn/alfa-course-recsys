import type {
  DemoProfile,
  MetaResponse,
  Profile,
  RecommendationsResponse,
} from '../types';
import { mockDemoProfiles, mockMeta, mockRecommendations } from './mock';

export const USE_MOCK = import.meta.env.VITE_USE_MOCK !== 'false';
const BASE = import.meta.env.VITE_API_BASE ?? '/api';

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

function delay(ms: number) {
  return new Promise((r) => setTimeout(r, ms));
}

async function requestJson<T>(url: string, init?: RequestInit): Promise<T> {
  let res: Response;
  try {
    res = await fetch(url, init);
  } catch {
    // Сетевой сбой не маскируем под «нет совпадений».
    throw new ApiError('NETWORK_ERROR', 'Сервис недоступен. Проверьте соединение и повторите запрос.');
  }
  if (!res.ok) {
    let code = 'REQUEST_FAILED';
    let message = 'Ошибка запроса. Повторите попытку позже.';
    try {
      const body = (await res.json()) as { code?: string; message?: string };
      if (body.code) code = body.code;
      if (body.message) message = body.message;
    } catch {
      // тело не JSON — оставляем дефолтные тексты
    }
    if (res.status === 503) {
      code = code === 'REQUEST_FAILED' ? 'DATA_UNAVAILABLE' : code;
      message = 'База данных временно недоступна. Повторите запрос позже.';
    }
    throw new ApiError(code, message, res.status);
  }
  return (await res.json()) as T;
}

export async function fetchMeta(): Promise<MetaResponse> {
  if (USE_MOCK) {
    await delay(200);
    return mockMeta;
  }
  return requestJson<MetaResponse>(`${BASE}/meta`);
}

export async function fetchDemoProfiles(): Promise<DemoProfile[]> {
  if (USE_MOCK) {
    await delay(200);
    return mockDemoProfiles;
  }
  return requestJson<DemoProfile[]>(`${BASE}/demo-profiles`);
}

/**
 * profile и demo_profile_id одновременно не передаются — так требует контракт
 * (POST /api/recommendations). Отправляем только одно из двух.
 */
export async function fetchRecommendations(
  profile: Profile,
  demoProfileId?: string,
): Promise<RecommendationsResponse> {
  if (USE_MOCK) return mockRecommendations(profile, demoProfileId);

  const body: Record<string, unknown> = demoProfileId ? { demo_profile_id: demoProfileId } : { profile };
  return requestJson<RecommendationsResponse>(`${BASE}/recommendations`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
}
