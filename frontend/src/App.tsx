import { useCallback, useEffect, useState } from 'react';
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
import { mockDemoProfiles, mockMeta, mockRecommendations } from './api/mock';
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
  // Сообщение о переходе на демонстрационные данные (бэкенд недоступен).
  const [fallbackNotice, setFallbackNotice] = useState<string | null>(null);
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

        if (USE_MOCK) {
          setBootError(message);
          return;
        }

        // Бэкенд не поднят/недоступен: не показываем тупик, а работаем на демо-данных.
        setMeta(mockMeta);
        setDemoProfiles(mockDemoProfiles);
        setFallbackNotice(
          `${message} Показаны демонстрационные данные — интерфейс и сценарий полностью работоспособны.`,
        );
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
        const apiError = e instanceof ApiError ? e : null;
        const backendDown =
          apiError !== null &&
          (apiError.code === 'NETWORK_ERROR' || apiError.code === 'DATA_UNAVAILABLE');

        if (!USE_MOCK && backendDown) {
          // Бэкенд недоступен — выдаём демо-подборку, но честно помечаем это.
          const fallback = await mockRecommendations(profile, demoProfileId);
          setResult(fallback);
          setView('results');
          setFallbackNotice(
            'Бэкенд недоступен — подборка сформирована на демонстрационных данных.',
          );
          notify('warning');
        } else {
          setRequestError(apiError ? apiError.message : 'Не удалось получить подборку.');
          notify('error');
        }
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

  const banner = fallbackNotice ? (
    <div className="fallback-banner" role="status">
      {fallbackNotice}
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
