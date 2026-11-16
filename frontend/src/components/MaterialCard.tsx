import type { MouseEvent } from 'react';
import { useTelegram } from '../hooks/useTelegram';
import Deepening from './Deepening';
import type { DeepeningContext, MaterialCard, MaterialKind } from '../types';

const TYPE_LABELS: Record<MaterialKind, string> = {
  article: 'Статья',
  course: 'Курс',
  webinar: 'Вебинар',
  document: 'Документ',
  calculator: 'Инструмент',
  podcast: 'Подкаст',
  term: 'Термин',
};

const ACCESS_LABELS: Record<MaterialCard['access'], { label: string; className: string }> = {
  free: { label: 'Бесплатно', className: 'access-free' },
  subscription: { label: 'По подписке', className: 'access-subscription' },
  paid: { label: 'Платно', className: 'access-paid' },
  unknown: { label: 'Условия на странице материала', className: 'access-unknown' },
};

interface Props {
  item: MaterialCard;
  onRefineRequest: () => void;
  position?: number;
  isDemo?: boolean;
  deepeningContext?: DeepeningContext;
}

export default function MaterialCardView({ item, onRefineRequest, position, isDemo = false, deepeningContext }: Props) {
  const { webApp } = useTelegram();
  const access = ACCESS_LABELS[item.access] ?? ACCESS_LABELS.unknown;
  const typeLabel = TYPE_LABELS[item.type] ?? 'Материал';
  const action = item.next_action;
  const canOpen = !isDemo && action.type !== 'refine_request' &&
    typeof item.source_url === 'string' && /^https?:\/\//i.test(item.source_url);

  // Внутри Telegram открываем через WebApp.openLink: обычный target="_blank"
  // в WebView может не сработать.
  const handleOpen = (event: MouseEvent<HTMLAnchorElement>) => {
    if (!webApp || !item.source_url) return;
    event.preventDefault();
    try {
      webApp.openLink(item.source_url);
    } catch {
      window.open(item.source_url, '_blank', 'noopener,noreferrer');
    }
  };

  return (
    <article className="card">
      <div className="card-top">
        <span className="type-chip">{position && <span className="card-number">{String(position).padStart(2, '0')}</span>}{typeLabel}</span>
        <span className={`access-chip ${access.className}`}>{access.label}</span>
      </div>

      <h3 className="card-title">{item.title}</h3>
      {item.safe_summary && <p className="card-summary">{item.safe_summary}</p>}

      {item.reasons.length > 0 && (
        <div className="card-reasons">
          <div className="reasons-title">Почему подходит</div>
          <ul>
            {item.reasons.map((reason) => (
              <li key={reason.code}>{reason.text}</li>
            ))}
          </ul>
        </div>
      )}

      <div className="card-action">
        {isDemo ? (
          <p className="demo-card-note">Пример материала. Переход появится при подключении к реальному каталогу.</p>
        ) : canOpen ? (
          <a
            className="btn btn-primary card-btn"
            href={item.source_url ?? undefined}
            target="_blank"
            rel="noopener noreferrer"
            onClick={handleOpen}
          >
            {action.label}
          </a>
        ) : (
          <button className="btn btn-secondary card-btn" onClick={onRefineRequest}>
            {action.type === 'refine_request' ? action.label : 'Изменить подборку'}
          </button>
        )}
      </div>
      {!isDemo && <Deepening materialKey={item.material_key} context={deepeningContext} />}
    </article>
  );
}
