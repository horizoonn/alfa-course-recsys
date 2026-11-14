export type HelpModeId = 'quick_answer' | 'tool' | 'deep_learning';

export interface Profile {
  stage?: string | null;
  industry?: string | null;
  task_id?: string | null;
  experience?: string | null;
  help_mode?: string | null;
}

export type AccessType = 'free' | 'subscription' | 'paid' | 'unknown';

export type MaterialKind =
  | 'article'
  | 'course'
  | 'document'
  | 'term'
  | 'calculator'
  | 'podcast'
  | 'webinar';

export type RecMode = 'personalized' | 'topic' | 'starter' | 'fallback';

export type ReasonCode =
  | 'task_match'
  | 'help_mode_match'
  | 'stage_match'
  | 'experience_match'
  | 'industry_match'
  | 'history_match'
  | 'general_selection'
  | 'broader_topic';

export interface RecommendationReason {
  code: ReasonCode;
  text: string;
}

export interface NextAction {
  type: 'open_material' | 'open_purchase' | 'refine_request';
  label: string;
}

export interface MaterialCard {
  material_key: string;
  title: string;
  type: MaterialKind;
  safe_summary: string | null;
  access: AccessType;
  source_url: string | null;
  reasons: RecommendationReason[];
  next_action: NextAction;
}

export interface Limitation {
  code:
    | 'HISTORY_UNAVAILABLE'
    | 'BROADER_TOPIC'
    | 'URL_UNVERIFIED'
    | 'NO_MATCHES'
    | 'PROFILE_INCOMPLETE';
  message: string;
}

export type RecommendationsRequest =
  | { profile?: Profile; demo_profile_id?: never; limit?: number }
  | { demo_profile_id: string; profile?: never; limit?: number };

export interface RecommendationsResponse {
  items: MaterialCard[];
  mode: RecMode;
  limitations: Limitation[];
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
  request_id: string;
  error: {
    code: 'INVALID_INPUT' | 'DATA_UNAVAILABLE' | 'INTERNAL_ERROR';
    message: string;
    fields: { field: string; message: string }[];
  };
}
