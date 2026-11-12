import type { MaterialCard } from '../types';

const TYPE_LABELS: Record<string, string> = {
  article: 'Статья',
  lesson: 'Урок',
  course: 'Курс',
  webinar: 'Вебинар',
  document: 'Документ',
  calculator: 'Инструмент',
  podcast: 'Подкаст',
  glossary: 'Глоссарий',
};

const ACCESS_LABELS: Record<MaterialCard['access'], { label: string; className: string }> = {
  free: { label: 'Бесплатно', className: 'access-free' },
  paid: { label: 'Платно', className: 'access-paid' },
  unknown: { label: 'Условия на странице материала', className: 'access-unknown' },
};

interface Props {
  item: MaterialCard;
}

export default function MaterialCardView({ item }: Props) {
  const access = ACCESS_LABELS[item.access];
  const typeLabel = TYPE_LABELS[item.type] ?? item.type;

  return (
    <article className="card">
      <div className="card-top">
        <span className="type-chip">{typeLabel}</span>
        <span className={`access-chip ${access.className}`}>{access.label}</span>
      </div>

      <h3 className="card-title">{item.title}</h3>
      <p className="card-summary">{item.safe_summary}</p>

      {item.reasons.length > 0 && (
        <div className="card-reasons">
          <div className="reasons-title">Почему подходит</div>
          <ul>
            {item.reasons.map((r) => (
              <li key={r}>{r}</li>
            ))}
          </ul>
        </div>
      )}

      <div className="card-action">
        {item.source_url ? (
          <a
            className="btn btn-primary card-btn"
            href={item.source_url}
            target="_blank"
            rel="noopener noreferrer"
          >
            {item.next_action}
          </a>
        ) : (
          // URL не выдумываем: предлагаем изменить подборку.
          <span className="hint">
            {item.next_action} — ссылка недоступна, измените подборку или откройте другой материал.
          </span>
        )}
      </div>
    </article>
  );
}
