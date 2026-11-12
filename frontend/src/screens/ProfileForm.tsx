import { useState } from 'react';
import type { ChangeEvent } from 'react';
import type { DemoProfile, MetaResponse, Profile } from '../types';
import { USE_MOCK } from '../api/client';
import { useTelegram } from '../hooks/useTelegram';

interface Props {
  meta: MetaResponse;
  demoProfiles: DemoProfile[];
  loading: boolean;
  error: string | null;
  onSubmit: (profile: Profile, demoProfileId?: string) => void;
}

const NOT_SET = '';

export default function ProfileForm({ meta, demoProfiles, loading, error, onSubmit }: Props) {
  const { userName } = useTelegram();
  const [profile, setProfile] = useState<Profile>({});
  const [demoProfileId, setDemoProfileId] = useState<string>(NOT_SET);

  const update =
    (field: keyof Profile) =>
    (e: ChangeEvent<HTMLSelectElement>) => {
      const value = e.target.value || null;
      setProfile((p) => ({ ...p, [field]: value }));
    };

  const handleDemoChange = (e: ChangeEvent<HTMLSelectElement>) => {
    setDemoProfileId(e.target.value);
  };

  // profile и demo_profile_id одновременно не передаются (контракт API).
  const submit = () => {
    if (demoProfileId) {
      onSubmit({}, demoProfileId);
    } else {
      onSubmit(profile);
    }
  };

  const skipAll = () => {
    // R3: холодный старт — можно пропустить все поля.
    setProfile({});
    setDemoProfileId(NOT_SET);
    onSubmit({});
  };

  const renderSelect = (
    field: keyof Profile,
    label: string,
    options: { id: string; label: string }[],
  ) => (
    <label className="field">
      <span className="field-label">{label}</span>
      <select value={profile[field] ?? NOT_SET} onChange={update(field)}>
        <option value={NOT_SET}>Не указано</option>
        {options.map((o) => (
          <option key={o.id} value={o.id}>
            {o.label}
          </option>
        ))}
      </select>
    </label>
  );

  return (
    <div className="screen">
      <header className="header">
        <h1>Альфа Курс</h1>
        <p className="subtitle">
          {userName ? `${userName}, подберём материал` : 'Подберём материал'} под вашу текущую
          задачу и подскажем следующий шаг.
        </p>
        {USE_MOCK && (
          <div className="mock-badge" title="Фронт работает без бэкенда на синтетических данных">
            Демо: мок-данные
          </div>
        )}
      </header>

      {demoProfiles.length > 0 && (
        <section className="section">
          <h2 className="section-title">Проверка на демо-профилях</h2>
          <label className="field">
            <span className="field-label">Профиль P01–P10</span>
            <select value={demoProfileId} onChange={handleDemoChange}>
              <option value={NOT_SET}>Свой ввод</option>
              {demoProfiles.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.title}
                </option>
              ))}
            </select>
          </label>
          {demoProfileId && (
            <p className="hint">
              {demoProfiles.find((p) => p.id === demoProfileId)?.description}
            </p>
          )}
        </section>
      )}

      <section className="section">
        <h2 className="section-title">Ваша задача</h2>
        <p className="hint">Все поля необязательные — можно пропустить любое или все.</p>
        {renderSelect('task_id', 'Задача', meta.tasks)}
        {renderSelect('stage', 'Стадия бизнеса', meta.stages)}
        {renderSelect('industry', 'Сфера', meta.industries)}
        {renderSelect('experience', 'Опыт в выбранной теме', meta.experience)}
        {renderSelect('help_mode', 'Формат помощи', meta.help_modes)}
      </section>

      {error && <div className="error-banner">{error}</div>}

      <div className="actions">
        <button className="btn btn-primary" onClick={submit} disabled={loading}>
          {loading ? 'Подбираем…' : 'Подобрать материалы'}
        </button>
        <button className="btn btn-secondary" onClick={skipAll} disabled={loading}>
          Пропустить анкету
        </button>
      </div>
    </div>
  );
}
