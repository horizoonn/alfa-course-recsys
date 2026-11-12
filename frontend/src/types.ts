// Контракты соответствуют документу Alfa_Course_MVP_Developers_v0_2.docx, раздел 5.
// Названия полей — это поля нового API, а не исходной БД.

export type HelpModeId = 'quick_answer' | 'tool' | 'deep_learning';

export interface Profile {
  /** Неизвестное значение передаётся как null, а не как догадка. */
  stage?: string | null;
  industry?: string | null;
  task_id?: string | null;
  experience?: string | null;
  help_mode?: string | null;
}

export type AccessType = 'free' | 'paid' | 'unknown';

export type RecMode = 'personalized' | 'topic' | 'starter' | 'fallback';

export interface MaterialCard {
  material_key: string;
  title: string;
  type: string;
  safe_summary: string;
  access: AccessType;
  /** null, если предоставленный URL отсутствует — адрес не выдумываем. */
  source_url: string | null;
  /** Причины формируются только из реально сработавших признаков. */
  reasons: string[];
  next_action: string;
}

export interface RecommendationsResponse {
  items: MaterialCard[];
  mode: RecMode;
  limitations?: string;
  request_id: string;
}

export interface MetaOption {
  id: string;
  label: string;
}

export interface MetaResponse {
  tasks: MetaOption[];
  stages: MetaOption[];
  industries: MetaOption[];
  experience: MetaOption[];
  help_modes: MetaOption[];
}

export interface DemoProfile {
  id: string;
  title: string;
  description: string;
  profile: Profile;
}

export interface ApiErrorResponse {
  code: string;
  message: string;
}
