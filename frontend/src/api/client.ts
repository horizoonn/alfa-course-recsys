import type {
  ApiErrorResponse,
  DemoProfile,
  MetaResponse,
  Profile,
  RecommendationsRequest,
  RecommendationsResponse,
} from '../types';
import { mockDemoProfiles, mockMeta, mockRecommendations } from './mock';

export const USE_MOCK = import.meta.env.VITE_USE_MOCK !== 'false';
const BASE = (import.meta.env.VITE_API_BASE ?? '/api').replace(/\/+$/, '');

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
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function requestJson<T>(url: string, init?: RequestInit): Promise<T> {
  let res: Response;
  try {
    res = await fetch(url, init);
  } catch {
    throw new ApiError('NETWORK_ERROR', 'Сервис недоступен. Проверьте соединение и повторите запрос.');
  }
  if (!res.ok) {
    let code = res.status === 503 ? 'DATA_UNAVAILABLE' : 'REQUEST_FAILED';
    let message = res.status === 503
      ? 'Сервис временно недоступен. Повторите запрос позже.'
      : 'Ошибка запроса. Повторите попытку позже.';
    try {
      const body = (await res.json()) as Partial<ApiErrorResponse>;
      if (typeof body.error?.code === 'string') code = body.error.code;
      if (typeof body.error?.message === 'string') message = body.error.message;
    } catch {}
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

export async function fetchRecommendations(
  profile: Profile,
  demoProfileId?: string,
): Promise<RecommendationsResponse> {
  if (USE_MOCK) return mockRecommendations(profile, demoProfileId);

  const body: RecommendationsRequest = demoProfileId ? { demo_profile_id: demoProfileId } : { profile };
  return requestJson<RecommendationsResponse>(`${BASE}/recommendations`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
}
