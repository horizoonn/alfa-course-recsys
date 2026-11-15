import type { RecommendationsResponse } from '../types';
import MaterialCardView from '../components/MaterialCard';
import { EmptyState, ErrorState, LoadingState } from '../components/StatusStates';

interface Props {
  result: RecommendationsResponse;
  loading: boolean;
  error: string | null;
  onRetry: () => void;
  onChangeRequest: () => void;
  isDemo?: boolean;
}

const MODE_LABELS: Record<string, string> = {
  personalized: 'Персональная подборка',
  topic: 'Тематическая подборка',
  starter: 'Стартовая подборка',
  fallback: 'Общая подборка',
};

export default function Results({ result, loading, error, onRetry, onChangeRequest, isDemo = false }: Props) {
  if (loading) {
    return <LoadingState label="Подбираем материалы…" />;
  }

  if (error) {
    return <ErrorState message={error} onRetry={onRetry} onChangeRequest={onChangeRequest} />;
  }

  if (result.items.length === 0) {
    const message = result.limitations.map((limitation) => limitation.message).join(' ');
    return (
      <EmptyState
        message={message || 'Подходящих материалов не нашлось. Измените запрос.'}
        onChangeRequest={onChangeRequest}
      />
    );
  }

  const modeLabel = MODE_LABELS[result.mode] ?? result.mode;

  return (
    <div className="screen">
      <header className="header">
        <button className="back-link" onClick={onChangeRequest}>
          ← Изменить запрос
        </button>
        <p className="eyebrow">Альфа Курс · {isDemo ? 'пример подборки' : 'ваш следующий шаг'}</p>
        <h1>Есть с чего начать</h1>
        <p className="subtitle">До трёх материалов по вашему запросу. Выберите то, что удобно изучить сейчас.</p>
        <div className="mode-badge" data-mode={result.mode}>
          {modeLabel}
        </div>
        {result.limitations.length > 0 && (
          <ul className="hint">
            {result.limitations.map((limitation) => (
              <li key={limitation.code}>{limitation.message}</li>
            ))}
          </ul>
        )}
      </header>

      <div className="cards">
        {result.items.map((item, index) => (
          <MaterialCardView
            key={item.material_key}
            item={item}
            onRefineRequest={onChangeRequest}
            position={index + 1}
            isDemo={isDemo}
          />
        ))}
      </div>
    </div>
  );
}
