import { useCallback, useEffect, useRef, useState } from 'react';
import ProfileForm from './screens/ProfileForm';
import Results from './screens/Results';
import { ErrorState, LoadingState } from './components/StatusStates';
import { ApiError, USE_MOCK, fetchDemoProfiles, fetchHistoryProfiles, fetchMeta, fetchRecommendations } from './api/client';
import type { DemoProfile, HistoryProfile, MetaResponse, Profile, RecommendationsResponse } from './types';
import { useTelegram } from './hooks/useTelegram';

type View = 'form' | 'results';

interface LastRequest {
  profile: Profile;
  demoProfileId?: string;
  query?: string;
  historyProfileId?: string;
}

export default function App() {
  const { webApp, haptic, notify } = useTelegram();
  const activeRequest = useRef<AbortController | null>(null);
  useEffect(() => () => activeRequest.current?.abort(), []);
  const [meta, setMeta] = useState<MetaResponse | null>(null);
  const [historyProfiles, setHistoryProfiles] = useState<HistoryProfile[]>([]);
  const [demoProfiles, setDemoProfiles] = useState<DemoProfile[]>([]);
  const [bootError, setBootError] = useState<string | null>(null);
  const [view, setView] = useState<View>('form');
  const [submitting, setSubmitting] = useState(false);
  const [result, setResult] = useState<RecommendationsResponse | null>(null);
  const [requestError, setRequestError] = useState<string | null>(null);
  const [lastRequest, setLastRequest] = useState<LastRequest | null>(null);
  useEffect(() => {
    let cancelled = false;
    Promise.all([fetchMeta(), fetchDemoProfiles().catch(() => []), fetchHistoryProfiles().catch(() => [])])
      .then(([m, dp, hp]) => {
        if (cancelled) return;
        setMeta(m);
        setDemoProfiles(dp);
        setHistoryProfiles(hp);
      })
      .catch((e: unknown) => {
        if (!cancelled) {
          setBootError(
            e instanceof ApiError ? e.message : 'Не удалось загрузить справочники.',
          );
        }
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const submit = useCallback(
    async (profile: Profile, demoProfileId?: string, query?: string, historyProfileId?: string) => {
      activeRequest.current?.abort();
      const controller = new AbortController();
      activeRequest.current = controller;
      haptic('medium');
      setSubmitting(true);
      setRequestError(null);
      setLastRequest({ profile, demoProfileId, query, historyProfileId });
      try {
        const res = await fetchRecommendations(profile, demoProfileId, query, controller.signal, historyProfileId);
        if (controller.signal.aborted || activeRequest.current !== controller) return;
        setResult(res);
        setView('results');
        notify(res.items.length === 0 ? 'warning' : 'success');
      } catch (e: unknown) {
        if (controller.signal.aborted || activeRequest.current !== controller) return;
        setRequestError(
          e instanceof ApiError ? e.message : 'Не удалось получить подборку.',
        );
        notify('error');
      } finally {
        if (activeRequest.current === controller) setSubmitting(false);
      }
    },
    [haptic, notify],
  );


  const retry = useCallback(() => {
    if (lastRequest) void submit(lastRequest.profile, lastRequest.demoProfileId, lastRequest.query, lastRequest.historyProfileId);
  }, [lastRequest, submit]);

  const backToForm = useCallback(() => {
    activeRequest.current?.abort();
    setSubmitting(false);
    setRequestError(null);
    haptic('light');
    setView('form');
  }, [haptic]);

  useEffect(() => {
    const button = webApp?.BackButton;
    if (!button || view !== 'results') return;
    button.show();
    button.onClick(backToForm);
    return () => { button.offClick(backToForm); button.hide(); };
  }, [webApp, view, backToForm]);
  useEffect(() => { window.scrollTo({ top: 0, behavior: 'instant' }); }, [view]);
  const banner = USE_MOCK ? <div className="fallback-banner" role="status"><strong>Демонстрационный режим</strong><p>Показаны учебные примеры, не реальный каталог.</p></div> : null;

  if (bootError) {
    return <ErrorState message={bootError} onRetry={() => window.location.reload()} />;
  }

  if (!meta) {
    return <LoadingState label="Загрузка справочников…" />;
  }

  if (view === 'results' && result) {
    return (
      <>
      {banner}
      <Results
        isDemo={USE_MOCK}
        result={result}
        deepeningContext={{
          ...(lastRequest?.query ? { query: lastRequest.query } : {}),
          ...(lastRequest?.historyProfileId ? { history_profile_id: lastRequest.historyProfileId } : {}),
          ...(lastRequest?.demoProfileId ? { demo_profile_id: lastRequest.demoProfileId } : {}),
        }}
        historyProfile={historyProfiles.find((item) => item.id === lastRequest?.historyProfileId)}
        loading={submitting}
        error={requestError}
        onRetry={retry}
        onChangeRequest={backToForm}
      />
      </>
    );
  }

  return (
    <>
    {banner}
    <ProfileForm
      meta={meta}
      demoProfiles={demoProfiles}
      historyProfiles={historyProfiles}
      initialHistoryProfileId={lastRequest?.historyProfileId}
      initialQuery={lastRequest?.query}
      initialProfile={lastRequest?.profile}
      initialDemoProfileId={lastRequest?.demoProfileId}
      loading={submitting}
      error={requestError}
      onSubmit={submit}
    />
    </>
  );
}
