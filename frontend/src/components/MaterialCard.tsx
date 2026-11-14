import type { MaterialCard, MaterialKind } from '../types';

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
}

export default function MaterialCardView({ item, onRefineRequest }: Props) {
  const access = ACCESS_LABELS[item.access] ?? ACCESS_LABELS.unknown;
  const typeLabel = TYPE_LABELS[item.type] ?? 'Материал';
  const action = item.next_action;
  const canOpen = action.type !== 'refine_request' && item.source_url !== null;

  return (
    <article className="card">
      <div className="card-top">
        <span className="type-chip">{typeLabel}</span>
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
        {canOpen ? (
          <a
            className="btn btn-primary card-btn"
            href={item.source_url ?? undefined}
            target="_blank"
            rel="noopener noreferrer"
          >
            {action.label}
          </a>
        ) : (
          <button className="btn btn-secondary card-btn" onClick={onRefineRequest}>
            {action.type === 'refine_request' ? action.label : 'Изменить подборку'}
          </button>
        )}
      </div>
    </article>
  );
}
