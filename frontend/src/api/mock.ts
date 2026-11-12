import type {
  DemoProfile,
  MaterialCard,
  MetaResponse,
  Profile,
  RecommendationsResponse,
} from '../types';

// Моки позволяют разрабатывать и демо на чек-пойнтах без восстановленной БД.
// Это СИНТЕТИЧЕСКИЕ данные, не реальные материалы каталога Альфа Курса.
// Все ID и названия — демонстрационные, их нужно убрать после подключения Go API.

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

// Базовый профиль P02, на который ссылаются P03, P04 и P10.
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

// Мок-каталог: task_id и контекстные поля имитируют реальные связи материала.
interface MockMaterial {
  material_key: string;
  title: string;
  type: string;
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
    material_key: 'articles_first-steps-101',
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
    material_key: 'webinars_first-clients-204',
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
    material_key: 'calculators_client-channel-301',
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
    material_key: 'courses_marketing-deep-401',
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
    material_key: 'articles_promotion-basics-501',
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
    material_key: 'courses_finance-system-601',
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
    material_key: 'lessons_accounting-glossary-701',
    title: 'Учёт для владельца торговли: базовые понятия',
    type: 'lesson',
    task_id: 'accounting',
    industry_id: 'retail',
    experience_id: 'beginner',
    help_mode_id: 'quick_answer',
    safe_summary: 'Короткий урок: выручка, прибыль, товарные остатки — что означает каждый термин.',
    access: 'free',
    source_url: 'https://example.com/materials/accounting-basics',
    next_action: 'Изучить урок',
  },
  {
    material_key: 'documents_growth-plan-801',
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
    material_key: 'podcasts_business-stories-901',
    title: 'Подкаст: истории владельцев малого бизнеса',
    type: 'podcast',
    task_id: 'promotion',
    safe_summary: 'Разговоры с предпринимателями о продвижении без бюджета и первых ошибках.',
    access: 'free',
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

/**
 * Упрощённая имитация правил из раздела 4 документа:
 * 4T + 2M + S + E + I, сортировка по T, затем по баллу, затем по ключу.
 * Реальная логика будет на Go — здесь мок только для разработки UI.
 */
function scoreMaterial(m: MockMaterial, profile: Profile): { t: number; total: number } {
  let t = 0;
  if (profile.task_id && m.task_id === profile.task_id) t = 2;
  const mm = profile.help_mode && m.help_mode_id === profile.help_mode ? 1 : 0;
  const s = profile.stage && m.stage_id === profile.stage ? 1 : 0;
  const e = profile.experience && m.experience_id === profile.experience ? 1 : 0;
  const i = profile.industry && m.industry_id === profile.industry ? 1 : 0;
  return { t, total: 4 * t + 2 * mm + s + e + i };
}

const delay = (ms: number) => new Promise((r) => setTimeout(r, ms));
let requestCounter = 0;

export async function mockRecommendations(
  profile: Profile,
  demoProfileId?: string,
): Promise<RecommendationsResponse> {
  await delay(500); // имитация сетевой задержки

  const profile_ = demoProfileId
    ? mockDemoProfiles.find((p) => p.id === demoProfileId)?.profile ?? profile
    : profile;

  const hasAny =
    !!profile_.task_id || !!profile_.stage || !!profile_.industry || !!profile_.experience;

  // P09: задача есть, но по мок-аудиту материалов нет — честное пустое состояние.
  if (profile_.task_id === 'export') {
    return {
      items: [],
      mode: 'fallback',
      limitations:
        'По задаче «Выход на зарубежные рынки» материалов пока нет. Попробуйте другую задачу или выберите более общую тему.',
      request_id: `mock-${++requestCounter}`,
    };
  }

  const scored = catalog
    .map((m) => ({ m, ...scoreMaterial(m, profile_) }))
    .filter((x) => x.t > 0)
    .sort(
      (a, b) =>
        b.t - a.t || b.total - a.total || a.m.material_key.localeCompare(b.m.material_key),
    )
    .slice(0, 3);

  // Холодный старт: пустой профиль → стартовая подборка.
  if (scored.length === 0 && !hasAny) {
    return {
      items: takeStarter().map((m) => toCard(m)),
      mode: 'starter',
      limitations:
        'Профиль пустой, поэтому показана общая стартовая подборка. Уточните задачу — станет точнее.',
      request_id: `mock-${++requestCounter}`,
    };
  }

  // Нет прямого совпадения — fallback на общую тему.
  if (scored.length === 0) {
    return {
      items: takeStarter().map((m) => toCard(m)),
      mode: 'fallback',
      limitations: `Точного совпадения по задаче «${taskLabel(profile_.task_id) ?? '—'}» нет, предлагаем более общую тему.`,
      request_id: `mock-${++requestCounter}`,
    };
  }

  const hasContext = !!profile_.stage || !!profile_.industry || !!profile_.experience;
  const mode: RecommendationsResponse['mode'] =
    profile_.task_id && hasContext ? 'personalized' : 'topic';

  const items = scored.map(({ m, t, total }) => toCard(m, reasonsFor(m, profile_, t, total)));

  return { items, mode, request_id: `mock-${++requestCounter}` };
}

function takeStarter(): MockMaterial[] {
  return [catalog[0], catalog[1], catalog[4]];
}

function reasonsFor(m: MockMaterial, profile: Profile, t: number, total: number): string[] {
  const reasons: string[] = [];
  const label = taskLabel(profile.task_id);
  if (t > 0 && label) reasons.push(`по вашей задаче «${label}»`);
  if (profile.help_mode && m.help_mode_id === profile.help_mode) {
    reasons.push(`подходит для выбранного формата: ${helpModeLabel(profile.help_mode)?.toLowerCase()}`);
  }
  if (profile.stage && m.stage_id === profile.stage) reasons.push('соответствует стадии бизнеса');
  if (profile.experience && m.experience_id === profile.experience) reasons.push('для вашего уровня в теме');
  if (profile.industry && m.industry_id === profile.industry) reasons.push('по вашей сфере бизнеса');
  if (reasons.length === 0) reasons.push(total > 0 ? 'связано с выбранной темой' : 'из общей подборки');
  return reasons;
}

function toCard(m: MockMaterial, reasons?: string[]): MaterialCard {
  return {
    material_key: m.material_key,
    title: m.title,
    type: m.type,
    safe_summary: m.safe_summary,
    access: m.access,
    source_url: m.source_url,
    reasons: reasons ?? ['из стартовой подборки'],
    next_action: m.next_action,
  };
}
