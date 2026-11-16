import type { MouseEvent } from 'react';
import { useTelegram } from '../hooks/useTelegram';
import { useEffect, useState } from 'react';
import { fetchDeepening, USE_MOCK } from '../api/client';
import type { DeepeningContext, DeepeningResponse } from '../types';

export default function Deepening({ materialKey, context = {} }: { materialKey: string; context?: DeepeningContext }) {
  const { webApp } = useTelegram();
  const openLink = (event: MouseEvent<HTMLAnchorElement>, url: string | null) => {
    if (!webApp || !url) return;
    event.preventDefault();
    try { webApp.openLink(url); } catch { window.open(url, "_blank", "noopener,noreferrer"); }
  };
  const [result, setResult] = useState<DeepeningResponse | null>(null);
  const [failed, setFailed] = useState(false);
  const [revision, setRevision] = useState(0);
  const { query, history_profile_id: historyProfileId, demo_profile_id: demoProfileId } = context;
  const excluded = JSON.stringify(context.exclude_keys ?? []);
  useEffect(() => {
    if (USE_MOCK) return;
    const controller = new AbortController();
    setResult(null);
    setFailed(false);
    fetchDeepening(materialKey, {
      ...(query ? { query } : {}),
      ...(historyProfileId ? { history_profile_id: historyProfileId } : {}),
      ...(demoProfileId ? { demo_profile_id: demoProfileId } : {}),
      exclude_keys: JSON.parse(excluded) as string[],
    }, controller.signal)
      .then((value) => { if (!controller.signal.aborted) setResult(value); })
      .catch(() => { if (!controller.signal.aborted) setFailed(true); });
    return () => controller.abort();
  }, [materialKey, query, historyProfileId, demoProfileId, excluded, revision]);
  if (failed) return <button className="text-button" onClick={() => setRevision((value) => value + 1)}>Повторить подбор углубления</button>;
  if (!result?.items.length) return null;
  return (
    <aside className="deepening" aria-label="Углубить тему">
      <h4>Углубить тему</h4>
      {result.items.map((item) => (
        <div className="deepening-item" key={item.material_key}>
          <span className="hint">{item.type === 'course' ? 'Курс' : 'Статья'} · {item.access === 'free' ? 'Бесплатно' : item.access === 'paid' ? 'Платно' : item.access === 'subscription' ? 'По подписке' : 'Условия на странице материала'}</span>
          <a href={item.source_url ?? undefined} target="_blank" rel="noopener noreferrer" onClick={(event) => openLink(event, item.source_url)}>{item.title}</a>
          {item.reasons.map((reason) => <p className="hint" key={reason.code}>{reason.text}</p>)}
        </div>
      ))}
    </aside>
  );
}
