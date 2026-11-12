interface LoadingProps {
  label?: string;
}

export function LoadingState({ label = 'Загрузка…' }: LoadingProps) {
  return (
    <div className="screen state">
      <div className="spinner" aria-hidden />
      <p className="hint">{label}</p>
    </div>
  );
}

interface ErrorProps {
  message: string;
  onRetry?: () => void;
  onChangeRequest?: () => void;
}

export function ErrorState({ message, onRetry, onChangeRequest }: ErrorProps) {
  return (
    <div className="screen state">
      <div className="state-icon">⚠️</div>
      <h2>Что-то пошло не так</h2>
      <p className="hint">{message}</p>
      <div className="actions">
        {onRetry && (
          <button className="btn btn-primary" onClick={onRetry}>
            Повторить запрос
          </button>
        )}
        {onChangeRequest && (
          <button className="btn btn-secondary" onClick={onChangeRequest}>
            Изменить запрос
          </button>
        )}
      </div>
    </div>
  );
}

interface EmptyProps {
  message: string;
  onChangeRequest: () => void;
}

export function EmptyState({ message, onChangeRequest }: EmptyProps) {
  return (
    <div className="screen state">
      <div className="state-icon">🔍</div>
      <h2>Подходящих материалов не нашлось</h2>
      <p className="hint">{message}</p>
      <div className="actions">
        <button className="btn btn-primary" onClick={onChangeRequest}>
          Изменить запрос
        </button>
      </div>
    </div>
  );
}
