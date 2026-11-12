import { useCallback, useEffect, useState } from 'react';
import ProfileForm from './screens/ProfileForm';
import Results from './screens/Results';
import { ErrorState, LoadingState } from './components/StatusStates';
import { ApiError, fetchDemoProfiles, fetchMeta, fetchRecommendations } from './api/client';
import type { DemoProfile, MetaResponse, Profile, RecommendationsResponse } from './types';
import { useTelegram } from './hooks/useTelegram';

type View = 'form' | 'results';

interface LastRequest {
  profile: Profile;
  demoProfileId?: string;
}

export default function App() {
  const { haptic, notify } = useTelegram();
  const [meta, setMeta] = useState<MetaResponse | null>(null);
  const [demoProfiles, setDemoProfiles] = useState<DemoProfile[]>([]);
  const [bootError, setBootError] = useState<string | null>(null);
  const [view, setView] = useState<View>('form');
  const [submitting, setSubmitting] = useState(false);
  const [result, setResult] = useState<RecommendationsResponse | null>(null);
  const [requestError, setRequestError] = useState<string | null>(null);
  const [lastRequest, setLastRequest] = useState<LastRequest | null>(null);

  // Справочники и демо-профили (P01–P10). Демо-профили необязательны —
  // их отсутствие не ломает основной сценарий.
  useEffect(() => {
    let cancelled = false;
    Promise.all([fetchMeta(), fetchDemoProfiles().catch(() => [])])
      .then(([m, dp]) => {
        if (cancelled) return;
        setMeta(m);
        setDemoProfiles(dp);
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
    async (profile: Profile, demoProfileId?: string) => {
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
        setRequestError(
          e instanceof ApiError ? e.message : 'Не удалось получить подборку.',
        );
        notify('error');
      } finally {
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

  if (bootError) {
    return <ErrorState message={bootError} onRetry={() => window.location.reload()} />;
  }

  if (!meta) {
    return <LoadingState label="Загрузка справочников…" />;
  }

  if (view === 'results' && result) {
    return (
      <Results
        result={result}
        loading={submitting}
        error={requestError}
        onRetry={retry}
        onChangeRequest={backToForm}
      />
    );
  }

  return (
    <ProfileForm
      meta={meta}
      demoProfiles={demoProfiles}
      loading={submitting}
      error={requestError}
      onSubmit={submit}
    />
  );
}
