import { useCallback, useEffect, useRef, useState } from 'react';
import ProfileForm from './screens/ProfileForm';
import Results from './screens/Results';
import { ErrorState, LoadingState } from './components/StatusStates';
import {
  ApiError,
  USE_MOCK,
  fetchDemoProfiles,
  fetchMeta,
  fetchRecommendations,
} from './api/client';
import type { DemoProfile, MetaResponse, Profile, RecommendationsResponse } from './types';
import { useTelegram } from './hooks/useTelegram';

type View = 'form' | 'results';

interface LastRequest {
  profile: Profile;
  demoProfileId?: string;
}

export default function App() {
  const { webApp, haptic, notify } = useTelegram();
  const [meta, setMeta] = useState<MetaResponse | null>(null);
  const [demoProfiles, setDemoProfiles] = useState<DemoProfile[]>([]);
  const [bootError, setBootError] = useState<string | null>(null);
  const inFlight = useRef(false);
  const [view, setView] = useState<View>('form');
  const [submitting, setSubmitting] = useState(false);
  const [result, setResult] = useState<RecommendationsResponse | null>(null);
  const [requestError, setRequestError] = useState<string | null>(null);
  const [lastRequest, setLastRequest] = useState<LastRequest | null>(null);
  useEffect(() => {
    let cancelled = false;
    Promise.all([fetchMeta(), fetchDemoProfiles().catch(() => [])])
      .then(([m, dp]) => {
        if (cancelled) return;
        setMeta(m);
        setDemoProfiles(dp);
      })
      .catch((e: unknown) => {
        if (cancelled) return;
        const message =
          e instanceof ApiError ? e.message : 'Не удалось загрузить справочники.';

        setBootError(message);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const submit = useCallback(
    async (profile: Profile, demoProfileId?: string) => {
      if (inFlight.current) return;
      inFlight.current = true;
      haptic('medium');
      setSubmitting(true);
      setRequestError(null);
      setLastRequest({ profile, demoProfileId });
      try {
        const res = await fetchRecommendations(profile, demoProfileId);
        setResult(res);
        setView('results');
        notify(res.items.length === 0 ? 'warning' : 'success');
      } catch (e: unknown) {
        const apiError = e instanceof ApiError ? e : null;
        setRequestError(apiError ? apiError.message : 'Не удалось получить подборку.');
        notify('error');
      } finally {
        inFlight.current = false;
        setSubmitting(false);
      }
    },
    [haptic, notify],
  );

  const retry = useCallback(() => {
    if (lastRequest) void submit(lastRequest.profile, lastRequest.demoProfileId);
  }, [lastRequest, submit]);

  const backToForm = useCallback(() => {
    haptic('light');
    setView('form');
  }, [haptic]);

  useEffect(() => {
    const button = webApp?.BackButton;
    if (!button || view !== 'results') return;
    button.show();
    button.onClick(backToForm);
    return () => {
      button.offClick(backToForm);
      button.hide();
    };
  }, [webApp, view, backToForm]);

  useEffect(() => {
    window.scrollTo({ top: 0, behavior: 'instant' });
  }, [view]);

  const banner = USE_MOCK ? (
    <div className="fallback-banner" role="status">
      <strong>Демонстрационный режим</strong>
      <p>Сейчас показаны примеры на учебных данных. Это не рекомендации из реального каталога.</p>
    </div>
  ) : null;

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
          result={result}
          loading={submitting}
          error={requestError}
          onRetry={retry}
          onChangeRequest={backToForm}
          isDemo={USE_MOCK}
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
        initialProfile={lastRequest?.profile}
        initialDemoProfileId={lastRequest?.demoProfileId}
        loading={submitting}
        error={requestError}
        onSubmit={submit}
      />
    </>
  );
}
