import type {
  DemoProfile,
  MaterialCard,
  MetaResponse,
  Profile,
  RecommendationReason,
  RecommendationsResponse,
} from '../types';

export const mockMeta: MetaResponse = {
  tasks: [
    { id: 'first_steps', label: 'Первые шаги в бизнесе' },
    { id: 'attract_clients', label: 'Привлечь клиентов' },
    { id: 'promotion', label: 'Продвижение бизнеса' },
    { id: 'finance', label: 'Управление финансами' },
    { id: 'accounting', label: 'Разобраться в учёте' },
    { id: 'sales', label: 'Увеличить продажи' },
    { id: 'team', label: 'Найм и команда' },
    { id: 'export', label: 'Выход на зарубежные рынки' },
  ],
  stages: [
    { id: 'idea', label: 'Только идея' },
    { id: 'starting', label: 'Планирую открыть' },
    { id: 'running', label: 'Бизнес уже работает' },
    { id: 'growing', label: 'Развиваю и расту' },
  ],
  industries: [
    { id: 'services', label: 'Услуги' },
    { id: 'retail', label: 'Торговля' },
    { id: 'it', label: 'IT и онлайн' },
    { id: 'food', label: 'Общепит' },
    { id: 'production', label: 'Производство' },
  ],
  experience: [
    { id: 'beginner', label: 'Начальный' },
    { id: 'intermediate', label: 'Средний' },
    { id: 'advanced', label: 'Продвинутый' },
  ],
  help_modes: [
    { id: 'quick_answer', label: 'Быстрый ответ' },
    { id: 'tool', label: 'Практический инструмент' },
    { id: 'deep_learning', label: 'Подробное обучение' },
  ],
};
const p02Profile: Profile = {
  stage: 'running',
  industry: 'services',
  experience: 'beginner',
  task_id: 'attract_clients',
  help_mode: 'quick_answer',
};

export const mockDemoProfiles: DemoProfile[] = [
  {
    id: 'P01',
    title: 'P01 — начинающий',
    description:
      'Начинающий; планирует открыть дело в сфере услуг; опыт начальный; задача — первые шаги; помощь — быстрый ответ.',
    profile: {
      stage: 'starting',
      industry: 'services',
      experience: 'beginner',
      task_id: 'first_steps',
      help_mode: 'quick_answer',
    },
  },
  {
    id: 'P02',
    title: 'P02 — микробизнес, маркетинг',
    description:
      'Микробизнес в услугах уже работает; опыт в маркетинге начальный; задача — привлечь клиентов; помощь — быстрый ответ.',
    profile: p02Profile,
  },
  {
    id: 'P03',
    title: 'P03 — как P02, но инструмент',
    description: 'Все данные P02; помощь — практический инструмент.',
    profile: { ...p02Profile, help_mode: 'tool' },
  },
  {
    id: 'P04',
    title: 'P04 — как P02, но обучение',
    description: 'Все данные P02; помощь — подробное обучение.',
    profile: { ...p02Profile, help_mode: 'deep_learning' },
  },
  {
    id: 'P05',
    title: 'P05 — соло-фаундер, финансы',
    description:
      'Соло-фаундер; развитие бизнеса в услугах; опыт в теме средний; задача — управление финансами; помощь — подробное обучение.',
    profile: {
      stage: 'growing',
      industry: 'services',
      experience: 'intermediate',
      task_id: 'finance',
      help_mode: 'deep_learning',
    },
  },
  {
    id: 'P06',
    title: 'P06 — торговля, учёт',
    description:
      'Владелец небольшой торговли; бизнес работает; опыт в учёте начальный; задача — разобраться в учёте; помощь — быстрый ответ.',
    profile: {
      stage: 'running',
      industry: 'retail',
      experience: 'beginner',
      task_id: 'accounting',
      help_mode: 'quick_answer',
    },
  },
  {
    id: 'P07',
    title: 'P07 — пустой профиль',
    description: 'Профиль полностью пустой; история отсутствует. Ожидается стартовая подборка.',
    profile: {},
  },
  {
    id: 'P08',
    title: 'P08 — только задача',
    description:
      'Известна только задача — продвижение бизнеса. Ожидается тематическая выдача без выдуманных сведений.',
    profile: { task_id: 'promotion' },
  },
  {
    id: 'P09',
    title: 'P09 — нет подходящих материалов',
    description:
      'Выбрана задача, для которой после аудита подтверждено отсутствие прямого материала. Ожидается честное пустое состояние.',
    profile: { task_id: 'export' },
  },
  {
    id: 'P10',
    title: 'P10 — P02 + история',
    description:
      'Профиль P02 плюс синтетические завершение материала A и положительная оценка B (в моке история не моделируется).',
    profile: p02Profile,
  },
];
interface MockMaterial {
  material_key: string;
  title: string;
  type: MaterialCard['type'];
  safe_summary: string;
  access: MaterialCard['access'];
  source_url: string | null;
  next_action: string;
  task_id: string;
  industry_id?: string;
  stage_id?: string;
  experience_id?: string;
  help_mode_id?: string;
}

const catalog: MockMaterial[] = [
  {
    material_key: 'article:00000000-0000-4000-8000-000000000101',
    title: 'С чего начать своё дело: чек-лист первых шагов',
    type: 'article',
    task_id: 'first_steps',
    stage_id: 'starting',
    experience_id: 'beginner',
    help_mode_id: 'quick_answer',
    safe_summary:
      'Короткий чек-лист: выбор формы, регистрация, первые документы и первые деньги. Понятный язык без терминов.',
    access: 'free',
    source_url: 'https://example.com/materials/first-steps',
    next_action: 'Читать статью',
  },
  {
    material_key: 'webinar:00000000-0000-4000-8000-000000000204',
    title: 'Первые клиенты: где искать и как договариваться',
    type: 'webinar',
    task_id: 'attract_clients',
    industry_id: 'services',
    stage_id: 'running',
    help_mode_id: 'quick_answer',
    safe_summary: 'Запись вебинара: три рабочих канала привлечения для малого бизнеса в услугах.',
    access: 'free',
    source_url: 'https://example.com/materials/first-clients',
    next_action: 'Смотреть запись',
  },
  {
    material_key: 'calculator:00000000-0000-4000-8000-000000000301',
    title: 'Шаблон: подбор канала привлечения клиентов',
    type: 'calculator',
    task_id: 'attract_clients',
    industry_id: 'services',
    experience_id: 'beginner',
    help_mode_id: 'tool',
    safe_summary: 'Заполняете поля о бизнесе — получаете приоритизированный список каналов привлечения.',
    access: 'free',
    source_url: 'https://example.com/materials/channel-tool',
    next_action: 'Открыть инструмент',
  },
  {
    material_key: 'course:00000000-0000-4000-8000-000000000401',
    title: 'Курс: маркетинг для малого бизнеса',
    type: 'course',
    task_id: 'attract_clients',
    industry_id: 'services',
    experience_id: 'beginner',
    help_mode_id: 'deep_learning',
    safe_summary:
      'Углублённый курс: позиционирование, каналы, воронка. Понадобится несколько часов в неделю.',
    access: 'paid',
    source_url: 'https://example.com/materials/marketing-course',
    next_action: 'Изучить курс',
  },
  {
    material_key: 'article:00000000-0000-4000-8000-000000000501',
    title: 'Продвижение бизнеса: базовые каналы простыми словами',
    type: 'article',
    task_id: 'promotion',
    experience_id: 'beginner',
    help_mode_id: 'quick_answer',
    safe_summary: 'Обзор каналов продвижения без бюджета: соцсети, карты, сарафан.',
    access: 'free',
    source_url: 'https://example.com/materials/promotion-basics',
    next_action: 'Читать статью',
  },
  {
    material_key: 'course:00000000-0000-4000-8000-000000000601',
    title: 'Курс: финансы развивающегося бизнеса',
    type: 'course',
    task_id: 'finance',
    industry_id: 'services',
    stage_id: 'growing',
    experience_id: 'intermediate',
    help_mode_id: 'deep_learning',
    safe_summary: 'Системный курс: денежный поток, юнит-экономика, планирование роста.',
    access: 'paid',
    source_url: 'https://example.com/materials/finance-course',
    next_action: 'Изучить курс',
  },
  {
    material_key: 'term:00000000-0000-4000-8000-000000000701',
    title: 'Базовые понятия учёта для владельца торговли',
    type: 'term',
    task_id: 'accounting',
    industry_id: 'retail',
    experience_id: 'beginner',
    help_mode_id: 'quick_answer',
    safe_summary: 'Кратко о выручке, прибыли и товарных остатках: что означает каждый термин.',
    access: 'free',
    source_url: 'https://example.com/materials/accounting-basics',
    next_action: 'Прочитать определения',
  },
  {
    material_key: 'document:00000000-0000-4000-8000-000000000801',
    title: 'Шаблон: план развития на квартал',
    type: 'document',
    task_id: 'finance',
    stage_id: 'growing',
    help_mode_id: 'tool',
    safe_summary: 'Готовая таблица для планирования целей и показателей на квартал.',
    access: 'free',
    source_url: 'https://example.com/materials/growth-plan',
    next_action: 'Открыть шаблон',
  },
  {
    material_key: 'podcast:00000000-0000-4000-8000-000000000901',
    title: 'Подкаст: истории владельцев малого бизнеса',
    type: 'podcast',
    task_id: 'promotion',
    safe_summary: 'Разговоры с предпринимателями о продвижении без бюджета и первых ошибках.',
    access: 'subscription',
    source_url: null,
    next_action: 'Открыть материал',
  },
];

function taskLabel(taskId: string | null | undefined): string | null {
  return mockMeta.tasks.find((t) => t.id === taskId)?.label ?? null;
}

function helpModeLabel(id: string | null | undefined): string | null {
  return mockMeta.help_modes.find((m) => m.id === id)?.label ?? null;
}

function scoreMaterial(m: MockMaterial, profile: Profile): { t: number; total: number } {
  let t = 0;
  if (profile.task_id && m.task_id === profile.task_id) t = 2;
  const mm = profile.help_mode && m.help_mode_id === profile.help_mode ? 1 : 0;
  const s = profile.stage && m.stage_id === profile.stage ? 1 : 0;
  const e = profile.experience && m.experience_id === profile.experience ? 1 : 0;
  const i = profile.industry && m.industry_id === profile.industry ? 1 : 0;
  return { t, total: 4 * t + 2 * mm + s + e + i };
}

const delay = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

function createMockRequestId(): string {
  const bytes = crypto.getRandomValues(new Uint8Array(16));
  bytes[6] = (bytes[6] & 0x0f) | 0x40;
  bytes[8] = (bytes[8] & 0x3f) | 0x80;
  const hex = Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('');
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}

export async function mockRecommendations(
  profile: Profile,
  demoProfileId?: string,
): Promise<RecommendationsResponse> {
  await delay(500);

  const selected = demoProfileId
    ? mockDemoProfiles.find((item) => item.id === demoProfileId)?.profile ?? profile
    : profile;
  const request_id = createMockRequestId();
  const historyLimitations: RecommendationsResponse['limitations'] = demoProfileId === 'P10'
    ? [{ code: 'HISTORY_UNAVAILABLE', message: 'В этом демо история обучения ещё не моделируется.' }]
    : [];
  const ranked = catalog
    .map((material) => ({ material, ...scoreMaterial(material, selected) }))
    .sort((a, b) =>
      b.t - a.t || b.total - a.total ||
      a.material.material_key.localeCompare(b.material.material_key),
    );
  const direct = ranked.filter((item) => item.t > 0).slice(0, 3);

  if (selected.task_id && direct.length === 0) {
    return {
      items: [],
      mode: 'fallback',
      limitations: [
        ...historyLimitations,
        { code: 'NO_MATCHES', message: `В демонстрационном каталоге нет совпадений по задаче «${taskLabel(selected.task_id) ?? selected.task_id}». Измените запрос.` },
      ],
      request_id,
    };
  }

  let items: MaterialCard[];
  let mode: RecommendationsResponse['mode'];
  const limitations = [...historyLimitations];

  if (!Object.values(selected).some(Boolean)) {
    items = takeStarter().map((material) => toCard(material));
    mode = 'starter';
    limitations.push({
      code: 'PROFILE_INCOMPLETE',
      message: 'Профиль пустой, поэтому показана общая стартовая подборка. Уточните задачу — станет точнее.',
    });
  } else if (!selected.task_id) {
    const contextual = ranked.filter((item) => item.total > 0).slice(0, 3);
    items = contextual.length > 0
      ? contextual.map(({ material, t, total }) => toCard(material, reasonsFor(material, selected, t, total)))
      : takeStarter().map((material) => toCard(material));
    mode = 'fallback';
    limitations.push({
      code: 'PROFILE_INCOMPLETE',
      message: 'Задача не указана. Подборка учитывает доступные сведения профиля; уточните задачу для более точного результата.',
    });
  } else {
    items = direct.map(({ material, t, total }) => toCard(material, reasonsFor(material, selected, t, total)));
    mode = direct.some((item) => item.total > 4 * item.t) ? 'personalized' : 'topic';
  }

  if (items.some((item) => item.source_url === null)) {
    limitations.push({ code: 'URL_UNVERIFIED', message: 'Для части материалов ссылка неизвестна. Можно изменить подборку.' });
  }
  return { items, mode, limitations, request_id };
}

function takeStarter(): MockMaterial[] {
  return [catalog[0], catalog[1], catalog[4]];
}

function reasonsFor(
  material: MockMaterial,
  profile: Profile,
  taskMatch: number,
  total: number,
): RecommendationReason[] {
  const reasons: RecommendationReason[] = [];
  const label = taskLabel(profile.task_id);
  if (taskMatch > 0 && label) reasons.push({ code: 'task_match', text: `По вашей задаче «${label}»` });
  if (profile.help_mode && material.help_mode_id === profile.help_mode) {
    reasons.push({ code: 'help_mode_match', text: `Подходит для выбранного формата: ${helpModeLabel(profile.help_mode)?.toLowerCase()}` });
  }
  if (profile.stage && material.stage_id === profile.stage) reasons.push({ code: 'stage_match', text: 'Соответствует стадии бизнеса' });
  if (profile.experience && material.experience_id === profile.experience) reasons.push({ code: 'experience_match', text: 'Для вашего уровня в теме' });
  if (profile.industry && material.industry_id === profile.industry) reasons.push({ code: 'industry_match', text: 'По вашей сфере бизнеса' });
  if (reasons.length === 0) reasons.push({ code: 'general_selection', text: total > 0 ? 'Связано с выбранным контекстом' : 'Из общей подборки' });
  return reasons;
}

function toCard(
  material: MockMaterial,
  reasons: RecommendationReason[] = [{ code: 'general_selection', text: 'Из стартовой подборки' }],
): MaterialCard {
  return {
    material_key: material.material_key,
    title: material.title,
    type: material.type,
    safe_summary: material.safe_summary,
    access: material.access,
    source_url: material.source_url,
    reasons,
    next_action: material.source_url
      ? { type: 'open_material', label: material.next_action }
      : { type: 'refine_request', label: 'Изменить подборку' },
  };
}
