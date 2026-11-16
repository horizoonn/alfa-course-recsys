import { useState } from 'react';
import type { ChangeEvent } from 'react';
import type { DemoProfile, HistoryProfile, MetaResponse, Profile } from '../types';
import { useTelegram } from '../hooks/useTelegram';

interface Props {
  meta: MetaResponse;
  demoProfiles: DemoProfile[];
  historyProfiles: HistoryProfile[];
  initialHistoryProfileId?: string;
  initialQuery?: string;
  initialProfile?: Profile;
  initialDemoProfileId?: string;
  loading: boolean;
  error: string | null;
  onSubmit: (profile: Profile, demoProfileId?: string, query?: string, historyProfileId?: string) => void;
}

const NOT_SET = '';

export default function ProfileForm({
  meta,
  demoProfiles,
  historyProfiles,
  initialHistoryProfileId,
  initialQuery,
  initialProfile,
  initialDemoProfileId,
  loading,
  error,
  onSubmit,
}: Props) {
  const { userName } = useTelegram();
  const initialDemo = demoProfiles.find((demo) => demo.id === initialDemoProfileId);
  const [profile, setProfile] = useState<Profile>(() => ({
    ...(initialDemo?.profile ?? initialProfile ?? {}),
  }));
  const [query, setQuery] = useState(initialQuery ?? "");
  const [historyProfileId, setHistoryProfileId] = useState(initialHistoryProfileId ?? NOT_SET);
  const [demoProfileId, setDemoProfileId] = useState(initialDemo?.id ?? NOT_SET);

  const update =
    (field: keyof Profile) =>
    (e: ChangeEvent<HTMLSelectElement>) => {
      const value = e.target.value || null;
      setDemoProfileId(NOT_SET);
      if (field === "task_id") setQuery("");
      setProfile((p) => ({ ...p, [field]: value }));
    };

  const handleDemoChange = (e: ChangeEvent<HTMLSelectElement>) => {
    setQuery("");
    setHistoryProfileId(NOT_SET);
    const demo = demoProfiles.find((item) => item.id === e.target.value);
    setDemoProfileId(demo?.id ?? NOT_SET);
    if (demo) setProfile({ ...demo.profile });
  };
  const submit = () => {
    if (demoProfileId) {
      onSubmit({}, demoProfileId);
    } else {
      onSubmit(profile, undefined, query, historyProfileId || undefined);
    }
  };

  const skipAll = () => {
    setQuery("");
    setHistoryProfileId(NOT_SET);
    setProfile({});
    setDemoProfileId(NOT_SET);
    onSubmit({});
  };

  const renderSelect = (
    field: keyof Profile,
    label: string,
    options: { id: string; label: string }[],
  ) => (
    <div className="field">
      <label className="field-label" htmlFor={`profile-${field}`}>{label}</label>
      <select id={`profile-${field}`} value={profile[field] ?? NOT_SET} onChange={update(field)} disabled={loading}>
        <option value={NOT_SET}>Не указано</option>
        {options.map((o) => (
          <option key={o.id} value={o.id}>
            {o.label}
          </option>
        ))}
      </select>
    </div>
  );

  return (
    <div className="screen">
      <header className="header">
        <div className="brand"><span className="brand-mark" aria-hidden="true">А</span><span>Альфа Курс</span><span className="brand-caption">для бизнеса</span></div>
        <p className="eyebrow">Знания в нужный момент</p>
        <h1>{userName ? `${userName}, что сейчас важно для бизнеса?` : 'Следующий шаг для вашего бизнеса'}</h1>
        <p className="subtitle">
          Короткий ответ, рабочий инструмент или подробный курс — найдём материалы под вашу задачу.
        </p>
      </header>

      <form onSubmit={(event) => { event.preventDefault(); submit(); }} aria-busy={loading}>
        <section className="section main-section">
          <div className="section-heading"><span className="step-number">01</span><h2 className="section-title">Что хотите решить?</h2></div>
          <div className="field">
            <label className="field-label" htmlFor="business-query">Опишите задачу своими словами</label>
            <textarea id="business-query" maxLength={1000} rows={4} value={query} disabled={loading} placeholder="Например: покупатели приходят один раз и больше не возвращаются" onChange={(event) => {
              setQuery(event.target.value); setDemoProfileId(NOT_SET); setProfile((value) => ({ ...value, task_id: null }));
            }} />
            <span className="hint">{query.length}/1000</span>
          </div>
          {renderSelect('task_id', 'Или выберите тему', meta.tasks)}
          <fieldset className="format-fieldset" disabled={loading}>
            <legend className="field-label">Как вам удобнее?</legend>
            <div className="format-options">
              {meta.help_modes.map((mode) => (
                <button
                  key={mode.id}
                  type="button"
                  className={`format-option${profile.help_mode === mode.id ? ' selected' : ''}`}
                  aria-label={mode.label}
                  aria-pressed={profile.help_mode === mode.id}
                  onClick={() => {
                    setDemoProfileId(NOT_SET);
                    setProfile((p) => ({ ...p, help_mode: p.help_mode === mode.id ? null : mode.id }));
                  }}
                >
                  <span aria-hidden="true" className="format-icon">{mode.id === 'quick_answer' ? '↗' : mode.id === 'tool' ? '⊞' : '≡'}</span>
                  {mode.id === 'tool' ? 'Инструмент' : mode.id === 'deep_learning' ? 'Обучение' : mode.label}
                </button>
              ))}
            </div>
          </fieldset>
        </section>

        <details className="section refinement">
          <summary><span><span className="step-number">02</span>Уточнить подбор</span><span className="optional-label">необязательно</span></summary>
          <p className="hint">Расскажите немного о себе. Подбор учтёт те признаки, которые есть у материалов.</p>
          {renderSelect('stage', 'Стадия бизнеса', meta.stages)}
          {renderSelect('industry', 'Сфера', meta.industries)}
          {renderSelect('experience', 'Опыт в выбранной теме', meta.experience)}
        </details>

        {error && <div className="error-banner" role="alert">{error}</div>}

        <div className="actions">
          <button type="submit" className="btn btn-primary" disabled={loading}>
            {loading ? 'Подбираем…' : <>Подобрать материалы <span aria-hidden="true">→</span></>}
          </button>
          <button type="button" className="btn btn-secondary" onClick={skipAll} disabled={loading}>
            Пропустить анкету
          </button>
        </div>
        <p className="privacy-note">Без регистрации. Можно пропустить любое поле.</p>
      </form>

      {historyProfiles.length > 0 && <details className="examples">
        <summary>История из базы Альфа-Курса</summary>
        <p className="hint">Обезличенные сценарии реального обучения для проверки. Это не история вашего аккаунта.</p>
        <div className="field"><label className="field-label" htmlFor="history-profile">Сценарий истории</label>
          <select id="history-profile" value={historyProfileId} disabled={loading} onChange={(event) => { setHistoryProfileId(event.target.value); setDemoProfileId(NOT_SET); }}>
            <option value={NOT_SET}>Без сценария из базы</option>
            {historyProfiles.map((item) => <option key={item.id} value={item.id}>{item.id} — {item.title}</option>)}
          </select>
        </div>
        {historyProfileId && <p className="hint">{historyProfiles.find((item) => item.id === historyProfileId)?.description}</p>}
      </details>}
      {demoProfiles.length > 0 && (
        <details className="examples" open={demoProfileId ? true : undefined}>
          <summary>Попробовать на примере</summary>
          <p className="hint">Готовые ситуации помогут познакомиться с подбором.</p>
          <div className="field">
            <label className="field-label" htmlFor="demo-profile">Пример ситуации</label>
            <select id="demo-profile" value={demoProfileId} onChange={handleDemoChange} disabled={loading}>
              <option value={NOT_SET}>Свой ввод</option>
              {demoProfiles.map((p) => <option key={p.id} value={p.id}>{p.title}</option>)}
            </select>
          </div>
          {demoProfileId && <p className="hint">{demoProfiles.find((p) => p.id === demoProfileId)?.description}</p>}
        </details>
      )}
      <footer className="page-footer">Меньше искать. Больше делать.</footer>
    </div>
  );
}
