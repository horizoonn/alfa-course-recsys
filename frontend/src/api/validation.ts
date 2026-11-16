import type { DeepeningResponse, HistoryProfile, DemoProfile, MetaResponse, RecommendationsResponse } from '../types';

const record = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null && !Array.isArray(value);
const text = (value: unknown): value is string => typeof value === 'string';
const nullableText = (value: unknown) => value === null || text(value);

export function validMeta(value: unknown): value is MetaResponse {
  return record(value) && ['tasks', 'stages', 'industries', 'experience', 'help_modes'].every((key) =>
    Array.isArray(value[key]) && value[key].every((option) =>
      record(option) && text(option.id) && text(option.label)));
}

export function validDemoProfiles(value: unknown): value is DemoProfile[] {
  return Array.isArray(value) && value.every((profile) => record(profile) &&
    text(profile.id) && text(profile.title) && text(profile.description) && record(profile.profile) &&
    Object.values(profile.profile).every(nullableText));
}

export function validRecommendations(value: unknown): value is RecommendationsResponse {
  return record(value) && text(value.request_id) &&
    ['personalized', 'topic', 'starter', 'fallback'].includes(String(value.mode)) &&
    Array.isArray(value.limitations) && value.limitations.every((limitation) =>
      record(limitation) && text(limitation.code) && text(limitation.message)) &&
    Array.isArray(value.items) && value.items.length <= 3 &&
    value.items.every((item) => record(item) && text(item.material_key) && text(item.title) &&
      text(item.type) && text(item.access) && nullableText(item.safe_summary) &&
      nullableText(item.source_url) && Array.isArray(item.reasons) &&
      item.reasons.every((reason) => record(reason) && text(reason.code) && text(reason.text)) &&
      record(item.next_action) && text(item.next_action.label) &&
      ['open_material', 'open_purchase', 'refine_request'].includes(String(item.next_action.type)));
}

export function validHistoryProfiles(value: unknown): value is HistoryProfile[] {
  return Array.isArray(value) && value.every((item) => record(item) && text(item.id) && text(item.title) && text(item.description) &&
    ['completed_courses', 'learning_courses', 'positive_ratings', 'negative_ratings'].every((key) => typeof item[key] === 'number' && Number.isInteger(item[key]) && Number(item[key]) >= 0));
}
export function validDeepening(value: unknown): value is DeepeningResponse {
  return record(value) && text(value.request_id) && Array.isArray(value.items) && value.items.length <= 2 &&
    validRecommendations({ ...value, mode: 'topic', limitations: [] });
}
